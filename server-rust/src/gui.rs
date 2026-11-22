//! Native Win32/GDI dark-themed desktop window for PC Remote v5.0.
//! Owns its HWND and message pump on a dedicated thread; closing hides to tray.
use crate::auth::{Auth, now};
use serde_json::{Value, json};
use std::{
    net::SocketAddr,
    path::PathBuf,
    sync::{
        Arc, Mutex,
        atomic::{AtomicUsize, Ordering},
    },
    time::{Duration, Instant},
};
use windows::{
    Win32::{
        Foundation::{COLORREF, HINSTANCE, HWND, LPARAM, LRESULT, RECT, WPARAM},
        Graphics::Gdi::*,
        System::{
            LibraryLoader::GetModuleHandleW,
            Registry::{
                HKEY, HKEY_CURRENT_USER, KEY_READ, KEY_WRITE, REG_SZ, RegCloseKey, RegDeleteValueW,
                RegOpenKeyExW, RegQueryValueExW, RegSetValueExW,
            },
        },
        UI::{
            Controls::{DRAWITEMSTRUCT, ODS_SELECTED},
            Shell::ShellExecuteW,
            WindowsAndMessaging::*,
        },
    },
    core::{HSTRING, PCWSTR, w},
};

const WINDOW_CLASS: PCWSTR = w!("PCRemoteNativeWindow");
const RUN_KEY: PCWSTR = w!("Software\\Microsoft\\Windows\\CurrentVersion\\Run");
const RUN_VALUE: PCWSTR = w!("PCRemoteRust");
const TAILSCALE_DOWNLOAD_URL: PCWSTR = w!("https://tailscale.com/download/windows");
const TAILSCALE_ANDROID_URL: PCWSTR = w!("https://tailscale.com/download/android");

const WM_APP_SHOW_FOCUS: u32 = WM_APP + 10;
const WM_APP_QUIT: u32 = WM_APP + 11;
const EM_SETLIMITTEXT_MSG: u32 = 0x00C5;
const BST_UNCHECKED_VAL: usize = 0;
const BST_CHECKED_VAL: usize = 1;

const TIMER_COUNTDOWN_ID: usize = 1;
const TIMER_POLL_ID: usize = 2;

const ID_BTN_STOP: i32 = 101;
const ID_BTN_NEW_QR: i32 = 102;
const ID_BTN_TS_INSTALL: i32 = 103;
const ID_BTN_TS_LOGIN: i32 = 104;
const ID_LIST_DEVICES: i32 = 105;
const ID_BTN_REVOKE: i32 = 106;
const ID_EDIT_PIN: i32 = 107;
const ID_BTN_SAVE_PIN: i32 = 108;
const ID_CHK_AUTOSTART: i32 = 109;
const ID_BTN_TS_MOBILE: i32 = 110;

const WIN_W: i32 = 840;
const WIN_H: i32 = 636;

const fn rgb(r: u8, g: u8, b: u8) -> COLORREF {
    COLORREF((r as u32) | ((g as u32) << 8) | ((b as u32) << 16))
}

const COL_BG: COLORREF = rgb(15, 23, 42);
const COL_CARD: COLORREF = rgb(30, 41, 59);
const COL_CARD_BORDER: COLORREF = rgb(51, 65, 85);
const COL_INPUT_BG: COLORREF = rgb(15, 23, 42);
const COL_TEXT: COLORREF = rgb(248, 250, 252);
const COL_MUTED: COLORREF = rgb(148, 163, 184);
const COL_ACCENT: COLORREF = rgb(56, 189, 248);
const COL_BTN_PRIMARY: COLORREF = rgb(37, 99, 235);
const COL_BTN_PRIMARY_PRESS: COLORREF = rgb(29, 78, 216);
const COL_BTN_SLATE: COLORREF = rgb(51, 65, 85);
const COL_BTN_SLATE_PRESS: COLORREF = rgb(71, 85, 105);
const COL_BTN_DANGER: COLORREF = rgb(220, 38, 38);
const COL_BTN_DANGER_PRESS: COLORREF = rgb(185, 28, 28);
const COL_GREEN: COLORREF = rgb(34, 197, 94);
const COL_AMBER: COLORREF = rgb(245, 158, 11);
const COL_RED: COLORREF = rgb(239, 68, 68);
const COL_WHITE: COLORREF = rgb(255, 255, 255);
const COL_BLACK: COLORREF = rgb(0, 0, 0);

static GUI_HWND: AtomicUsize = AtomicUsize::new(0);

#[derive(Clone)]
pub struct GuiContext {
    pub auth: Arc<Mutex<Auth>>,
    pub endpoints: Arc<Mutex<Value>>,
    pub port: u16,
    pub shutdown: axum_server::Handle<SocketAddr>,
    pub refresh_notify: Arc<tokio::sync::Notify>,
}

struct UiState {
    ctx: GuiContext,
    qr_modules: Vec<bool>,
    qr_width: usize,
    qr_expires_at: Option<Instant>,
    qr_locked: bool,
    device_ids: Vec<String>,
    status_message: String,
    status_is_error: bool,
    status_until: Option<Instant>,
    ts_installed_cached: bool,
    font_title: HFONT,
    font_heading: HFONT,
    font_body: HFONT,
    font_small: HFONT,
    brush_bg: HBRUSH,
    brush_card: HBRUSH,
    brush_input: HBRUSH,
    hwnd_devices: HWND,
    hwnd_pin: HWND,
    hwnd_autostart: HWND,
    hwnd_btn_install: HWND,
    hwnd_btn_new_qr: HWND,
}

impl Drop for UiState {
    fn drop(&mut self) {
        unsafe {
            for font in [
                self.font_title,
                self.font_heading,
                self.font_body,
                self.font_small,
            ] {
                if !font.0.is_null() {
                    let _ = DeleteObject(font.into());
                }
            }
            for brush in [self.brush_bg, self.brush_card, self.brush_input] {
                if !brush.0.is_null() {
                    let _ = DeleteObject(brush.into());
                }
            }
        }
    }
}

