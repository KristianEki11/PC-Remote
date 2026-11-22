//! Per-user notification icon. The hidden window owns the icon and menu lifetime.
use std::cell::{Cell, RefCell};
use windows::{
    Win32::{
        Foundation::{HINSTANCE, HWND, LPARAM, LRESULT, POINT, WPARAM},
        System::LibraryLoader::GetModuleHandleW,
        UI::{Shell::*, WindowsAndMessaging::*},
    },
    core::{PCWSTR, w},
};

const TRAY_MESSAGE: u32 = WM_APP + 8;
thread_local! {
    static TASKBAR_CREATED: Cell<u32> = const { Cell::new(0) };
    static EXIT: RefCell<Option<Box<dyn Fn()>>> = const { RefCell::new(None) };
}

pub fn start(_port: u16, on_exit: impl Fn() + Send + 'static) {
    std::thread::spawn(move || {
        EXIT.with(|slot| *slot.borrow_mut() = Some(Box::new(on_exit)));
        if let Err(error) = run() {
            eprintln!("Tray unavailable: {error:#}");
        }
    });
}

fn run() -> anyhow::Result<()> {
    unsafe {
        let instance = GetModuleHandleW(None)?;
        let class = WNDCLASSW {
            lpfnWndProc: Some(window_proc),
            hInstance: instance.into(),
            lpszClassName: w!("PCRemoteTray"),
            ..Default::default()
        };
        anyhow::ensure!(RegisterClassW(&class) != 0, "cannot register tray window");
        TASKBAR_CREATED.set(RegisterWindowMessageW(w!("TaskbarCreated")));
        let window = CreateWindowExW(
            WINDOW_EX_STYLE::default(),
            w!("PCRemoteTray"),
            w!("PC Remote"),
            WINDOW_STYLE::default(),
            0,
            0,
            0,
            0,
            None,
            None,
            Some(instance.into()),
            None,
        )?;
        add_icon(window)?;
        let mut message = MSG::default();
        while GetMessageW(&mut message, None, 0, 0).0 > 0 {
            let _ = TranslateMessage(&message);
            DispatchMessageW(&message);
        }
    }
    Ok(())
}

unsafe fn add_icon(window: HWND) -> anyhow::Result<()> {
    unsafe {
        let instance: HINSTANCE = GetModuleHandleW(None)?.into();
        let tray_icon = LoadImageW(
            Some(instance),
            PCWSTR(1 as *const u16),
            IMAGE_ICON,
            GetSystemMetrics(SM_CXSMICON),
            GetSystemMetrics(SM_CYSMICON),
            LR_SHARED,
        )
        .ok()
        .map(|h| HICON(h.0))
        .unwrap_or(LoadIconW(None, IDI_APPLICATION)?);
        let mut icon = NOTIFYICONDATAW {
            cbSize: std::mem::size_of::<NOTIFYICONDATAW>() as u32,
            hWnd: window,
            uID: 1,
            uFlags: NIF_ICON | NIF_MESSAGE | NIF_TIP,
            uCallbackMessage: TRAY_MESSAGE,
            hIcon: tray_icon,
            ..Default::default()
        };
        for (target, character) in icon
            .szTip
            .iter_mut()
            .zip("PC Remote — LAN / Tailscale".encode_utf16())
        {
            *target = character;
        }
        anyhow::ensure!(
            Shell_NotifyIconW(NIM_ADD, &icon).as_bool(),
            "cannot add tray icon"
        );
    }
    Ok(())
}

fn open_native_gui() {
    crate::gui::show_or_focus();
}

unsafe extern "system" fn window_proc(
    window: HWND,
    message: u32,
    wparam: WPARAM,
    lparam: LPARAM,
) -> LRESULT {
    unsafe {
        if TASKBAR_CREATED.get() != 0 && message == TASKBAR_CREATED.get() {
            let _ = add_icon(window);
            return LRESULT(0);
        }
        match message {
            TRAY_MESSAGE => {
                if lparam.0 as u32 == WM_LBUTTONDBLCLK {
                    open_native_gui();
                }
                if lparam.0 as u32 == WM_RBUTTONUP
                    && let Ok(menu) = CreatePopupMenu()
                {
                    let _ = AppendMenuW(menu, MF_STRING, 1, w!("Open PC Remote"));
                    let _ = AppendMenuW(menu, MF_SEPARATOR, 0, None);
                    let _ = AppendMenuW(menu, MF_STRING, 2, w!("Stop server"));
                    let mut cursor = POINT::default();
                    let _ = GetCursorPos(&mut cursor);
                    let _ = SetForegroundWindow(window);
                    let chosen = TrackPopupMenu(
                        menu,
                        TPM_RETURNCMD | TPM_RIGHTBUTTON,
                        cursor.x,
                        cursor.y,
                        None,
                        window,
                        None,
                    )
                    .0;
                    let _ = DestroyMenu(menu);
                    match chosen {
                        1 => open_native_gui(),
                        2 => {
                            EXIT.with(|slot| {
                                if let Some(callback) = slot.borrow().as_ref() {
                                    callback();
                                }
                            });
                            let _ = DestroyWindow(window);
                        }
                        _ => (),
                    }
                }
                LRESULT(0)
            }
            WM_DESTROY => {
                let icon = NOTIFYICONDATAW {
                    cbSize: std::mem::size_of::<NOTIFYICONDATAW>() as u32,
                    hWnd: window,
                    uID: 1,
                    ..Default::default()
                };
                let _ = Shell_NotifyIconW(NIM_DELETE, &icon);
                PostQuitMessage(0);
                LRESULT(0)
            }
            _ => DefWindowProcW(window, message, wparam, lparam),
        }
    }
}
