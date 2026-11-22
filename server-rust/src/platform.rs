//! Native Windows operations. Call on a blocking worker, never a Tokio reactor.
//! COM objects are created and released on the calling thread; no COM pointers
//! are shared between requests. Session 0 is rejected for interactive operations.
use anyhow::{Context, Result, bail, ensure};
use serde_json::{Value, json};
#[cfg(windows)]
#[path = "platform_overlay.rs"]
mod overlay;

fn number(body: &Value, key: &str, max: f64) -> Result<f64> {
    let value = body
        .get(key)
        .and_then(Value::as_f64)
        .context(format!("{key} must be a number"))?;
    ensure!(
        value.is_finite() && value >= 0.0 && value <= max,
        "{key} out of range (0..{max})"
    );
    Ok(value)
}
fn boolean(body: &Value, key: &str) -> Result<bool> {
    body.get(key)
        .and_then(Value::as_bool)
        .context(format!("{key} must be a boolean"))
}
fn text<'a>(body: &'a Value, key: &str) -> Result<&'a str> {
    let value = body
        .get(key)
        .and_then(Value::as_str)
        .context(format!("{key} must be a string"))?;
    ensure!(
        !value.is_empty() && !value.contains('\0') && value.len() <= 8192,
        "invalid {key}"
    );
    Ok(value)
}
fn browser_url(body: &Value) -> Result<String> {
    let input = text(body, "url")?;
    let url = url::Url::parse(input)?;
    ensure!(
        matches!(url.scheme(), "http" | "https") && url.host_str().is_some(),
        "only HTTP(S) URLs are allowed"
    );
    ensure!(
        url.username().is_empty() && url.password().is_none(),
        "URL credentials are not allowed"
    );
    Ok(url.to_string())
}

#[cfg(not(windows))]
pub fn execute(_path: &str, _body: &Value) -> Result<Value> {
    bail!("Windows is required")
}

#[cfg(windows)]
pub fn execute(path: &str, body: &Value) -> Result<Value> {
    native::execute(path, body)
}