pub fn start(ctx: GuiContext, show_initially: bool) {
    std::thread::spawn(move || {
        if let Err(error) = run_window(ctx, show_initially) {
            eprintln!("GUI unavailable: {error:#}");
        }
    });
}

pub fn show_or_focus() {
    let raw = GUI_HWND.load(Ordering::Acquire);
    if raw != 0 {
        unsafe {
            let hwnd = HWND(raw as *mut _);
            let _ = PostMessageW(Some(hwnd), WM_APP_SHOW_FOCUS, WPARAM(0), LPARAM(0));
        }
    }
}

pub fn focus_existing() {
    unsafe {
        if let Ok(hwnd) = FindWindowW(WINDOW_CLASS, None)
            && !hwnd.0.is_null()
        {
            let _ = ShowWindow(hwnd, SW_RESTORE);
            let _ = SetForegroundWindow(hwnd);
            let _ = PostMessageW(Some(hwnd), WM_APP_SHOW_FOCUS, WPARAM(0), LPARAM(0));
        }
    }
}

pub fn quit() {
    let raw = GUI_HWND.load(Ordering::Acquire);
    if raw != 0 {
        unsafe {
            let hwnd = HWND(raw as *mut _);
            let _ = PostMessageW(Some(hwnd), WM_APP_QUIT, WPARAM(0), LPARAM(0));
        }
    }
}

unsafe fn create_font(height: i32, weight: i32) -> HFONT {
    unsafe {
        CreateFontW(
            height,
            0,
            0,
            0,
            weight,
            0,
            0,
            0,
            DEFAULT_CHARSET,
            OUT_DEFAULT_PRECIS,
            CLIP_DEFAULT_PRECIS,
            CLEARTYPE_QUALITY,
            (DEFAULT_PITCH.0 | FF_SWISS.0) as u32,
            w!("Segoe UI"),
        )
    }
}

unsafe fn load_app_icon(instance: HINSTANCE, cx: i32, cy: i32) -> HICON {
    unsafe {
        LoadImageW(
            Some(instance),
            PCWSTR(1 as *const u16),
            IMAGE_ICON,
            cx,
            cy,
            LR_SHARED,
        )
        .ok()
        .map(|h| HICON(h.0))
        .or_else(|| LoadIconW(None, IDI_APPLICATION).ok())
        .unwrap_or_default()
    }
}

fn run_window(ctx: GuiContext, show_initially: bool) -> anyhow::Result<()> {
    unsafe {
        let instance: HINSTANCE = GetModuleHandleW(None)?.into();
        let big_icon = load_app_icon(
            instance,
            GetSystemMetrics(SM_CXICON),
            GetSystemMetrics(SM_CYICON),
        );
        let small_icon = load_app_icon(
            instance,
            GetSystemMetrics(SM_CXSMICON),
            GetSystemMetrics(SM_CYSMICON),
        );
        let class = WNDCLASSW {
            lpfnWndProc: Some(window_proc),
            hInstance: instance,
            lpszClassName: WINDOW_CLASS,
            hCursor: LoadCursorW(None, IDC_ARROW)?,
            hIcon: big_icon,
            hbrBackground: CreateSolidBrush(COL_BG),
            ..Default::default()
        };
        if RegisterClassW(&class) == 0 {
            let err = std::io::Error::last_os_error();
            // ERROR_CLASS_ALREADY_EXISTS = 1410
            if err.raw_os_error() != Some(1410) {
                anyhow::bail!("cannot register GUI window class: {err}");
            }
        }

        let screen_w = GetSystemMetrics(SM_CXSCREEN);
        let screen_h = GetSystemMetrics(SM_CYSCREEN);
        let x = ((screen_w - WIN_W) / 2).max(0);
        let y = ((screen_h - WIN_H) / 2).max(0);

        let mut state = Box::new(UiState {
            ctx,
            qr_modules: Vec::new(),
            qr_width: 0,
            qr_expires_at: None,
            qr_locked: false,
            device_ids: Vec::new(),
            status_message: String::new(),
            status_is_error: false,
            status_until: None,
            ts_installed_cached: false,
            font_title: create_font(-22, FW_BOLD.0 as i32),
            font_heading: create_font(-16, FW_SEMIBOLD.0 as i32),
            font_body: create_font(-14, FW_NORMAL.0 as i32),
            font_small: create_font(-12, FW_NORMAL.0 as i32),
            brush_bg: CreateSolidBrush(COL_BG),
            brush_card: CreateSolidBrush(COL_CARD),
            brush_input: CreateSolidBrush(COL_INPUT_BG),
            hwnd_devices: HWND::default(),
            hwnd_pin: HWND::default(),
            hwnd_autostart: HWND::default(),
            hwnd_btn_install: HWND::default(),
            hwnd_btn_new_qr: HWND::default(),
        });

        state.generate_qr();

        let state_ptr = Box::into_raw(state);
        let style = WS_OVERLAPPED | WS_CAPTION | WS_SYSMENU | WS_MINIMIZEBOX | WS_CLIPCHILDREN;
        let hwnd = CreateWindowExW(
            WINDOW_EX_STYLE::default(),
            WINDOW_CLASS,
            w!("PC Remote v5.0"),
            style,
            x,
            y,
            WIN_W,
            WIN_H,
            None,
            None,
            Some(instance),
            Some(state_ptr.cast()),
        )?;

        GUI_HWND.store(hwnd.0 as usize, Ordering::Release);
        SendMessageW(
            hwnd,
            WM_SETICON,
            Some(WPARAM(ICON_BIG as usize)),
            Some(LPARAM(big_icon.0 as isize)),
        );
        SendMessageW(
            hwnd,
            WM_SETICON,
            Some(WPARAM(ICON_SMALL as usize)),
            Some(LPARAM(small_icon.0 as isize)),
        );

        if show_initially {
            let _ = ShowWindow(hwnd, SW_SHOW);
            let _ = SetForegroundWindow(hwnd);
            let _ = SetTimer(Some(hwnd), TIMER_COUNTDOWN_ID, 1000, None);
            let _ = SetTimer(Some(hwnd), TIMER_POLL_ID, 2000, None);
        }

        let mut msg = MSG::default();
        while GetMessageW(&mut msg, None, 0, 0).0 > 0 {
            let _ = TranslateMessage(&msg);
            DispatchMessageW(&msg);
        }
        GUI_HWND.store(0, Ordering::Release);
    }
    Ok(())
}

