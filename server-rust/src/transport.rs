use crate::auth::{Auth, now, random_token};
use anyhow::Result;
use axum::{
    Json, Router,
    body::Bytes,
    extract::{
        ConnectInfo, DefaultBodyLimit, State,
        ws::{Message, WebSocket, WebSocketUpgrade},
    },
    http::{HeaderMap, Method, StatusCode, Uri},
    response::{Html, IntoResponse, Response},
    routing::{get, post},
};
use futures_util::SinkExt;
use serde_json::{Value, json};
use std::{
    collections::HashMap,
    net::{IpAddr, SocketAddr},
    path::PathBuf,
    sync::{Arc, Mutex},
    time::Duration,
};
use tokio::sync::{Semaphore, broadcast};

type CommandCache = HashMap<String, (u64, String, StatusCode, Value)>;
#[derive(Clone)]
pub struct App {
    auth: Arc<Mutex<Auth>>,
    port: u16,
    admin_token: String,
    endpoints: Arc<Mutex<Value>>,
    events: broadcast::Sender<Value>,
    command_gate: Arc<Semaphore>,
    read_gate: Arc<Semaphore>,
    auth_gate: Arc<Semaphore>,
    cache: Arc<Mutex<CommandCache>>,
    sockets: Arc<Semaphore>,
    shutdown: axum_server::Handle<SocketAddr>,
    refresh_lock: Arc<tokio::sync::Mutex<()>>,
    refresh_notify: Arc<tokio::sync::Notify>,
}
fn error(status: StatusCode, message: impl ToString) -> Response {
    (status, Json(json!({"error":message.to_string()}))).into_response()
}
fn token(h: &HeaderMap) -> &str {
    h.get("authorization")
        .and_then(|v| v.to_str().ok())
        .and_then(|v| v.strip_prefix("Bearer "))
        .unwrap_or("")
}
async fn authorized(a: &App, h: &HeaderMap) -> Option<String> {
    let auth = a.auth.clone();
    let token = token(h).to_string();
    tokio::task::spawn_blocking(move || auth.lock().unwrap().authenticate(&token).map(|d| d.id))
        .await
        .ok()
        .flatten()
}
async fn valid_session(a: &App, token: &str) -> bool {
    let auth = a.auth.clone();
    let token = token.to_string();
    tokio::task::spawn_blocking(move || auth.lock().unwrap().authenticate(&token).is_some())
        .await
        .unwrap_or(false)
}
fn origin_allowed(h: &HeaderMap) -> bool {
    !h.contains_key("origin")
} // native clients only on remote API