/// Retain WinRT subscriptions on a parked worker. Callbacks only enqueue a
/// notification; the transport fetches a snapshot on its blocking executor.
#[cfg(windows)]
pub fn start_event_listener(notify: impl Fn() + Send + Sync + 'static) {
    std::thread::spawn(move || {
        if let Err(error) = native::listen(std::sync::Arc::new(notify)) {
            eprintln!("Media events unavailable: {error:#}");
        }
    });
}
#[cfg(not(windows))]
pub fn start_event_listener(_notify: impl Fn() + Send + Sync + 'static) {}

#[cfg(windows)]
mod native {
    use super::*;
    use std::{collections::HashMap, os::windows::process::CommandExt, process::Command};
    use windows::{
        Media::Control::GlobalSystemMediaTransportControlsSessionManager as MediaManager,
        Win32::{
            Devices::FunctionDiscovery::PKEY_Device_FriendlyName,
            Media::Audio::{Endpoints::IAudioEndpointVolume, *},
            System::{
                Com::{
                    StructuredStorage::{PropVariantClear, PropVariantToStringAlloc},
                    *,
                },
                Power::SetSuspendState,
                RemoteDesktop::ProcessIdToSessionId,
                Shutdown::LockWorkStation,
                Threading::GetCurrentProcessId,
            },
            UI::{Shell::ShellExecuteW, WindowsAndMessaging::*},
        },
        core::{HSTRING, w},
    };

    struct Com;
    impl Com {
        fn init() -> Result<Self> {
            unsafe {
                CoInitializeEx(None, COINIT_MULTITHREADED).ok()?;
            }
            Ok(Self)
        }
    }
    impl Drop for Com {
        fn drop(&mut self) {
            unsafe { CoUninitialize() }
        }
    }

    #[windows::core::implement(
        windows::Win32::Media::Audio::Endpoints::IAudioEndpointVolumeCallback
    )]
    struct VolumeChanged(std::sync::Arc<dyn Fn() + Send + Sync>);
    impl windows::Win32::Media::Audio::Endpoints::IAudioEndpointVolumeCallback_Impl
        for VolumeChanged_Impl
    {
        fn OnNotify(
            &self,
            _data: *mut AUDIO_VOLUME_NOTIFICATION_DATA,
        ) -> windows::core::Result<()> {
            (self.0)();
            Ok(())
        }
    }

    #[windows::core::implement(IMMNotificationClient)]
    struct DevicesChanged(std::sync::mpsc::SyncSender<()>);
    impl DevicesChanged_Impl {
        fn changed(&self) -> windows::core::Result<()> {
            let _ = self.0.try_send(());
            Ok(())
        }
    }
    impl IMMNotificationClient_Impl for DevicesChanged_Impl {
        fn OnDeviceStateChanged(
            &self,
            _id: &windows::core::PCWSTR,
            _state: DEVICE_STATE,
        ) -> windows::core::Result<()> {
            self.changed()
        }
        fn OnDeviceAdded(&self, _id: &windows::core::PCWSTR) -> windows::core::Result<()> {
            self.changed()
        }
        fn OnDeviceRemoved(&self, _id: &windows::core::PCWSTR) -> windows::core::Result<()> {
            self.changed()
        }
        fn OnDefaultDeviceChanged(
            &self,
            _flow: EDataFlow,
            _role: ERole,
            _id: &windows::core::PCWSTR,
        ) -> windows::core::Result<()> {
            self.changed()
        }
        fn OnPropertyValueChanged(
            &self,
            _id: &windows::core::PCWSTR,
            _key: &windows::Win32::Foundation::PROPERTYKEY,
        ) -> windows::core::Result<()> {
            self.changed()
        }
    }

    pub fn listen(notify: std::sync::Arc<dyn Fn() + Send + Sync>) -> Result<()> {
        use windows::{Foundation::TypedEventHandler, Media::Control::*};
        let _com = Com::init()?;
        let manager = MediaManager::RequestAsync()?.join()?;
        let (tx, rx) = std::sync::mpsc::sync_channel(1);
        let en = unsafe { enumerator()? };
        let devices_callback: IMMNotificationClient = DevicesChanged(tx.clone()).into();
        unsafe {
            en.RegisterEndpointNotificationCallback(&devices_callback)?;
        }
        // Rebuild these subscriptions on hotplug and default-device changes.
        let mut audio_subscriptions: Vec<(
            IAudioEndpointVolume,
            windows::Win32::Media::Audio::Endpoints::IAudioEndpointVolumeCallback,
        )> = Vec::new();
        let change_tx = tx.clone();
        let manager_token =
            manager.CurrentSessionChanged(&TypedEventHandler::new(move |_, _| {
                let _ = change_tx.try_send(());
                Ok(())
            }))?;
        let mut previous: Option<(GlobalSystemMediaTransportControlsSession, i64, i64)> = None;
        let _ = tx.try_send(());
        while rx.recv().is_ok() {
            unsafe {
                for (endpoint, callback) in audio_subscriptions.drain(..) {
                    let _ = endpoint.UnregisterControlChangeNotify(&callback);
                }
                for device in all_devices(&en)? {
                    if let Ok(endpoint) = volume(&device) {
                        let callback:windows::Win32::Media::Audio::Endpoints::IAudioEndpointVolumeCallback=VolumeChanged(notify.clone()).into();
                        if endpoint.RegisterControlChangeNotify(&callback).is_ok() {
                            audio_subscriptions.push((endpoint, callback));
                        }
                    }
                }
            }
            if let Some((session, playback, properties)) = previous.take() {
                let _ = session.RemovePlaybackInfoChanged(playback);
                let _ = session.RemoveMediaPropertiesChanged(properties);
            }
            if let Ok(session) = manager.GetCurrentSession() {
                let changed = notify.clone();
                let playback =
                    session.PlaybackInfoChanged(&TypedEventHandler::new(move |_, _| {
                        changed();
                        Ok(())
                    }))?;
                let changed = notify.clone();
                let properties =
                    session.MediaPropertiesChanged(&TypedEventHandler::new(move |_, _| {
                        changed();
                        Ok(())
                    }))?;
                previous = Some((session, playback, properties));
            }
            notify();
        }
        let _ = manager.RemoveCurrentSessionChanged(manager_token);
        unsafe {
            let _ = en.UnregisterEndpointNotificationCallback(&devices_callback);
        }
        for (endpoint, callback) in audio_subscriptions {
            unsafe {
                let _ = endpoint.UnregisterControlChangeNotify(&callback);
            }
        }
        Ok(())
    }

    fn interactive() -> Result<()> {
        let mut session = 0;
        unsafe {
            ProcessIdToSessionId(GetCurrentProcessId(), &mut session)?;
        }
        ensure!(
            session != 0,
            "desktop session unavailable: run the agent in the signed-in user's session"
        );
        Ok(())
    }
    unsafe fn enumerator() -> Result<IMMDeviceEnumerator> {
        Ok(unsafe { CoCreateInstance(&MMDeviceEnumerator, None, CLSCTX_ALL)? })
    }
    unsafe fn name(device: &IMMDevice) -> Result<String> {
        unsafe {
            let store = device.OpenPropertyStore(STGM_READ)?;
            let mut value = store.GetValue(&PKEY_Device_FriendlyName)?;
            let allocated = PropVariantToStringAlloc(&value);
            let _ = PropVariantClear(&mut value);
            let ptr = allocated?;
            let result = ptr.to_string();
            CoTaskMemFree(Some(ptr.0.cast()));
            Ok(result?)
        }
    }
    unsafe fn all_devices(en: &IMMDeviceEnumerator) -> Result<Vec<IMMDevice>> {
        unsafe {
            let collection = en.EnumAudioEndpoints(eAll, DEVICE_STATE_ACTIVE)?;
            (0..collection.GetCount()?)
                .map(|i| collection.Item(i).map_err(Into::into))
                .collect()
        }
    }
    fn channel_name(channel: &str) -> Result<&'static str> {
        match channel {
            "media" => Ok("Sonar - Media"),
            "gaming" => Ok("Sonar - Gaming"),
            "chat" => Ok("Sonar - Chat"),
            "mic" => Ok("Sonar - Microphone"),
            _ => bail!("unknown channel"),
        }
    }
    unsafe fn target(en: &IMMDeviceEnumerator, body: &Value, path: &str) -> Result<IMMDevice> {
        unsafe {
            if path.contains("/device/") {
                return Ok(en.GetDevice(&HSTRING::from(text(body, "device_id")?))?);
            }
            let wanted = if path.contains("/channel/") {
                channel_name(text(body, "channel")?)?
            } else {
                "Sonar - Media"
            };
            for device in all_devices(en)? {
                if name(&device)?.contains(wanted) {
                    return Ok(device);
                }
            }
            ensure!(
                !path.contains("/channel/"),
                "requested channel is not installed"
            );
            Ok(en.GetDefaultAudioEndpoint(eRender, eMultimedia)?)
        }
    }
    unsafe fn volume(device: &IMMDevice) -> Result<IAudioEndpointVolume> {
        Ok(unsafe { device.Activate(CLSCTX_ALL, None)? })
    }
    fn audio(path: &str, body: &Value) -> Result<Value> {
        let _com = Com::init()?;
        unsafe {
            let en = enumerator()?;
            if path == "/audio/devices" || path == "/audio/channels" {
                let mut devices = Vec::new();
                let mut channels = serde_json::Map::new();
                for device in all_devices(&en)? {
                    let Ok(ep) = volume(&device) else { continue };
                    let label = name(&device)?;
                    let level = ep.GetMasterVolumeLevelScalar()?;
                    let muted = ep.GetMute()?.as_bool();
                    let ptr = device.GetId()?;
                    let id = ptr.to_string();
                    CoTaskMemFree(Some(ptr.0.cast()));
                    devices
                        .push(json!({"id": id?, "name": label, "volume": level, "muted": muted}));
                    for channel in ["media", "gaming", "chat", "mic"] {
                        if label.contains(channel_name(channel)?) {
                            channels.insert(
                                channel.into(),
                                json!({"name":label,"level":level,"muted":muted}),
                            );
                        }
                    }
                }
                return Ok(if path.ends_with("devices") {
                    json!({"devices":devices})
                } else {
                    json!({"channels":channels})
                });
            }
            let ep = volume(&target(&en, body, path)?)?;
            if path.ends_with("/status") {
                return Ok(
                    json!({"level":ep.GetMasterVolumeLevelScalar()?,"muted":ep.GetMute()?.as_bool()}),
                );
            }
            if path.ends_with("/volume") {
                ep.SetMasterVolumeLevelScalar(
                    number(body, "level", 1.0)? as f32,
                    std::ptr::null(),
                )?;
            } else if path.ends_with("/mute") {
                let key = if body.get("mute").is_some() {
                    "mute"
                } else {
                    "muted"
                };
                ep.SetMute(boolean(body, key)?, std::ptr::null())?;
            } else {
                bail!("unknown audio operation");
            }
        }
        Ok(json!({"success":true}))
    }
    fn media(path: &str) -> Result<Value> {
        let _com = Com::init()?;
        let manager = MediaManager::RequestAsync()?.join()?;
        let session = manager.GetCurrentSession();
        if path == "/media/status" {
            let Ok(session) = session else {
                return Ok(
                    json!({"success":true,"status":"Closed","title":"","artist":"","album":""}),
                );
            };
            let properties = session.TryGetMediaPropertiesAsync()?.join()?;
            let playback = session.GetPlaybackInfo()?.PlaybackStatus()?;
            let status = match playback.0 {
                4 => "Playing",
                5 => "Paused",
                3 => "Stopped",
                _ => "Closed",
            };
            return Ok(
                json!({"success":true,"status":status,"title":properties.Title()?.to_string(),"artist":properties.Artist()?.to_string(),"album":properties.AlbumTitle()?.to_string(),"app_id":session.SourceAppUserModelId()?.to_string()}),
            );
        }
        let session = session.context("no active media session")?;
        let accepted = match path {
            "/media/play" => session.TryTogglePlayPauseAsync()?.join()?,
            "/media/next" => session.TrySkipNextAsync()?.join()?,
            "/media/prev" => session.TrySkipPreviousAsync()?.join()?,
            "/media/seek-forward" | "/media/seek-backward" => {
                let timeline = session.GetTimelineProperties()?;
                let direction = if path.ends_with("forward") { 1 } else { -1 };
                let position = (timeline.Position()?.Duration + direction * 100_000_000)
                    .clamp(timeline.StartTime()?.Duration, timeline.EndTime()?.Duration);
                session.TryChangePlaybackPositionAsync(position)?.join()?
            }
            _ => bail!("unknown media operation"),
        };
        ensure!(accepted, "media application rejected the command");
        Ok(json!({"success":true}))
    }
    fn shutdown(args: &[&str]) -> Result<()> {
        // shutdown.exe is a fixed system executable, only invoked on explicit command.
        // No PowerShell, shell interpretation, polling process or arbitrary arguments.
        let executable = std::env::var_os("SystemRoot").context("SystemRoot unavailable")?;
        let status =
            Command::new(std::path::PathBuf::from(executable).join("System32/shutdown.exe"))
                .args(args)
                .creation_flags(0x08000000)
                .status()?;
        ensure!(
            status.success(),
            "Windows rejected shutdown operation ({status})"
        );
        Ok(())
    }
    fn sleep() -> Result<()> {
        use windows::Win32::{
            Foundation::{CloseHandle, ERROR_NOT_ALL_ASSIGNED, GetLastError, HANDLE, LUID},
            Security::*,
            System::Threading::{GetCurrentProcess, OpenProcessToken},
        };
        unsafe {
            let mut token = HANDLE::default();
            OpenProcessToken(
                GetCurrentProcess(),
                TOKEN_ADJUST_PRIVILEGES | TOKEN_QUERY,
                &mut token,
            )?;
            let result = (|| -> Result<()> {
                let mut luid = LUID::default();
                LookupPrivilegeValueW(None, SE_SHUTDOWN_NAME, &mut luid)?;
                let state = TOKEN_PRIVILEGES {
                    PrivilegeCount: 1,
                    Privileges: [LUID_AND_ATTRIBUTES {
                        Luid: luid,
                        Attributes: SE_PRIVILEGE_ENABLED,
                    }],
                };
                let mut previous = TOKEN_PRIVILEGES::default();
                let mut length = 0;
                AdjustTokenPrivileges(
                    token,
                    false,
                    Some(&state),
                    std::mem::size_of::<TOKEN_PRIVILEGES>() as u32,
                    Some(&mut previous),
                    Some(&mut length),
                )?;
                ensure!(
                    GetLastError() != ERROR_NOT_ALL_ASSIGNED,
                    "shutdown privilege unavailable"
                );
                let slept = SetSuspendState(false, false, false);
                let error = std::io::Error::last_os_error();
                let _ = AdjustTokenPrivileges(token, false, Some(&previous), 0, None, None);
                ensure!(slept, "Windows rejected sleep: {error}");
                Ok(())
            })();
            let _ = CloseHandle(token);
            result
        }
    }
    fn hardware_brightness(body: &Value) -> Result<Value> {
        // WMI's COMLibrary intentionally does not uninitialize; balance that
        // initialization explicitly on this worker once WMI objects are dropped.
        let library = wmi::COMLibrary::new()?;
        let _com = Com;
        let connection = wmi::WMIConnection::with_namespace_path("ROOT\\WMI", library)?;
        if body.get("level").is_some() {
            let level = number(body, "level", 100.0)?.round() as u8;
            let instances: Vec<HashMap<String, wmi::Variant>> = connection
                .raw_query("SELECT __PATH FROM WmiMonitorBrightnessMethods WHERE Active = TRUE")?;
            ensure!(
                !instances.is_empty(),
                "hardware brightness is not supported by this monitor"
            );
            for instance in instances {
                let Some(wmi::Variant::String(path)) = instance.get("__PATH") else {
                    bail!("missing brightness instance path")
                };
                #[derive(serde::Deserialize)]
                #[serde(rename = "WmiMonitorBrightnessMethods")]
                struct BrightnessMethods;
                #[derive(serde::Serialize)]
                struct Params {
                    #[serde(rename = "Timeout")]
                    timeout: u32,
                    #[serde(rename = "Brightness")]
                    brightness: u8,
                }
                #[derive(serde::Deserialize)]
                struct Output {
                    #[serde(rename = "ReturnValue")]
                    result: u32,
                }
                let output: Output = connection.exec_instance_method::<BrightnessMethods, Output>(
                    path,
                    "WmiSetBrightness",
                    Params {
                        timeout: 0,
                        brightness: level,
                    },
                )?;
                ensure!(output.result == 0, "brightness rejected: {}", output.result);
            }
            return Ok(json!({"success":true,"brightness":level}));
        }
        #[derive(serde::Deserialize)]
        struct Reading {
            #[serde(rename = "CurrentBrightness")]
            brightness: u8,
        }
        let readings: Vec<Reading> = connection
            .raw_query("SELECT CurrentBrightness FROM WmiMonitorBrightness WHERE Active = TRUE")?;
        Ok(
            json!({"brightness":readings.first().context("hardware brightness is not supported by this monitor")?.brightness}),
        )
    }
    fn brightness(body: &Value) -> Result<Value> {
        if body.get("level").is_some() {
            let level = number(body, "level", 100.0)?;
            if level >= 50.0 {
                // Always allow users to remove software dimming, even on monitors
                // without hardware brightness support.
                super::overlay::set(0)?;
                // Report an unsupported monitor instead of claiming a successful hardware change.
                hardware_brightness(&json!({"level":(level-50.0)*2.0}))?;
            } else {
                // Software dimming still works on external monitors without WMI support.
                let _ = hardware_brightness(&json!({"level":0.0}));
                super::overlay::set(((50.0 - level) * 255.0 / 50.0).round() as u8)?;
            }
            return Ok(json!({"success":true,"brightness":level}));
        }
        let alpha = super::overlay::alpha();
        if alpha > 0 {
            return Ok(json!({"brightness":50.0-f64::from(alpha)*50.0/255.0}));
        }
        let hardware = hardware_brightness(&json!({}))?;
        Ok(
            json!({"brightness":50.0+hardware["brightness"].as_f64().context("invalid brightness")?/2.0}),
        )
    }
    pub fn execute(path: &str, body: &Value) -> Result<Value> {
        interactive()?;
        if path.starts_with("/audio/") {
            return audio(path, body);
        }
        if path.starts_with("/media/") {
            return media(path);
        }
        match path {
            "/system/brightness" => return brightness(body),
            "/system/lock" => unsafe {
                LockWorkStation()?;
            },
            "/system/display/off" => super::overlay::set(255)?,
            "/system/sleep" => sleep()?,
            "/system/restart" => shutdown(&["/r", "/t", "5"])?,
            "/system/shutdown/cancel" => shutdown(&["/a"])?,
            "/system/shutdown" => {
                let seconds = if body.get("delay_seconds").is_some() {
                    number(body, "delay_seconds", 86400.0)?
                } else if body.get("delay_minutes").is_some() {
                    number(body, "delay_minutes", 1440.0)? * 60.0
                } else {
                    0.0
                };
                ensure!(seconds.fract() == 0.0, "delay must be integral");
                shutdown(&["/s", "/t", &format!("{seconds:.0}")])?;
            }
            "/browser/open" => {
                let url = HSTRING::from(browser_url(body)?);
                let result =
                    unsafe { ShellExecuteW(None, w!("open"), &url, None, None, SW_SHOWNORMAL) };
                ensure!(result.0 as isize > 32, "Windows failed to open browser");
            }
            _ => bail!("unknown operation"),
        }
        Ok(json!({"success":true}))
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    #[cfg(windows)]
    #[test]
    #[ignore = "requires an interactive Windows desktop and audio hardware; read-only"]
    fn native_read_only_smoke() {
        let status = execute("/audio/status", &json!({})).unwrap();
        assert!((0.0..=1.0).contains(&status["level"].as_f64().unwrap()));
        assert!(status["muted"].is_boolean());
        let devices = execute("/audio/devices", &json!({})).unwrap();
        assert!(devices["devices"].is_array());
        let media = execute("/media/status", &json!({})).unwrap();
        assert!(media["status"].is_string());
    }
    #[test]
    fn rejects_invalid_levels() {
        for body in [
            json!({}),
            json!({"level":"1"}),
            json!({"level":-1}),
            json!({"level":1.01}),
        ] {
            assert!(number(&body, "level", 1.0).is_err());
        }
        assert_eq!(number(&json!({"level":0.75}), "level", 1.0).unwrap(), 0.75);
    }
    #[test]
    fn rejects_non_boolean_mute() {
        assert!(boolean(&json!({"muted":0}), "muted").is_err());
    }
    #[test]
    fn browser_does_not_accept_shell_or_file_urls() {
        for value in [
            "file:///C:/Windows/System32/cmd.exe",
            "javascript:alert(1)",
            "https://user:pass@example.com",
            "cmd.exe",
        ] {
            assert!(browser_url(&json!({"url":value})).is_err());
        }
        assert_eq!(
            browser_url(&json!({"url":"https://example.com/a?x=1&y=2"})).unwrap(),
            "https://example.com/a?x=1&y=2"
        );
    }
}