impl UiState {
    fn has_active_device(&self) -> bool {
        let auth = self.ctx.auth.lock().unwrap();
        auth.store.devices.iter().any(|d| d.expires_at > now())
    }

    fn clear_qr(&mut self) {
        self.qr_modules.clear();
        self.qr_width = 0;
        self.qr_expires_at = None;
    }

    unsafe fn sync_qr_lock_state(&mut self, hwnd: HWND) {
        let locked = self.has_active_device();
        if locked {
            self.clear_qr();
            unsafe {
                let _ = ShowWindow(self.hwnd_btn_new_qr, SW_HIDE);
            }
            if !self.qr_locked {
                self.qr_locked = true;
                let qr_card = RECT {
                    left: 20,
                    top: 78,
                    right: 380,
                    bottom: 592,
                };
                unsafe {
                    let _ = InvalidateRect(Some(hwnd), Some(&qr_card), false);
                }
            }
        } else {
            unsafe {
                let _ = ShowWindow(self.hwnd_btn_new_qr, SW_SHOW);
            }
            if self.qr_locked {
                self.qr_locked = false;
                self.generate_qr();
                let qr_card = RECT {
                    left: 20,
                    top: 78,
                    right: 380,
                    bottom: 592,
                };
                unsafe {
                    let _ = InvalidateRect(Some(hwnd), Some(&qr_card), false);
                }
            }
        }
    }

    fn generate_qr(&mut self) {
        if self.has_active_device() {
            self.clear_qr();
            self.qr_locked = true;
            return;
        }
        self.qr_locked = false;
        let ep = self.ctx.endpoints.lock().unwrap().clone();
        let hosts = ep["hosts"].as_array().cloned().unwrap_or_default();
        let host = hosts
            .first()
            .and_then(Value::as_str)
            .unwrap_or("127.0.0.1")
            .to_string();
        let tailscale_url = ep["tailscale_url"].as_str().unwrap_or("").to_string();

        let (pair_token, fingerprint, server_id) = {
            let mut auth = self.ctx.auth.lock().unwrap();
            let p = auth.new_pair();
            (
                p,
                auth.store.fingerprint.clone(),
                auth.store.server_id.clone(),
            )
        };

        let payload = json!({
            "host": host,
            "port": self.ctx.port,
            "protocol": "https",
            "pair_token": pair_token,
            "fingerprint": fingerprint,
            "server_name": "PC Remote Rust",
            "server_id": server_id,
            "protocol_version": 1,
            "alternate_hosts": hosts,
            "tailscale_url": tailscale_url,
        });

        if let Ok(qr) = qrcode::QrCode::new(payload.to_string()) {
            let width = qr.width();
            let colors = qr.to_colors();
            self.qr_width = width;
            self.qr_modules = colors
                .into_iter()
                .map(|c| c == qrcode::Color::Dark)
                .collect();
            self.qr_expires_at = Some(Instant::now() + Duration::from_secs(120));
        }
    }

    fn qr_remaining_secs(&self) -> u64 {
        match self.qr_expires_at {
            Some(exp) => exp.saturating_duration_since(Instant::now()).as_secs(),
            None => 0,
        }
    }

    fn set_feedback(&mut self, msg: impl Into<String>, is_error: bool) {
        self.status_message = msg.into();
        self.status_is_error = is_error;
        self.status_until = Some(Instant::now() + Duration::from_secs(4));
    }

    unsafe fn refresh_devices_listbox(&mut self) {
        unsafe {
            let devices = {
                let auth = self.ctx.auth.lock().unwrap();
                auth.store
                    .devices
                    .iter()
                    .filter(|d| d.expires_at > now())
                    .cloned()
                    .collect::<Vec<_>>()
            };
            let new_ids: Vec<String> = devices.iter().map(|d| d.id.clone()).collect();
            if new_ids == self.device_ids {
                return;
            }
            let _ = SendMessageW(
                self.hwnd_devices,
                LB_RESETCONTENT,
                Some(WPARAM(0)),
                Some(LPARAM(0)),
            );
            self.device_ids = new_ids;
            for d in devices {
                let short_id: String = d.id.chars().take(8).collect();
                let label = format!("{} ({short_id})", d.name);
                let wide: Vec<u16> = label.encode_utf16().chain(Some(0)).collect();
                let _ = SendMessageW(
                    self.hwnd_devices,
                    LB_ADDSTRING,
                    Some(WPARAM(0)),
                    Some(LPARAM(wide.as_ptr() as isize)),
                );
            }
            if !self.device_ids.is_empty() {
                let _ = SendMessageW(
                    self.hwnd_devices,
                    LB_SETCURSEL,
                    Some(WPARAM(0)),
                    Some(LPARAM(0)),
                );
            }
        }
    }

    unsafe fn sync_install_button(&mut self) {
        let installed = self
            .ctx
            .endpoints
            .lock()
            .unwrap()
            .get("tailscale_installed")
            .and_then(Value::as_bool)
            .unwrap_or(false);
        if installed != self.ts_installed_cached {
            self.ts_installed_cached = installed;
            let text = if installed {
                w!("\u{2713} 1. Tailscale Terpasang")
            } else {
                w!("1. Unduh / Pasang Tailscale")
            };
            unsafe {
                let _ = SetWindowTextW(self.hwnd_btn_install, text);
            }
        }
    }
}

