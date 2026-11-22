fn main() {
    println!("cargo:rerun-if-changed=assets/pc_remote.ico");
    if std::env::var("CARGO_CFG_TARGET_OS").as_deref() == Ok("windows") {
        let mut res = winresource::WindowsResource::new();
        res.set_icon_with_id("assets/pc_remote.ico", "1");
        res.compile().expect("failed to compile Windows icon resource");
    }
}