pub async fn run() -> Result<()> {
    let port = std::env::var("PCREMOTE_PORT")
        .unwrap_or("8443".into())
        .parse::<u16>()?;
    let dir = std::env::var_os("PCREMOTE_DATA_DIR")
        .map(PathBuf::from)
        .unwrap_or_else(|| {
            PathBuf::from(std::env::var_os("LOCALAPPDATA").unwrap_or_default())
                .join("PCRemoteRust/data")
        });
    // Claim the per-user setup port before touching persistent identity files.
    let admin_port = port
        .checked_add(1)
        .ok_or_else(|| anyhow::anyhow!("port must be less than 65535"))?;
    let listener =
        match tokio::net::TcpListener::bind((std::net::Ipv4Addr::LOCALHOST, admin_port)).await {
            Ok(listener) => listener,
            Err(e) if e.kind() == std::io::ErrorKind::AddrInUse => {
                if !std::env::args().any(|a| a == "--background") {
                    crate::gui::focus_existing();
                }
                return Ok(());
            }
            Err(e) => return Err(e.into()),
        };
    let auth = Auth::load(&dir)?;
    let tls = axum_server::tls_rustls::RustlsConfig::from_pem(
        auth.store.cert_pem.as_bytes().to_vec(),
        auth.store.key_pem.as_bytes().to_vec(),
    )
    .await?;
    let (events, _) = broadcast::channel(64);
    let app = App {
        auth: Arc::new(Mutex::new(auth)),
        port,
        admin_token: random_token(),
        endpoints: Arc::new(Mutex::new(json!({}))),
        events,
        command_gate: Arc::new(Semaphore::new(1)),
        read_gate: Arc::new(Semaphore::new(4)),
        auth_gate: Arc::new(Semaphore::new(2)),
        cache: Arc::new(Mutex::new(HashMap::new())),
        sockets: Arc::new(Semaphore::new(16)),
        shutdown: axum_server::Handle::new(),
        refresh_lock: Arc::new(tokio::sync::Mutex::new(())),
        refresh_notify: Arc::new(tokio::sync::Notify::new()),
    };
    refresh_endpoints(&app).await;
    let show_gui = !std::env::args().any(|a| a == "--background");
    crate::tray::start(admin_port, {
        let shutdown = app.shutdown.clone();
        move || {
            shutdown.graceful_shutdown(Some(Duration::from_secs(2)));
            crate::gui::quit();
        }
    });
    crate::gui::start(
        crate::gui::GuiContext {
            auth: app.auth.clone(),
            endpoints: app.endpoints.clone(),
            port: app.port,
            shutdown: app.shutdown.clone(),
            refresh_notify: app.refresh_notify.clone(),
        },
        show_gui,
    );
    let notify = app.events.clone();
    crate::platform::start_event_listener(move || {
        let _ = notify.send(json!({"type":"state_changed"}));
    });
    let api = Router::new()
        .route("/health", get(health))
        .route("/auth/pair", post(pair))
        .route("/auth/login", post(login))
        .route("/events", get(event_upgrade))
        .fallback(api)
        .layer(DefaultBodyLimit::max(16384))
        .with_state(app.clone());
    let local = Router::new()
        .route("/", get(dashboard))
        .route("/setup", post(setup))
        .layer(DefaultBodyLimit::max(16384))
        .with_state(app.clone());
    let local_task = tokio::spawn(async move { axum::serve(listener, local).await });
    let updater = app.clone();
    let update_task = tokio::spawn(async move {
        let started = std::time::Instant::now();
        loop {
            let tailscale_empty = updater
                .endpoints
                .lock()
                .unwrap()
                .get("tailscale_url")
                .and_then(Value::as_str)
                .unwrap_or("")
                .is_empty();
            let secs = if started.elapsed() < Duration::from_secs(90) || tailscale_empty {
                3
            } else {
                30
            };
            tokio::select! {
                _ = tokio::time::sleep(Duration::from_secs(secs)) => {}
                _ = updater.refresh_notify.notified() => {}
            }
            refresh_endpoints(&updater).await;
        }
    });
    println!("PC Remote Rust: HTTPS :{port}; local setup http://127.0.0.1:{admin_port}/");
    let handle = app.shutdown.clone();
    let shutdown = handle.clone();
    tokio::spawn(async move {
        let _ = tokio::signal::ctrl_c().await;
        shutdown.graceful_shutdown(Some(Duration::from_secs(3)));
        crate::gui::quit();
    });
    let result = axum_server::bind_rustls(SocketAddr::from(([0, 0, 0, 0], port)), tls)
        .handle(handle)
        .serve(api.into_make_service_with_connect_info::<SocketAddr>())
        .await;
    crate::gui::quit();
    local_task.abort();
    update_task.abort();
    result?;
    Ok(())
}

fn build_endpoints_snapshot(
    status: &Value,
    hosts: Vec<String>,
    tailscale_installed: bool,
    port: u16,
    server_id: String,
    fingerprint: String,
) -> Value {
    let tail_ip = status["Self"]["TailscaleIPs"]
        .as_array()
        .and_then(|ips| {
            ips.iter()
                .filter_map(Value::as_str)
                .find(|s| s.parse::<std::net::Ipv4Addr>().is_ok())
        })
        .unwrap_or("");
    let tailscale_dns = status["Self"]["DNSName"]
        .as_str()
        .unwrap_or("")
        .trim_end_matches('.')
        .to_string();
    let url = if tail_ip.is_empty() {
        String::new()
    } else {
        format!("https://{tail_ip}:{port}")
    };
    let mut tailscale_urls = Vec::new();
    if !tail_ip.is_empty() {
        tailscale_urls.push(format!("https://{tail_ip}:{port}"));
    }
    if !tailscale_dns.is_empty() {
        tailscale_urls.push(format!("https://{tailscale_dns}:{port}"));
    }
    let tailscale_state = status["BackendState"]
        .as_str()
        .unwrap_or(if tailscale_installed {
            "Unavailable"
        } else {
            "NotInstalled"
        });
    json!({
        "hosts": hosts,
        "tailscale_url": url,
        "tailscale_urls": tailscale_urls,
        "tailscale_dns": tailscale_dns,
        "tailscale_state": tailscale_state,
        "tailscale_installed": tailscale_installed,
        "server_id": server_id,
        "version": "5.0.0",
        "fingerprint": fingerprint,
    })
}