unsafe fn create_controls(
    hwnd: HWND,
    instance: HINSTANCE,
    state: &mut UiState,
) -> anyhow::Result<()> {
    unsafe {
        let btn = |id: i32, text: PCWSTR, x: i32, y: i32, w: i32, h: i32| -> anyhow::Result<HWND> {
            Ok(CreateWindowExW(
                WINDOW_EX_STYLE::default(),
                w!("BUTTON"),
                text,
                WS_CHILD | WS_VISIBLE | WS_TABSTOP | WINDOW_STYLE(BS_OWNERDRAW as u32),
                x,
                y,
                w,
                h,
                Some(hwnd),
                Some(HMENU(id as usize as *mut _)),
                Some(instance),
                None,
            )?)
        };

        btn(ID_BTN_STOP, w!("Hentikan Server"), 660, 22, 144, 34)?;
        state.hwnd_btn_new_qr = btn(ID_BTN_NEW_QR, w!("Buat QR Pairing Baru"), 52, 536, 296, 36)?;

        state.hwnd_btn_install = btn(
            ID_BTN_TS_INSTALL,
            w!("1. Unduh / Pasang Tailscale"),
            412,
            224,
            190,
            30,
        )?;
        btn(
            ID_BTN_TS_LOGIN,
            w!("2. Buka & Login Tailscale PC"),
            610,
            224,
            190,
            30,
        )?;
        btn(
            ID_BTN_TS_MOBILE,
            w!("3. Unduh Tailscale HP"),
            412,
            258,
            388,
            28,
        )?;

        state.hwnd_devices = CreateWindowExW(
            WS_EX_CLIENTEDGE,
            w!("LISTBOX"),
            w!(""),
            WS_CHILD | WS_VISIBLE | WS_VSCROLL | WS_TABSTOP | WINDOW_STYLE(LBS_NOTIFY as u32),
            412,
            384,
            262,
            78,
            Some(hwnd),
            Some(HMENU(ID_LIST_DEVICES as usize as *mut _)),
            Some(instance),
            None,
        )?;
        let _ = SendMessageW(
            state.hwnd_devices,
            WM_SETFONT,
            Some(WPARAM(state.font_body.0 as usize)),
            Some(LPARAM(1)),
        );

        btn(ID_BTN_REVOKE, w!("Cabut Akses"), 684, 384, 116, 34)?;

        state.hwnd_pin = CreateWindowExW(
            WS_EX_CLIENTEDGE,
            w!("EDIT"),
            w!(""),
            WS_CHILD
                | WS_VISIBLE
                | WS_TABSTOP
                | WINDOW_STYLE((ES_NUMBER | ES_PASSWORD | ES_AUTOHSCROLL) as u32),
            492,
            484,
            182,
            28,
            Some(hwnd),
            Some(HMENU(ID_EDIT_PIN as usize as *mut _)),
            Some(instance),
            None,
        )?;
        let _ = SendMessageW(
            state.hwnd_pin,
            EM_SETLIMITTEXT_MSG,
            Some(WPARAM(8)),
            Some(LPARAM(0)),
        );
        let _ = SendMessageW(
            state.hwnd_pin,
            WM_SETFONT,
            Some(WPARAM(state.font_body.0 as usize)),
            Some(LPARAM(1)),
        );

        btn(ID_BTN_SAVE_PIN, w!("Simpan PIN"), 684, 481, 116, 32)?;

        state.hwnd_autostart = CreateWindowExW(
            WINDOW_EX_STYLE::default(),
            w!("BUTTON"),
            w!("Jalankan otomatis saat login Windows"),
            WS_CHILD | WS_VISIBLE | WS_TABSTOP | WINDOW_STYLE(BS_AUTOCHECKBOX as u32),
            412,
            536,
            380,
            24,
            Some(hwnd),
            Some(HMENU(ID_CHK_AUTOSTART as usize as *mut _)),
            Some(instance),
            None,
        )?;
        let _ = SendMessageW(
            state.hwnd_autostart,
            WM_SETFONT,
            Some(WPARAM(state.font_body.0 as usize)),
            Some(LPARAM(1)),
        );
        let autostart_enabled = is_autostart_enabled();
        let _ = SendMessageW(
            state.hwnd_autostart,
            BM_SETCHECK,
            Some(WPARAM(if autostart_enabled {
                BST_CHECKED_VAL
            } else {
                BST_UNCHECKED_VAL
            })),
            Some(LPARAM(0)),
        );

        state.refresh_devices_listbox();
        state.sync_install_button();
        state.sync_qr_lock_state(hwnd);
    }
    Ok(())
}

fn is_autostart_enabled() -> bool {
    unsafe {
        let mut hkey = HKEY::default();
        if RegOpenKeyExW(HKEY_CURRENT_USER, RUN_KEY, Some(0), KEY_READ, &mut hkey).0 != 0 {
            return false;
        }
        let mut len: u32 = 0;
        let exists = RegQueryValueExW(hkey, RUN_VALUE, None, None, None, Some(&mut len)).0 == 0;
        let _ = RegCloseKey(hkey);
        exists
    }
}

fn set_autostart(enabled: bool) -> anyhow::Result<()> {
    unsafe {
        let mut hkey = HKEY::default();
        let rc = RegOpenKeyExW(HKEY_CURRENT_USER, RUN_KEY, Some(0), KEY_WRITE, &mut hkey);
        anyhow::ensure!(rc.0 == 0, "gagal membuka registry Run ({})", rc.0);
        let result = if enabled {
            let exe = std::env::current_exe()?;
            let cmd = format!("\"{}\" --background", exe.display());
            let wide: Vec<u16> = cmd.encode_utf16().chain(Some(0)).collect();
            let bytes = std::slice::from_raw_parts(wide.as_ptr() as *const u8, wide.len() * 2);
            let wr = RegSetValueExW(hkey, RUN_VALUE, Some(0), REG_SZ, Some(bytes));
            anyhow::ensure!(wr.0 == 0, "gagal menulis registry Run ({})", wr.0);
            Ok(())
        } else {
            let dr = RegDeleteValueW(hkey, RUN_VALUE);
            // ERROR_FILE_NOT_FOUND = 2 is fine when disabling.
            anyhow::ensure!(
                dr.0 == 0 || dr.0 == 2,
                "gagal menghapus registry Run ({})",
                dr.0
            );
            Ok(())
        };
        let _ = RegCloseKey(hkey);
        result
    }
}

