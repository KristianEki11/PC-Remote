use anyhow::{Context, Result, bail};
use rand::RngCore;
use serde::{Deserialize, Serialize};
use sha2::{Digest, Sha256};
use std::{
    collections::HashMap,
    path::{Path, PathBuf},
    time::{SystemTime, UNIX_EPOCH},
};
use subtle::ConstantTimeEq;

pub fn now() -> u64 {
    SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .unwrap_or_default()
        .as_secs()
}
pub fn random_token() -> String {
    let mut b = [0; 32];
    rand::rngs::OsRng.fill_bytes(&mut b);
    hex::encode(b)
}
pub fn digest(s: &str) -> String {
    hex::encode(Sha256::digest(s.as_bytes()))
}

#[derive(Clone, Serialize, Deserialize)]
pub struct Device {
    pub id: String,
    pub name: String,
    pub token_hash: String,
    pub expires_at: u64,
}
#[derive(Serialize, Deserialize)]
pub struct Store {
    pub server_id: String,
    pub key_pem: String,
    pub cert_pem: String,
    pub fingerprint: String,
    pub pin_hash: Option<String>,
    pub devices: Vec<Device>,
}
pub struct Auth {
    pub store: Store,
    path: PathBuf,
    pair: Option<(String, u64)>,
    failures: HashMap<String, (u32, u64)>,
}