async fn refresh_endpoints(a: &App) {
    let _guard = a.refresh_lock.lock().await;
    let hosts: Vec<String> = if_addrs::get_if_addrs()
        .unwrap_or_default()
        .into_iter()
        .filter_map(|i| match i.ip() {
            IpAddr::V4(ip) if !ip.is_loopback() && !ip.is_link_local() => Some(ip.to_string()),
            _ => None,
        })
        .collect();
    let pf = PathBuf::from(
        std::env::var_os("ProgramFiles").unwrap_or_else(|| "C:\\Program Files".into()),
    );
    let cli_path = pf.join("Tailscale/tailscale.exe");
    let ipn_path = pf.join("Tailscale/tailscale-ipn.exe");
    let tailscale_installed = cli_path.exists() || ipn_path.exists();
    let mut cmd = tokio::process::Command::new(if cli_path.exists() {
        cli_path
    } else {
        PathBuf::from("tailscale.exe")
    });
    cmd.args(["status", "--json"])
        .kill_on_drop(true)
        .creation_flags(0x08000000);
    // Read-only status: never login, reset, change DNS, routes or exit nodes.
    let status = match tokio::time::timeout(Duration::from_secs(3), cmd.output()).await {
        Ok(Ok(o)) if o.status.success() => {
            serde_json::from_slice::<Value>(&o.stdout).unwrap_or(Value::Null)
        }
        _ => Value::Null,
    };
    let (server_id, fingerprint) = {
        let auth = a.auth.lock().unwrap();
        (auth.store.server_id.clone(), auth.store.fingerprint.clone())
    };
    let new = build_endpoints_snapshot(
        &status,
        hosts,
        tailscale_installed,
        a.port,
        server_id,
        fingerprint,
    );
    let mut current = a.endpoints.lock().unwrap();
    if *current != new {
        *current = new.clone();
        let _ = a.events.send(json!({"type":"endpoints","data":new}));
    }
}
async fn health(State(a): State<App>) -> Response {
    blocking_response(move||Json(json!({"status":"ok","version":"5.0.0","platform":"windows","protocol_version":1,"server_id":a.auth.lock().unwrap().store.server_id})).into_response()).await
}
async fn pair(
    State(a): State<App>,
    ConnectInfo(peer): ConnectInfo<SocketAddr>,
    headers: HeaderMap,
    Json(body): Json<Value>,
) -> Response {
    let Ok(permit) = a.auth_gate.clone().try_acquire_owned() else {
        return error(StatusCode::TOO_MANY_REQUESTS, "authentication busy");
    };
    blocking_response(move || {
        let _permit = permit;
        pair_sync(a, peer, headers, body)
    })
    .await
}
async fn blocking_response(f: impl FnOnce() -> Response + Send + 'static) -> Response {
    tokio::task::spawn_blocking(f).await.unwrap_or_else(|_| {
        error(
            StatusCode::INTERNAL_SERVER_ERROR,
            "authentication worker failed",
        )
    })
}
fn pair_sync(a: App, peer: SocketAddr, headers: HeaderMap, body: Value) -> Response {
    if !origin_allowed(&headers) {
        return error(StatusCode::FORBIDDEN, "native pairing only");
    }
    let mut auth = a.auth.lock().unwrap();
    let ip = peer.ip().to_string();
    if auth.locked(&ip) {
        return error(StatusCode::TOO_MANY_REQUESTS, "try again later");
    }
    match auth.pair(body["pair_token"].as_str().unwrap_or(""),headers.get("x-device-name").and_then(|v|v.to_str().ok()).unwrap_or("Mobile"),&ip){Ok((token,d))=>Json(json!({"token":token,"device_id":d.id,"expires_at":d.expires_at,"server_id":auth.store.server_id})).into_response(),Err(e)=>error(StatusCode::UNAUTHORIZED,e)}
}
async fn login(
    State(a): State<App>,
    ConnectInfo(peer): ConnectInfo<SocketAddr>,
    headers: HeaderMap,
    Json(body): Json<Value>,
) -> Response {
    let Ok(permit) = a.auth_gate.clone().try_acquire_owned() else {
        return error(StatusCode::TOO_MANY_REQUESTS, "authentication busy");
    };
    blocking_response(move || {
        let _permit = permit;
        login_sync(a, peer, headers, body)
    })
    .await
}
fn login_sync(a: App, peer: SocketAddr, headers: HeaderMap, body: Value) -> Response {
    if !origin_allowed(&headers) {
        return error(StatusCode::FORBIDDEN, "native login only");
    }
    let mut auth = a.auth.lock().unwrap();
    let ip = peer.ip().to_string();
    if auth.locked(&ip) {
        return error(StatusCode::TOO_MANY_REQUESTS, "try again later");
    }
    match auth.login(body["pin"].as_str().unwrap_or(""), "Mobile", &ip) {
        Ok((token, d)) => {
            Json(json!({"token":token,"device_id":d.id,"expires_at":d.expires_at})).into_response()
        }
        Err(e) => error(StatusCode::UNAUTHORIZED, e),
    }
}
fn read_path(path: &str) -> bool {
    matches!(
        path,
        "/audio/status"
            | "/audio/devices"
            | "/audio/channels"
            | "/media/status"
            | "/system/brightness"
    )
}
fn write_path(path: &str) -> bool {
    matches!(
        path,
        "/audio/volume"
            | "/audio/mute"
            | "/audio/device/volume"
            | "/audio/device/mute"
            | "/audio/channel/volume"
            | "/audio/channel/mute"
            | "/media/play"
            | "/media/next"
            | "/media/prev"
            | "/media/seek-forward"
            | "/media/seek-backward"
            | "/system/lock"
            | "/system/sleep"
            | "/system/restart"
            | "/system/shutdown"
            | "/system/shutdown/cancel"
            | "/system/display/off"
            | "/system/brightness"
            | "/browser/open"
    )
}
async fn api(
    State(a): State<App>,
    method: Method,
    uri: Uri,
    headers: HeaderMap,
    bytes: Bytes,
) -> Response {
    let received = std::time::Instant::now();
    if !origin_allowed(&headers) {
        return error(StatusCode::FORBIDDEN, "browser origins are not accepted");
    }
    let Some(device) = authorized(&a, &headers).await else {
        return error(StatusCode::UNAUTHORIZED, "pair this device first");
    };
    let path = uri.path();
    if method == Method::GET && !bytes.is_empty() {
        return error(
            StatusCode::BAD_REQUEST,
            "GET requests must not contain a body",
        );
    }
    if path == "/auth/verify" && method == Method::GET {
        return Json(json!({"success":true})).into_response();
    }
    if path == "/connection" && method == Method::GET {
        if uri
            .query()
            .is_some_and(|q| q.split('&').any(|p| p == "refresh=1"))
        {
            refresh_endpoints(&a).await;
        }
        return Json(a.endpoints.lock().unwrap().clone()).into_response();
    }
    let body = if bytes.is_empty() {
        json!({})
    } else {
        match serde_json::from_slice::<Value>(&bytes) {
            Ok(v) => v,
            Err(_) => return error(StatusCode::BAD_REQUEST, "invalid JSON"),
        }
    };
    if method == Method::POST && path == "/system/brightness" && body.get("level").is_none() {
        return error(StatusCode::BAD_REQUEST, "level is required");
    }
    if path == "/auth/logout" && method == Method::POST {
        return blocking_response(move || match a.auth.lock().unwrap().revoke(&device) {
            Ok(()) => Json(json!({"success":true})).into_response(),
            Err(e) => error(StatusCode::INTERNAL_SERVER_ERROR, e),
        })
        .await;
    }
    if path == "/system/pin" && method == Method::POST {
        return blocking_response(move || {
            let mut auth = a.auth.lock().unwrap();
            if !auth.store.pin_hash.as_ref().is_some_and(|h| {
                bcrypt::verify(body["current_pin"].as_str().unwrap_or(""), h).unwrap_or(false)
            }) {
                return error(StatusCode::FORBIDDEN, "current PIN incorrect");
            }
            match auth.set_pin(body["new_pin"].as_str().unwrap_or("")) {
                Ok(()) => Json(json!({"success":true})).into_response(),
                Err(e) => error(StatusCode::BAD_REQUEST, e),
            }
        })
        .await;
    }
    if !((method == Method::GET && read_path(path)) || (method == Method::POST && write_path(path)))
    {
        return error(StatusCode::NOT_FOUND, "unsupported endpoint or method");
    }
    let id = headers
        .get("x-request-id")
        .and_then(|h| h.to_str().ok())
        .unwrap_or("");
    if id.len() > 128 {
        return error(StatusCode::BAD_REQUEST, "request ID too long");
    }
    let has_id = !id.is_empty();
    let deadline = match headers.get("x-command-deadline-ms") {
        Some(raw) => match raw.to_str().ok().and_then(|v| v.parse::<u128>().ok()) {
            Some(d) => Some(d),
            None => return error(StatusCode::BAD_REQUEST, "invalid command deadline"),
        },
        None => None,
    };
    let now_ms = std::time::SystemTime::now()
        .duration_since(std::time::UNIX_EPOCH)
        .unwrap_or_default()
        .as_millis();
    if deadline.is_some_and(|d| d < now_ms || d > now_ms + 30_000) {
        return error(
            StatusCode::BAD_REQUEST,
            "command deadline expired or device clock differs",
        );
    }
    let gate = if method == Method::GET {
        a.read_gate.clone()
    } else {
        a.command_gate.clone()
    };
    let permit = match gate.try_acquire_owned() {
        Ok(p) => p,
        _ => {
            return error(
                StatusCode::TOO_MANY_REQUESTS,
                "another command is running; do not queue commands",
            );
        }
    };
    if authorized(&a, &headers).await.is_none() {
        return error(StatusCode::UNAUTHORIZED, "session revoked");
    }
    let key = format!("{device}:{id}");
    let signature = format!("{method}:{path}:{body}");
    if method == Method::POST && !id.is_empty() {
        let mut cache = a.cache.lock().unwrap();
        cache.retain(|_, (t, _, _, _)| *t > now());
        if let Some((_, sig, status, result)) = cache.get(&key) {
            return if sig == &signature {
                (*status, Json(result.clone())).into_response()
            } else {
                error(
                    StatusCode::CONFLICT,
                    "request ID reused with different command",
                )
            };
        }
        if cache.len() >= 1024 {
            return error(
                StatusCode::TOO_MANY_REQUESTS,
                "command history full; retry later",
            );
        }
    }
    let path = path.to_string();
    // Detached worker retains permit and records outcome even if HTTP client disconnects.
    tokio::spawn(async move {
        let _permit = permit;
        let result = tokio::task::spawn_blocking(move || {
            let now_ms = std::time::SystemTime::now()
                .duration_since(std::time::UNIX_EPOCH)
                .unwrap_or_default()
                .as_millis();
            if received.elapsed() > Duration::from_secs(3) || deadline.is_some_and(|d| d < now_ms) {
                anyhow::bail!("command expired before execution")
            }
            crate::platform::execute(&path, &body)
        })
        .await;
        let (status, value) = match result {
            Ok(Ok(value)) => (StatusCode::OK, value),
            Ok(Err(e)) => (StatusCode::BAD_REQUEST, json!({"error":e.to_string()})),
            Err(_) => (
                StatusCode::INTERNAL_SERVER_ERROR,
                json!({"error":"platform worker failed; command outcome unknown"}),
            ),
        };
        if method == Method::POST {
            if has_id {
                a.cache
                    .lock()
                    .unwrap()
                    .insert(key, (now() + 300, signature, status, value.clone()));
            }
            let _ = a.events.send(json!({"type":"state_changed"}));
        }
        (status, Json(value)).into_response()
    })
    .await
    .unwrap_or_else(|_| {
        error(
            StatusCode::INTERNAL_SERVER_ERROR,
            "command outcome unknown; do not replay",
        )
    })
}