fn find_tailscale_ui_exe() -> PathBuf {
    let pf = PathBuf::from(
        std::env::var_os("ProgramFiles").unwrap_or_else(|| "C:\\Program Files".into()),
    );
    let ipn = pf.join("Tailscale\\tailscale-ipn.exe");
    if ipn.exists() {
        return ipn;
    }
    let cli = pf.join("Tailscale\\tailscale.exe");
    if cli.exists() {
        return cli;
    }
    PathBuf::from("tailscale-ipn.exe")
}

unsafe fn draw_text(
    hdc: HDC,
    font: HFONT,
    color: COLORREF,
    text: &str,
    rect: RECT,
    flags: DRAW_TEXT_FORMAT,
) {
    unsafe {
        let old_font = SelectObject(hdc, font.into());
        SetTextColor(hdc, color);
        let mut wide: Vec<u16> = text.encode_utf16().collect();
        let mut r = rect;
        DrawTextW(hdc, &mut wide, &mut r, flags);
        SelectObject(hdc, old_font);
    }
}

unsafe fn draw_card(hdc: HDC, x: i32, y: i32, w: i32, h: i32, state: &UiState) {
    unsafe {
        let pen = CreatePen(PS_SOLID, 1, COL_CARD_BORDER);
        let old_pen = SelectObject(hdc, pen.into());
        let old_brush = SelectObject(hdc, state.brush_card.into());
        let _ = RoundRect(hdc, x, y, x + w, y + h, 14, 14);
        SelectObject(hdc, old_brush);
        SelectObject(hdc, old_pen);
        let _ = DeleteObject(pen.into());
    }
}