// DPAPI binds TLS key and device records to this Windows account.
fn protect(data: &[u8], decrypt: bool) -> Result<Vec<u8>> {
    use windows::Win32::{
        Foundation::{HLOCAL, LocalFree},
        Security::Cryptography::{
            CRYPT_INTEGER_BLOB, CRYPTPROTECT_UI_FORBIDDEN, CryptProtectData, CryptUnprotectData,
        },
    };
    let input = CRYPT_INTEGER_BLOB {
        cbData: data.len().try_into()?,
        pbData: data.as_ptr() as *mut u8,
    };
    let mut output = CRYPT_INTEGER_BLOB::default();
    unsafe {
        if decrypt {
            CryptUnprotectData(
                &input,
                None,
                None,
                None,
                None,
                CRYPTPROTECT_UI_FORBIDDEN,
                &mut output,
            )?;
        } else {
            CryptProtectData(
                &input,
                None,
                None,
                None,
                None,
                CRYPTPROTECT_UI_FORBIDDEN,
                &mut output,
            )?;
        }
        let bytes = std::slice::from_raw_parts(output.pbData, output.cbData as usize).to_vec();
        let _ = LocalFree(Some(HLOCAL(output.pbData.cast())));
        Ok(bytes)
    }
}
impl Auth {
    pub fn load(dir: &Path) -> Result<Self> {
        std::fs::create_dir_all(dir)?;
        let path = dir.join("identity.dpapi");
        let store = if path.exists() {
            serde_json::from_slice(&protect(&std::fs::read(&path)?, true)?).context(
                "Cannot load protected identity; do not delete it unless re-pairing all devices",
            )?
        } else {
            let cert = rcgen::generate_simple_self_signed(vec!["localhost".into()])?;
            Store {
                server_id: random_token(),
                fingerprint: hex::encode(Sha256::digest(cert.cert.der())),
                cert_pem: cert.cert.pem(),
                key_pem: cert.key_pair.serialize_pem(),
                pin_hash: None,
                devices: vec![],
            }
        };
        let a = Self {
            store,
            path,
            pair: None,
            failures: HashMap::new(),
        };
        a.save()?;
        Ok(a)
    }
    pub fn save(&self) -> Result<()> {
        let temp = self.path.with_extension("tmp");
        std::fs::write(&temp, protect(&serde_json::to_vec(&self.store)?, false)?)?;
        // Windows rename replaces an existing file atomically via MoveFileExW.
        use std::os::windows::ffi::OsStrExt;
        use windows_sys::Win32::Storage::FileSystem::{
            MOVEFILE_REPLACE_EXISTING, MOVEFILE_WRITE_THROUGH, MoveFileExW,
        };
        let from: Vec<u16> = temp.as_os_str().encode_wide().chain(Some(0)).collect();
        let to: Vec<u16> = self.path.as_os_str().encode_wide().chain(Some(0)).collect();
        if unsafe {
            MoveFileExW(
                from.as_ptr(),
                to.as_ptr(),
                MOVEFILE_REPLACE_EXISTING | MOVEFILE_WRITE_THROUGH,
            )
        } == 0
        {
            return Err(std::io::Error::last_os_error().into());
        }
        Ok(())
    }
    pub fn new_pair(&mut self) -> String {
        let token = random_token();
        self.pair = Some((digest(&token), now() + 120));
        token
    }
    pub fn authenticate(&self, token: &str) -> Option<Device> {
        let hash = digest(token);
        self.store
            .devices
            .iter()
            .find(|d| {
                d.expires_at > now() && bool::from(d.token_hash.as_bytes().ct_eq(hash.as_bytes()))
            })
            .cloned()
    }
    pub fn locked(&mut self, ip: &str) -> bool {
        self.failures.retain(|_, (_, until)| *until > now());
        match self.failures.get(ip) {
            Some((attempts, _)) => *attempts >= 5,
            // Fail closed for new sources when the table is full. Never clear
            // existing lockouts merely because another source appeared.
            None => self.failures.len() >= 4096,
        }
    }
    fn failed(&mut self, ip: &str) {
        if self.failures.len() >= 4096 && !self.failures.contains_key(ip) {
            return;
        }
        let v = self.failures.entry(ip.into()).or_insert((0, now() + 900));
        v.0 = v.0.saturating_add(1);
    }
    pub fn pair(&mut self, token: &str, name: &str, ip: &str) -> Result<(String, Device)> {
        if self.locked(ip) {
            bail!("too many authentication attempts");
        }
        let hash = digest(token);
        let valid = self
            .pair
            .as_ref()
            .is_some_and(|(h, t)| *t > now() && bool::from(h.as_bytes().ct_eq(hash.as_bytes())));
        if !valid {
            self.failed(ip);
            bail!("invalid or expired pairing token");
        }
        let result = self.issue(name)?;
        self.pair = None;
        self.failures.remove(ip);
        Ok(result)
    }
    pub fn login(&mut self, pin: &str, name: &str, ip: &str) -> Result<(String, Device)> {
        if self.locked(ip) {
            bail!("too many authentication attempts");
        }
        if !self
            .store
            .pin_hash
            .as_ref()
            .is_some_and(|h| bcrypt::verify(pin, h).unwrap_or(false))
        {
            self.failed(ip);
            bail!("invalid PIN; pair using QR first");
        }
        let result = self.issue(name)?;
        self.failures.remove(ip);
        Ok(result)
    }
    fn issue(&mut self, name: &str) -> Result<(String, Device)> {
        if self
            .store
            .devices
            .iter()
            .filter(|d| d.expires_at > now())
            .count()
            >= 64
        {
            bail!("device limit reached; revoke an unused device");
        }
        let old = self.store.devices.clone();
        self.store.devices.retain(|d| d.expires_at > now());
        let token = random_token();
        let d = Device {
            id: random_token(),
            name: name.chars().take(80).collect(),
            token_hash: digest(&token),
            expires_at: now() + 365 * 86400,
        };
        self.store.devices.push(d.clone());
        if let Err(e) = self.save() {
            self.store.devices = old;
            return Err(e);
        }
        Ok((token, d))
    }
    pub fn revoke(&mut self, id: &str) -> Result<()> {
        let old = self.store.devices.clone();
        self.store.devices.retain(|d| d.id != id);
        if let Err(e) = self.save() {
            self.store.devices = old;
            return Err(e);
        }
        Ok(())
    }
    pub fn set_pin(&mut self, pin: &str) -> Result<()> {
        if !(4..=8).contains(&pin.len()) || !pin.bytes().all(|b| b.is_ascii_digit()) {
            bail!("PIN must have 4–8 digits")
        }
        let old = self.store.pin_hash.clone();
        self.store.pin_hash = Some(bcrypt::hash(pin, 12)?);
        if let Err(e) = self.save() {
            self.store.pin_hash = old;
            return Err(e);
        }
        Ok(())
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    #[test]
    fn pairing_is_single_use_and_persistent() {
        let dir = tempfile::tempdir().unwrap();
        let mut a = Auth::load(dir.path()).unwrap();
        let p = a.new_pair();
        let (t, d) = a.pair(&p, "phone", "127.0.0.1").unwrap();
        assert!(a.pair(&p, "other", "127.0.0.1").is_err());
        let mut b = Auth::load(dir.path()).unwrap();
        assert_eq!(b.authenticate(&t).unwrap().id, d.id);
        b.revoke(&d.id).unwrap();
        assert!(b.authenticate(&t).is_none());
    }
    #[test]
    fn failures_lock_only_source() {
        let dir = tempfile::tempdir().unwrap();
        let mut a = Auth::load(dir.path()).unwrap();
        for _ in 0..5 {
            assert!(a.pair("wrong", "x", "one").is_err());
        }
        assert!(a.locked("one"));
        assert!(!a.locked("two"));
    }
    #[test]
    fn full_failure_table_preserves_lockouts() {
        let dir = tempfile::tempdir().unwrap();
        let mut a = Auth::load(dir.path()).unwrap();
        for i in 0..4096 {
            a.failures.insert(i.to_string(), (5, now() + 900));
        }
        assert!(a.locked("new-source"));
        assert!(a.pair("bad", "phone", "new-source").is_err());
        assert!(a.locked("0"));
        assert_eq!(a.failures.len(), 4096);
        a.failures.get_mut("0").unwrap().1 = now();
        assert!(!a.locked("new-source"));
    }
    #[test]
    fn expired_records_do_not_exhaust_device_capacity() {
        let dir = tempfile::tempdir().unwrap();
        let mut a = Auth::load(dir.path()).unwrap();
        a.store.devices = (0..64)
            .map(|i| Device {
                id: i.to_string(),
                name: "old".into(),
                token_hash: digest("old"),
                expires_at: now(),
            })
            .collect();
        let qr = a.new_pair();
        assert!(a.pair(&qr, "new", "one").is_ok());
        assert_eq!(a.store.devices.len(), 1);
    }
    #[test]
    fn failed_persistence_rolls_back_issue_revoke_and_pin() {
        let dir = tempfile::tempdir().unwrap();
        let mut a = Auth::load(dir.path()).unwrap();
        let qr = a.new_pair();
        let (token, device) = a.pair(&qr, "existing", "one").unwrap();
        let original_path = a.path.clone();
        // Deliberately target an existing directory: atomic replacement must fail.
        a.path = dir.path().join("blocked");
        std::fs::create_dir(&a.path).unwrap();
        let new_qr = a.new_pair();
        assert!(a.pair(&new_qr, "new", "two").is_err());
        assert_eq!(a.store.devices.len(), 1);
        assert!(a.authenticate(&token).is_some());
        assert!(a.revoke(&device.id).is_err());
        assert!(a.authenticate(&token).is_some());
        assert!(a.set_pin("123456").is_err());
        assert!(a.store.pin_hash.is_none());
        a.path = original_path;
        // A storage failure must not consume the pairing token.
        assert!(a.pair(&new_qr, "new", "two").is_ok());
    }
}