async fn event_upgrade(State(a): State<App>, headers: HeaderMap, ws: WebSocketUpgrade) -> Response {
    if !origin_allowed(&headers) || authorized(&a, &headers).await.is_none() {
        return error(StatusCode::UNAUTHORIZED, "pair first");
    }
    let Ok(permit) = a.sockets.clone().try_acquire_owned() else {
        return error(StatusCode::TOO_MANY_REQUESTS, "too many event connections");
    };
    let bearer = token(&headers).to_string();
    ws.max_message_size(16384)
        .on_upgrade(move |socket| async move {
            let _permit = permit;
            event_loop(a, bearer, socket).await
        })
}
async fn event_loop(a: App, token: String, mut socket: WebSocket) {
    if !valid_session(&a, &token).await {
        return;
    }
    let mut rx = a.events.subscribe();
    let mut interval = tokio::time::interval(Duration::from_secs(20));
    let mut last_received = std::time::Instant::now();
    let first =
        json!({"type":"snapshot","protocol_version":1,"data":a.endpoints.lock().unwrap().clone()});
    if socket
        .send(Message::Text(first.to_string().into()))
        .await
        .is_err()
    {
        return;
    }
    loop {
        tokio::select! {
                    _=interval.tick()=>{if last_received.elapsed()>Duration::from_secs(65)||!valid_session(&a,&token).await{let _=socket.close().await;break}
        if socket.send(Message::Ping(vec![].into())).await.is_err(){break}},
                    event=rx.recv()=>{if !valid_session(&a,&token).await{break}let value=event.unwrap_or(json!({"type":"resync_required"}));if socket.send(Message::Text(value.to_string().into())).await.is_err(){break}},
            msg = socket.recv() => {
                last_received = std::time::Instant::now();
                match msg {
                    Some(Ok(Message::Text(text))) => {
                        if !valid_session(&a, &token).await {
                            break;
                        }
                        if text == "snapshot" {
                            let v = json!({"type":"snapshot","data":a.endpoints.lock().unwrap().clone()});
                            if socket.send(Message::Text(v.to_string().into())).await.is_err() {
                                break;
                            }
                        }
                    }
                    Some(Ok(Message::Close(_))) | None | Some(Err(_)) => break,
                    _ => {}
                }
            }
                }
    }
}