unsafe fn paint_window(hwnd: HWND, state: &UiState) {
    unsafe {
        let mut ps = PAINTSTRUCT::default();
        let screen_dc = BeginPaint(hwnd, &mut ps);
        let mut client = RECT::default();
        let _ = GetClientRect(hwnd, &mut client);
        let w = (client.right - client.left).max(1);
        let h = (client.bottom - client.top).max(1);
        let hdc = CreateCompatibleDC(Some(screen_dc));
        let mem_bmp = CreateCompatibleBitmap(screen_dc, w, h);
        let old_bmp = SelectObject(hdc, mem_bmp.into());

        FillRect(hdc, &client, state.brush_bg);
        SetBkMode(hdc, TRANSPARENT);

        // Header
        draw_text(
            hdc,
            state.font_title,
            COL_TEXT,
            "PC Remote v5.0 (Rust Native)",
            RECT {
                left: 22,
                top: 16,
                right: 420,
                bottom: 46,
            },
            DT_LEFT | DT_VCENTER | DT_SINGLELINE,
        );
        draw_text(
            hdc,
            state.font_small,
            COL_GREEN,
            &format!("\u{25CF} HTTPS :{} Aktif", state.ctx.port),
            RECT {
                left: 22,
                top: 46,
                right: 420,
                bottom: 66,
            },
            DT_LEFT | DT_VCENTER | DT_SINGLELINE,
        );

        // Left Card: QR Pairing
        draw_card(hdc, 20, 78, 360, 514, state);
        draw_text(
            hdc,
            state.font_heading,
            COL_TEXT,
            "Pairing QR Code",
            RECT {
                left: 40,
                top: 94,
                right: 360,
                bottom: 120,
            },
            DT_LEFT | DT_VCENTER | DT_SINGLELINE,
        );
        draw_text(
            hdc,
            state.font_small,
            COL_MUTED,
            "Pindai dari aplikasi PC Remote di HP Android",
            RECT {
                left: 40,
                top: 120,
                right: 360,
                bottom: 142,
            },
            DT_LEFT | DT_VCENTER | DT_SINGLELINE,
        );

        // QR Render box (248x248 with quiet zone)
        let qr_box = RECT {
            left: 76,
            top: 158,
            right: 324,
            bottom: 406,
        };
        if state.qr_locked {
            let pen = CreatePen(PS_SOLID, 1, COL_CARD_BORDER);
            let old_pen = SelectObject(hdc, pen.into());
            let old_brush = SelectObject(hdc, state.brush_input.into());
            let _ = RoundRect(
                hdc,
                qr_box.left,
                qr_box.top,
                qr_box.right,
                qr_box.bottom,
                12,
                12,
            );
            SelectObject(hdc, old_brush);
            SelectObject(hdc, old_pen);
            let _ = DeleteObject(pen.into());

            draw_text(
                hdc,
                state.font_heading,
                COL_GREEN,
                "Perangkat Terhubung\nQR Terkunci",
                RECT {
                    left: qr_box.left + 16,
                    top: qr_box.top + 64,
                    right: qr_box.right - 16,
                    bottom: qr_box.top + 132,
                },
                DT_CENTER | DT_WORDBREAK,
            );
            draw_text(
                hdc,
                state.font_small,
                COL_MUTED,
                "Cabut akses perangkat untuk membuat QR baru.",
                RECT {
                    left: qr_box.left + 20,
                    top: qr_box.top + 140,
                    right: qr_box.right - 20,
                    bottom: qr_box.bottom - 24,
                },
                DT_CENTER | DT_WORDBREAK,
            );
            draw_text(
                hdc,
                state.font_body,
                COL_GREEN,
                "\u{2713} Perangkat tepercaya aktif",
                RECT {
                    left: 40,
                    top: 424,
                    right: 360,
                    bottom: 452,
                },
                DT_CENTER | DT_VCENTER | DT_SINGLELINE,
            );
        } else {
            let white_brush = CreateSolidBrush(COL_WHITE);
            let black_brush = CreateSolidBrush(COL_BLACK);
            FillRect(hdc, &qr_box, white_brush);

            if state.qr_width > 0 && state.qr_modules.len() == state.qr_width * state.qr_width {
                let quiet = 16;
                let avail = (qr_box.right - qr_box.left) - quiet * 2;
                let cell = (avail / state.qr_width as i32).max(2);
                let total = cell * state.qr_width as i32;
                let ox = qr_box.left + ((qr_box.right - qr_box.left) - total) / 2;
                let oy = qr_box.top + ((qr_box.bottom - qr_box.top) - total) / 2;
                for r in 0..state.qr_width {
                    for c in 0..state.qr_width {
                        if state.qr_modules[r * state.qr_width + c] {
                            let mr = RECT {
                                left: ox + (c as i32) * cell,
                                top: oy + (r as i32) * cell,
                                right: ox + ((c + 1) as i32) * cell,
                                bottom: oy + ((r + 1) as i32) * cell,
                            };
                            FillRect(hdc, &mr, black_brush);
                        }
                    }
                }
            }
            let _ = DeleteObject(white_brush.into());
            let _ = DeleteObject(black_brush.into());

            let rem = state.qr_remaining_secs();
            let (countdown_text, countdown_col) = if rem > 0 {
                (format!("Berlaku: {rem} detik"), COL_ACCENT)
            } else {
                (
                    "QR kedaluwarsa \u{2014} buat QR baru".to_string(),
                    COL_AMBER,
                )
            };
            draw_text(
                hdc,
                state.font_body,
                countdown_col,
                &countdown_text,
                RECT {
                    left: 40,
                    top: 424,
                    right: 360,
                    bottom: 452,
                },
                DT_CENTER | DT_VCENTER | DT_SINGLELINE,
            );
        }
        draw_text(
            hdc,
            state.font_small,
            COL_MUTED,
            "Menutup jendela hanya menyembunyikan ke Tray.",
            RECT {
                left: 40,
                top: 462,
                right: 360,
                bottom: 486,
            },
            DT_CENTER | DT_VCENTER | DT_SINGLELINE,
        );

        // Right Top Card: Tailscale Wizard & Network Status
        draw_card(hdc, 396, 78, 420, 242, state);
        draw_text(
            hdc,
            state.font_heading,
            COL_TEXT,
            "Jaringan & Wizard Tailscale (3 Langkah)",
            RECT {
                left: 412,
                top: 92,
                right: 800,
                bottom: 118,
            },
            DT_LEFT | DT_VCENTER | DT_SINGLELINE,
        );

        let ep = state.ctx.endpoints.lock().unwrap().clone();
        let hosts = ep["hosts"]
            .as_array()
            .map(|arr| {
                arr.iter()
                    .filter_map(Value::as_str)
                    .collect::<Vec<_>>()
                    .join(", ")
            })
            .filter(|s| !s.is_empty())
            .unwrap_or_else(|| "-".into());
        let ts_url = ep["tailscale_url"].as_str().unwrap_or("");
        let ts_ip = ts_url
            .strip_prefix("https://")
            .and_then(|s| s.split(':').next())
            .filter(|s| !s.is_empty())
            .unwrap_or("-");
        let ts_dns = ep["tailscale_dns"]
            .as_str()
            .filter(|s| !s.is_empty())
            .unwrap_or("-");
        let ts_state = ep["tailscale_state"].as_str().unwrap_or("Unavailable");
        let ts_col = match ts_state {
            "Running" => COL_GREEN,
            "NeedsLogin" => COL_AMBER,
            "NotInstalled" | "Stopped" => COL_RED,
            _ => COL_MUTED,
        };

        draw_text(
            hdc,
            state.font_body,
            COL_TEXT,
            &format!("LAN IP: {hosts}"),
            RECT {
                left: 412,
                top: 124,
                right: 800,
                bottom: 146,
            },
            DT_LEFT | DT_VCENTER | DT_SINGLELINE | DT_END_ELLIPSIS,
        );
        draw_text(
            hdc,
            state.font_body,
            COL_TEXT,
            &format!("Tailscale IP: {ts_ip}"),
            RECT {
                left: 412,
                top: 148,
                right: 800,
                bottom: 170,
            },
            DT_LEFT | DT_VCENTER | DT_SINGLELINE | DT_END_ELLIPSIS,
        );
        draw_text(
            hdc,
            state.font_body,
            COL_MUTED,
            &format!("MagicDNS: {ts_dns}"),
            RECT {
                left: 412,
                top: 172,
                right: 800,
                bottom: 194,
            },
            DT_LEFT | DT_VCENTER | DT_SINGLELINE | DT_END_ELLIPSIS,
        );
        draw_text(
            hdc,
            state.font_body,
            ts_col,
            &format!("Status Tailscale: {ts_state}"),
            RECT {
                left: 412,
                top: 196,
                right: 800,
                bottom: 220,
            },
            DT_LEFT | DT_VCENTER | DT_SINGLELINE,
        );

        let (step3_text, step3_col) = if ts_state == "Running" {
            ("\u{2713} 3. Siap Luar Rumah (Tailscale Aktif)", COL_GREEN)
        } else {
            (
                "\u{25CB} 3. Belum Siap Luar Rumah (Selesaikan Langkah 1\u{2013}2)",
                COL_MUTED,
            )
        };
        draw_text(
            hdc,
            state.font_body,
            step3_col,
            step3_text,
            RECT {
                left: 412,
                top: 292,
                right: 800,
                bottom: 314,
            },
            DT_LEFT | DT_VCENTER | DT_SINGLELINE,
        );

        // Right Bottom Card: Devices, PIN, Autostart
        draw_card(hdc, 396, 334, 420, 258, state);
        draw_text(
            hdc,
            state.font_heading,
            COL_TEXT,
            "Perangkat Tepercaya, PIN & Autostart",
            RECT {
                left: 412,
                top: 348,
                right: 800,
                bottom: 374,
            },
            DT_LEFT | DT_VCENTER | DT_SINGLELINE,
        );
        draw_text(
            hdc,
            state.font_body,
            COL_TEXT,
            "PIN (4\u{2013}8 digit):",
            RECT {
                left: 412,
                top: 484,
                right: 490,
                bottom: 512,
            },
            DT_LEFT | DT_VCENTER | DT_SINGLELINE,
        );

        if let Some(until) = state.status_until
            && Instant::now() < until
            && !state.status_message.is_empty()
        {
            draw_text(
                hdc,
                state.font_small,
                if state.status_is_error {
                    COL_RED
                } else {
                    COL_GREEN
                },
                &state.status_message,
                RECT {
                    left: 412,
                    top: 564,
                    right: 800,
                    bottom: 586,
                },
                DT_LEFT | DT_VCENTER | DT_SINGLELINE | DT_END_ELLIPSIS,
            );
        }

        let _ = BitBlt(screen_dc, 0, 0, w, h, Some(hdc), 0, 0, SRCCOPY);
        SelectObject(hdc, old_bmp);
        let _ = DeleteObject(mem_bmp.into());
        let _ = DeleteDC(hdc);
        let _ = EndPaint(hwnd, &ps);
    }
}

