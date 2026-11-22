#![cfg_attr(not(debug_assertions), windows_subsystem = "windows")]
mod auth;
mod gui;
mod platform;
#[cfg(test)]
mod review_tests;
mod transport;
mod tray;

#[tokio::main(worker_threads = 2)]
async fn main() -> anyhow::Result<()> {
    rustls::crypto::ring::default_provider()
        .install_default()
        .ok();
    let result = transport::run().await;
    if let Err(error) = &result {
        let dir = std::env::var_os("PCREMOTE_DATA_DIR")
            .map(std::path::PathBuf::from)
            .unwrap_or_else(|| {
                std::path::PathBuf::from(std::env::var_os("LOCALAPPDATA").unwrap_or_default())
                    .join("PCRemoteRust/data")
            });
        let _ = std::fs::create_dir_all(&dir);
        let _ = std::fs::write(dir.join("startup-error.log"), format!("{error:#}"));
        if !std::env::args().any(|a| a == "--background") {
            use windows::{
                Win32::UI::WindowsAndMessaging::{MB_ICONERROR, MB_OK, MessageBoxW},
                core::{HSTRING, w},
            };
            unsafe {
                let _ = MessageBoxW(
                    None,
                    &HSTRING::from(format!("PC Remote could not start: {error:#}")),
                    w!("PC Remote"),
                    MB_OK | MB_ICONERROR,
                );
            }
        }
    }
    result
}