async fn dashboard(State(a): State<App>, headers: HeaderMap) -> Response {
    // Host check blocks DNS rebinding against the loopback-only setup server.
    if headers.get("host").and_then(|h| h.to_str().ok())
        != Some(&format!("127.0.0.1:{}", a.port + 1))
    {
        return error(StatusCode::FORBIDDEN, "open setup through 127.0.0.1");
    }
    let html = include_str!("dashboard.html").replace("__TOKEN__", &a.admin_token);
    let mut response = Html(html).into_response();
    let h = response.headers_mut();
    h.insert("cache-control", "no-store".parse().unwrap());
    h.insert("x-frame-options", "DENY".parse().unwrap());
    h.insert("referrer-policy", "no-referrer".parse().unwrap());
    response
}
async fn setup(State(a): State<App>, headers: HeaderMap, Json(body): Json<Value>) -> Response {
    blocking_response(move || setup_sync(a, headers, body)).await
}
fn setup_sync(a: App, headers: HeaderMap, body: Value) -> Response {
    if headers.get("x-setup-token").and_then(|h| h.to_str().ok()) != Some(&a.admin_token) {
        return error(StatusCode::FORBIDDEN, "invalid setup token");
    }
    let mut auth = a.auth.lock().unwrap();
    match body["action"].as_str().unwrap_or("status") {
        "stop"=>{a.shutdown.graceful_shutdown(Some(Duration::from_secs(2)));Json(json!({"success":true})).into_response()},
        "pair"=>{let p=auth.new_pair();let ep=a.endpoints.lock().unwrap().clone();let hosts=ep["hosts"].as_array().cloned().unwrap_or_default();let host=hosts.first().and_then(Value::as_str).unwrap_or("127.0.0.1");let payload=json!({"host":host,"port":a.port,"protocol":"https","pair_token":p,"fingerprint":auth.store.fingerprint,"server_name":"PC Remote Rust","server_id":auth.store.server_id,"protocol_version":1,"alternate_hosts":hosts,"tailscale_url":ep["tailscale_url"]});
            match qrcode::QrCode::new(payload.to_string()){Ok(qr)=>Json(json!({"svg":qr.render::<qrcode::render::svg::Color>().min_dimensions(320,320).build(),"expires_in":120})).into_response(),Err(e)=>error(StatusCode::INTERNAL_SERVER_ERROR,e)}
        },
        "revoke"=>match auth.revoke(body["device_id"].as_str().unwrap_or("")){Ok(())=>Json(json!({"success":true})).into_response(),Err(e)=>error(StatusCode::INTERNAL_SERVER_ERROR,e)},
        "pin"=>match auth.set_pin(body["pin"].as_str().unwrap_or("")){Ok(())=>Json(json!({"success":true})).into_response(),Err(e)=>error(StatusCode::BAD_REQUEST,e)},
        _=>Json(json!({"endpoints":a.endpoints.lock().unwrap().clone(),"server_id":auth.store.server_id,"devices":auth.store.devices.iter().map(|d|json!({"id":d.id,"name":d.name,"expires_at":d.expires_at})).collect::<Vec<_>>()})).into_response()
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    fn fixture() -> (tempfile::TempDir, App, HeaderMap, String) {
        let dir = tempfile::tempdir().unwrap();
        let mut auth = Auth::load(dir.path()).unwrap();
        let pair = auth.new_pair();
        let (token, device) = auth.pair(&pair, "test", "127.0.0.1").unwrap();
        let mut headers = HeaderMap::new();
        headers.insert("authorization", format!("Bearer {token}").parse().unwrap());
        let (events, _) = broadcast::channel(64);
        let app = App {
            auth: Arc::new(Mutex::new(auth)),
            port: 8443,
            admin_token: random_token(),
            endpoints: Arc::new(Mutex::new(json!({}))),
            events,
            command_gate: Arc::new(Semaphore::new(1)),
            read_gate: Arc::new(Semaphore::new(4)),
            auth_gate: Arc::new(Semaphore::new(2)),
            cache: Arc::new(Mutex::new(HashMap::new())),
            sockets: Arc::new(Semaphore::new(16)),
            shutdown: axum_server::Handle::new(),
            refresh_lock: Arc::new(tokio::sync::Mutex::new(())),
            refresh_notify: Arc::new(tokio::sync::Notify::new()),
        };
        (dir, app, headers, device.id)
    }
    #[tokio::test]
    async fn rejects_unpaired_and_browser_requests() {
        let (_dir, a, mut headers, _) = fixture();
        let path = "/system/lock".parse::<Uri>().unwrap();
        assert_eq!(
            api(
                State(a.clone()),
                Method::POST,
                path.clone(),
                HeaderMap::new(),
                Bytes::new()
            )
            .await
            .status(),
            StatusCode::UNAUTHORIZED
        );
        headers.insert("origin", "https://evil.example".parse().unwrap());
        assert_eq!(
            api(State(a), Method::POST, path, headers, Bytes::new())
                .await
                .status(),
            StatusCode::FORBIDDEN
        );
    }
    #[tokio::test]
    async fn verifies_and_revokes_individual_device() {
        let (_dir, a, h, id) = fixture();
        assert_eq!(
            api(
                State(a.clone()),
                Method::GET,
                "/auth/verify".parse().unwrap(),
                h.clone(),
                Bytes::new()
            )
            .await
            .status(),
            StatusCode::OK
        );
        a.auth.lock().unwrap().revoke(&id).unwrap();
        assert_eq!(
            api(
                State(a),
                Method::GET,
                "/auth/verify".parse().unwrap(),
                h,
                Bytes::new()
            )
            .await
            .status(),
            StatusCode::UNAUTHORIZED
        );
    }
    #[tokio::test]
    async fn dedup_returns_saved_result_without_actuating() {
        let (_dir, a, mut h, id) = fixture();
        h.insert("x-request-id", "req1".parse().unwrap());
        a.cache.lock().unwrap().insert(
            format!("{id}:req1"),
            (
                now() + 300,
                "POST:/system/lock:{}".into(),
                StatusCode::OK,
                json!({"success":true}),
            ),
        );
        let r = api(
            State(a.clone()),
            Method::POST,
            "/system/lock".parse().unwrap(),
            h.clone(),
            Bytes::new(),
        )
        .await;
        assert_eq!(r.status(), StatusCode::OK);
        assert_eq!(
            api(
                State(a),
                Method::POST,
                "/system/restart".parse().unwrap(),
                h,
                Bytes::new()
            )
            .await
            .status(),
            StatusCode::CONFLICT
        );
    }
    #[tokio::test]
    async fn rejects_missing_brightness_and_bad_json() {
        let (_dir, a, h, _) = fixture();
        assert_eq!(
            api(
                State(a.clone()),
                Method::POST,
                "/system/brightness".parse().unwrap(),
                h.clone(),
                Bytes::new()
            )
            .await
            .status(),
            StatusCode::BAD_REQUEST
        );
        assert_eq!(
            api(
                State(a),
                Method::POST,
                "/system/lock".parse().unwrap(),
                h,
                Bytes::from_static(b"{")
            )
            .await
            .status(),
            StatusCode::BAD_REQUEST
        );
    }
    #[tokio::test]
    async fn no_command_queue_when_busy() {
        let (_dir, a, h, _) = fixture();
        let _guard = a.command_gate.clone().acquire_owned().await.unwrap();
        assert_eq!(
            api(
                State(a),
                Method::POST,
                "/system/lock".parse().unwrap(),
                h,
                Bytes::new()
            )
            .await
            .status(),
            StatusCode::TOO_MANY_REQUESTS
        );
    }
    #[tokio::test]
    async fn local_setup_requires_capability() {
        let (_dir, a, _, _) = fixture();
        assert_eq!(
            setup(State(a), HeaderMap::new(), Json(json!({"action":"pair"})))
                .await
                .status(),
            StatusCode::FORBIDDEN
        );
    }
    #[test]
    fn build_endpoints_snapshot_extracts_tailscale_ip_and_trimmed_dns() {
        let status = json!({
            "BackendState": "Running",
            "Self": {
                "TailscaleIPs": ["100.88.1.42", "fd7a:115c:a1e0::1"],
                "DNSName": "laptop-ku.tailnet-abc.ts.net."
            }
        });
        let snap = build_endpoints_snapshot(
            &status,
            vec!["192.168.1.10".into()],
            true,
            8443,
            "srv123".into(),
            "fp456".into(),
        );
        assert_eq!(snap["tailscale_url"], "https://100.88.1.42:8443");
        assert_eq!(snap["tailscale_dns"], "laptop-ku.tailnet-abc.ts.net");
        assert_eq!(
            snap["tailscale_urls"],
            json!([
                "https://100.88.1.42:8443",
                "https://laptop-ku.tailnet-abc.ts.net:8443"
            ])
        );
        assert_eq!(snap["tailscale_state"], "Running");
        assert_eq!(snap["tailscale_installed"], true);
        assert_eq!(snap["server_id"], "srv123");
        assert_eq!(snap["version"], "5.0.0");
        assert_eq!(snap["fingerprint"], "fp456");

        let not_installed = build_endpoints_snapshot(
            &Value::Null,
            vec![],
            false,
            8443,
            "srv123".into(),
            "fp456".into(),
        );
        assert_eq!(not_installed["tailscale_state"], "NotInstalled");
        assert_eq!(not_installed["tailscale_installed"], false);
        assert_eq!(not_installed["tailscale_urls"], json!([]));
    }
    #[tokio::test]
    async fn connection_endpoint_supports_refresh_query() {
        let (_dir, a, h, _) = fixture();
        let res = api(
            State(a),
            Method::GET,
            "/connection?refresh=1".parse().unwrap(),
            h,
            Bytes::new(),
        )
        .await;
        assert_eq!(res.status(), StatusCode::OK);
        let body = axum::body::to_bytes(res.into_body(), 65536).await.unwrap();
        let parsed: Value = serde_json::from_slice(&body).unwrap();
        assert_eq!(parsed["version"], "5.0.0");
        assert!(parsed.get("tailscale_urls").is_some());
        assert!(parsed.get("tailscale_dns").is_some());
        assert!(parsed.get("tailscale_installed").is_some());
        assert!(parsed.get("fingerprint").is_some());
    }
}