unsafe fn draw_owner_button(dis: &DRAWITEMSTRUCT, state: &UiState) {
    unsafe {
        let pressed = (dis.itemState.0 & ODS_SELECTED.0) != 0;
        let id = dis.CtlID as i32;
        let bg_color = match id {
            ID_BTN_STOP | ID_BTN_REVOKE => {
                if pressed {
                    COL_BTN_DANGER_PRESS
                } else {
                    COL_BTN_DANGER
                }
            }
            ID_BTN_NEW_QR | ID_BTN_SAVE_PIN | ID_BTN_TS_LOGIN => {
                if pressed {
                    COL_BTN_PRIMARY_PRESS
                } else {
                    COL_BTN_PRIMARY
                }
            }
            _ => {
                if pressed {
                    COL_BTN_SLATE_PRESS
                } else {
                    COL_BTN_SLATE
                }
            }
        };
        let brush = CreateSolidBrush(bg_color);
        let pen = CreatePen(PS_SOLID, 1, bg_color);
        let old_pen = SelectObject(dis.hDC, pen.into());
        let old_brush = SelectObject(dis.hDC, brush.into());
        let _ = RoundRect(
            dis.hDC,
            dis.rcItem.left,
            dis.rcItem.top,
            dis.rcItem.right,
            dis.rcItem.bottom,
            8,
            8,
        );
        SelectObject(dis.hDC, old_brush);
        SelectObject(dis.hDC, old_pen);
        let _ = DeleteObject(brush.into());
        let _ = DeleteObject(pen.into());

        let mut buf = [0u16; 128];
        let len = GetWindowTextW(dis.hwndItem, &mut buf) as usize;
        let text = String::from_utf16_lossy(&buf[..len]);
        SetBkMode(dis.hDC, TRANSPARENT);
        draw_text(
            dis.hDC,
            state.font_body,
            COL_TEXT,
            &text,
            dis.rcItem,
            DT_CENTER | DT_VCENTER | DT_SINGLELINE,
        );
    }
}

