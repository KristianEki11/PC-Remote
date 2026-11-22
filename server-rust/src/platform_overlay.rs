//! Overlay owns its HWND and message pump on one thread. No global cursor changes.
use anyhow::{Result, ensure};
use std::sync::{
    OnceLock,
    atomic::{AtomicU8, Ordering},
};
use windows::{
    Win32::{
        Foundation::{COLORREF, HWND, LPARAM, LRESULT, WPARAM},
        Graphics::Gdi::{BLACK_BRUSH, GetStockObject, HBRUSH},
        System::{
            LibraryLoader::GetModuleHandleW,
            Power::{ES_CONTINUOUS, ES_SYSTEM_REQUIRED, SetThreadExecutionState},
        },
        UI::WindowsAndMessaging::*,
    },
    core::w,
};
static WINDOW: OnceLock<Result<usize, String>> = OnceLock::new();
static ALPHA: AtomicU8 = AtomicU8::new(0);
const UPDATE: u32 = WM_APP + 1;
pub fn alpha() -> u8 {
    ALPHA.load(Ordering::Relaxed)
}
pub fn set(alpha: u8) -> Result<()> {
    let handle = WINDOW.get_or_init(|| {
        let (tx, rx) = std::sync::mpsc::sync_channel(1);
        std::thread::spawn(move || unsafe {
            let result = (|| -> Result<HWND> {
                let instance = GetModuleHandleW(None)?;
                let class = WNDCLASSW {
                    lpfnWndProc: Some(window_proc),
                    hInstance: instance.into(),
                    lpszClassName: w!("PCRemoteDimOverlay"),
                    hbrBackground: HBRUSH(GetStockObject(BLACK_BRUSH).0),
                    ..Default::default()
                };
                ensure!(
                    RegisterClassW(&class) != 0,
                    "cannot register overlay window"
                );
                Ok(CreateWindowExW(
                    WS_EX_LAYERED | WS_EX_TOPMOST | WS_EX_TOOLWINDOW | WS_EX_TRANSPARENT,
                    w!("PCRemoteDimOverlay"),
                    w!("PC Remote dimming"),
                    WS_POPUP,
                    GetSystemMetrics(SM_XVIRTUALSCREEN),
                    GetSystemMetrics(SM_YVIRTUALSCREEN),
                    GetSystemMetrics(SM_CXVIRTUALSCREEN),
                    GetSystemMetrics(SM_CYVIRTUALSCREEN),
                    None,
                    None,
                    Some(instance.into()),
                    None,
                )?)
            })();
            match result {
                Ok(hwnd) => {
                    let _ = tx.send(Ok(hwnd.0 as usize));
                    let mut message = MSG::default();
                    while GetMessageW(&mut message, None, 0, 0).0 > 0 {
                        let _ = TranslateMessage(&message);
                        DispatchMessageW(&message);
                    }
                }
                Err(error) => {
                    let _ = tx.send(Err(error.to_string()));
                }
            }
        });
        rx.recv()
            .unwrap_or_else(|_| Err("overlay thread stopped".into()))
    });
    let raw = handle
        .as_ref()
        .map_err(|error| anyhow::anyhow!(error.clone()))?;
    let mut applied = 0usize;
    let result = unsafe {
        SendMessageTimeoutW(
            HWND(*raw as *mut _),
            UPDATE,
            WPARAM(alpha as usize),
            LPARAM(0),
            SMTO_ABORTIFHUNG,
            2000,
            Some(&mut applied),
        )
    };
    ensure!(result.0 != 0, "overlay did not respond");
    ensure!(applied != 0, "Windows rejected overlay opacity update");
    Ok(())
}
unsafe extern "system" fn window_proc(
    hwnd: HWND,
    message: u32,
    wparam: WPARAM,
    lparam: LPARAM,
) -> LRESULT {
    unsafe {
        match message {
            UPDATE => {
                let alpha = wparam.0 as u8;
                let mut style = WS_EX_LAYERED | WS_EX_TOPMOST | WS_EX_TOOLWINDOW;
                if alpha < 255 {
                    style |= WS_EX_TRANSPARENT | WS_EX_NOACTIVATE;
                }
                SetWindowLongPtrW(hwnd, GWL_EXSTYLE, style.0 as isize);
                if SetLayeredWindowAttributes(hwnd, COLORREF(0), alpha, LWA_ALPHA).is_err() {
                    return LRESULT(0);
                }
                ALPHA.store(alpha, Ordering::Relaxed);
                if alpha == 0 {
                    let _ = ShowWindow(hwnd, SW_HIDE);
                    SetThreadExecutionState(ES_CONTINUOUS);
                } else {
                    let _ = SetWindowPos(
                        hwnd,
                        Some(HWND_TOPMOST),
                        GetSystemMetrics(SM_XVIRTUALSCREEN),
                        GetSystemMetrics(SM_YVIRTUALSCREEN),
                        GetSystemMetrics(SM_CXVIRTUALSCREEN),
                        GetSystemMetrics(SM_CYVIRTUALSCREEN),
                        SWP_SHOWWINDOW | SWP_NOACTIVATE,
                    );
                    if alpha == 255 {
                        SetThreadExecutionState(ES_CONTINUOUS | ES_SYSTEM_REQUIRED);
                        let _ = SetForegroundWindow(hwnd);
                    } else {
                        SetThreadExecutionState(ES_CONTINUOUS);
                    }
                }
                LRESULT(1)
            }
            WM_LBUTTONDOWN | WM_RBUTTONDOWN | WM_KEYDOWN | WM_CLOSE => {
                ALPHA.store(0, Ordering::Relaxed);
                let _ = ShowWindow(hwnd, SW_HIDE);
                SetThreadExecutionState(ES_CONTINUOUS);
                LRESULT(0)
            }
            _ => DefWindowProcW(hwnd, message, wparam, lparam),
        }
    }
}