unsafe extern "system" fn window_proc(
    hwnd: HWND,
    message: u32,
    wparam: WPARAM,
    lparam: LPARAM,
) -> LRESULT {
    unsafe {
        if message == WM_NCCREATE {
            let cs = &*(lparam.0 as *const CREATESTRUCTW);
            SetWindowLongPtrW(hwnd, GWLP_USERDATA, cs.lpCreateParams as isize);
            return DefWindowProcW(hwnd, message, wparam, lparam);
        }

        let ptr = GetWindowLongPtrW(hwnd, GWLP_USERDATA) as *mut UiState;
        if ptr.is_null() {
            return DefWindowProcW(hwnd, message, wparam, lparam);
        }
        let state = &mut *ptr;

        match message {
            WM_CREATE => {
                let cs = &*(lparam.0 as *const CREATESTRUCTW);
                let _ = create_controls(hwnd, cs.hInstance, state);
                LRESULT(0)
            }
            WM_ERASEBKGND => LRESULT(1),
            WM_PAINT => {
                paint_window(hwnd, state);
                LRESULT(0)
            }
            WM_DRAWITEM => {
                let dis = &*(lparam.0 as *const DRAWITEMSTRUCT);
                draw_owner_button(dis, state);
                LRESULT(1)
            }
            WM_CTLCOLORSTATIC | WM_CTLCOLORBTN => {
                let hdc = HDC(wparam.0 as *mut _);
                SetTextColor(hdc, COL_TEXT);
                SetBkColor(hdc, COL_CARD);
                LRESULT(state.brush_card.0 as isize)
            }
            WM_CTLCOLOREDIT | WM_CTLCOLORLISTBOX => {
                let hdc = HDC(wparam.0 as *mut _);
                SetTextColor(hdc, COL_TEXT);
                SetBkColor(hdc, COL_INPUT_BG);
                LRESULT(state.brush_input.0 as isize)
            }
            WM_TIMER => {
                match wparam.0 {
                    TIMER_COUNTDOWN_ID => {
                        state.refresh_devices_listbox();
                        state.sync_qr_lock_state(hwnd);
                        if !state.qr_locked {
                            let countdown_rect = RECT {
                                left: 40,
                                top: 420,
                                right: 360,
                                bottom: 456,
                            };
                            let _ = InvalidateRect(Some(hwnd), Some(&countdown_rect), false);
                        }
                    }
                    TIMER_POLL_ID => {
                        state.ctx.refresh_notify.notify_one();
                        state.sync_install_button();
                        state.refresh_devices_listbox();
                        state.sync_qr_lock_state(hwnd);
                        let _ = InvalidateRect(Some(hwnd), None, false);
                    }
                    _ => {}
                }
                LRESULT(0)
            }
            WM_COMMAND => {
                let id = (wparam.0 & 0xFFFF) as i32;
                match id {
                    ID_BTN_STOP => {
                        state
                            .ctx
                            .shutdown
                            .graceful_shutdown(Some(Duration::from_secs(2)));
                        let _ = DestroyWindow(hwnd);
                    }
                    ID_BTN_NEW_QR => {
                        if state.has_active_device() {
                            state.sync_qr_lock_state(hwnd);
                            state.set_feedback(
                                "QR terkunci karena perangkat sudah terhubung.",
                                true,
                            );
                        } else {
                            state.generate_qr();
                            state
                                .set_feedback("QR pairing baru dibuat (berlaku 120 detik).", false);
                        }
                        let _ = InvalidateRect(Some(hwnd), None, false);
                    }
                    ID_BTN_TS_INSTALL => {
                        ShellExecuteW(
                            None,
                            w!("open"),
                            TAILSCALE_DOWNLOAD_URL,
                            None,
                            None,
                            SW_SHOWNORMAL,
                        );
                    }
                    ID_BTN_TS_LOGIN => {
                        let exe = find_tailscale_ui_exe();
                        let wide = HSTRING::from(exe.to_string_lossy().as_ref());
                        ShellExecuteW(None, w!("open"), &wide, None, None, SW_SHOWNORMAL);
                        state.ctx.refresh_notify.notify_one();
                    }
                    ID_BTN_TS_MOBILE => {
                        ShellExecuteW(
                            None,
                            w!("open"),
                            TAILSCALE_ANDROID_URL,
                            None,
                            None,
                            SW_SHOWNORMAL,
                        );
                    }
                    ID_BTN_REVOKE => {
                        let sel = SendMessageW(
                            state.hwnd_devices,
                            LB_GETCURSEL,
                            Some(WPARAM(0)),
                            Some(LPARAM(0)),
                        )
                        .0;
                        let idx = if sel >= 0 {
                            sel as usize
                        } else {
                            state.device_ids.len().saturating_sub(1)
                        };
                        if let Some(dev_id) = state.device_ids.get(idx).cloned() {
                            let res = state.ctx.auth.lock().unwrap().revoke(&dev_id);
                            match res {
                                Ok(()) => state.set_feedback("Akses perangkat dicabut.", false),
                                Err(e) => state.set_feedback(format!("Gagal: {e}"), true),
                            }
                            state.refresh_devices_listbox();
                            state.sync_qr_lock_state(hwnd);
                            let _ = InvalidateRect(Some(hwnd), None, false);
                        } else {
                            state.set_feedback("Tidak ada perangkat untuk dicabut.", true);
                            let _ = InvalidateRect(Some(hwnd), None, false);
                        }
                    }
                    ID_BTN_SAVE_PIN => {
                        let mut buf = [0u16; 32];
                        let len = GetWindowTextW(state.hwnd_pin, &mut buf) as usize;
                        let pin = String::from_utf16_lossy(&buf[..len]);
                        let res = state.ctx.auth.lock().unwrap().set_pin(pin.trim());
                        match res {
                            Ok(()) => {
                                let _ = SetWindowTextW(state.hwnd_pin, w!(""));
                                state.set_feedback("PIN berhasil disimpan.", false);
                            }
                            Err(e) => {
                                state.set_feedback(format!("PIN tidak valid: {e}"), true);
                            }
                        }
                        let _ = InvalidateRect(Some(hwnd), None, false);
                    }
                    ID_CHK_AUTOSTART => {
                        let checked = SendMessageW(
                            state.hwnd_autostart,
                            BM_GETCHECK,
                            Some(WPARAM(0)),
                            Some(LPARAM(0)),
                        )
                        .0 as usize
                            == BST_CHECKED_VAL;
                        match set_autostart(checked) {
                            Ok(()) => state.set_feedback(
                                if checked {
                                    "Autostart Windows diaktifkan."
                                } else {
                                    "Autostart Windows dinonaktifkan."
                                },
                                false,
                            ),
                            Err(e) => state.set_feedback(format!("Autostart gagal: {e}"), true),
                        }
                        let _ = InvalidateRect(Some(hwnd), None, false);
                    }
                    _ => {}
                }
                LRESULT(0)
            }
            WM_APP_SHOW_FOCUS => {
                state.refresh_devices_listbox();
                state.sync_install_button();
                state.sync_qr_lock_state(hwnd);
                state.ctx.refresh_notify.notify_one();
                let _ = ShowWindow(hwnd, SW_RESTORE);
                let _ = ShowWindow(hwnd, SW_SHOW);
                let _ = SetForegroundWindow(hwnd);
                let _ = SetTimer(Some(hwnd), TIMER_COUNTDOWN_ID, 1000, None);
                let _ = SetTimer(Some(hwnd), TIMER_POLL_ID, 2000, None);
                let _ = InvalidateRect(Some(hwnd), None, false);
                LRESULT(0)
            }
            WM_APP_QUIT => {
                let _ = DestroyWindow(hwnd);
                LRESULT(0)
            }
            WM_CLOSE => {
                let _ = KillTimer(Some(hwnd), TIMER_COUNTDOWN_ID);
                let _ = KillTimer(Some(hwnd), TIMER_POLL_ID);
                let _ = ShowWindow(hwnd, SW_HIDE);
                LRESULT(0)
            }
            WM_DESTROY => {
                let _ = KillTimer(Some(hwnd), TIMER_COUNTDOWN_ID);
                let _ = KillTimer(Some(hwnd), TIMER_POLL_ID);
                PostQuitMessage(0);
                let _ = Box::from_raw(ptr);
                SetWindowLongPtrW(hwnd, GWLP_USERDATA, 0);
                LRESULT(0)
            }
            _ => DefWindowProcW(hwnd, message, wparam, lparam),
        }
    }
}
