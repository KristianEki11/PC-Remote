//! Regression checks for identity isolation and authentication persistence.
use crate::auth::{Auth, now};

#[test]
fn independent_users_do_not_share_pairing_credentials() {
    let first = tempfile::tempdir().unwrap();
    let second = tempfile::tempdir().unwrap();
    let mut a = Auth::load(first.path()).unwrap();
    let b = Auth::load(second.path()).unwrap();
    let pair = a.new_pair();
    let (token, _) = a.pair(&pair, "phone A", "100.64.0.1").unwrap();
    assert_ne!(a.store.server_id, b.store.server_id);
    assert!(a.authenticate(&token).is_some());
    assert!(b.authenticate(&token).is_none());
    assert!(a.authenticate("").is_none());
}

#[test]
fn new_qr_invalidates_previous_qr() {
    let dir = tempfile::tempdir().unwrap();
    let mut auth = Auth::load(dir.path()).unwrap();
    let old = auth.new_pair();
    let new = auth.new_pair();
    assert!(auth.pair(&old, "phone", "100.64.0.1").is_err());
    assert!(auth.pair(&new, "phone", "100.64.0.1").is_ok());
}

#[test]
fn expiry_and_revocation_are_per_device() {
    let dir = tempfile::tempdir().unwrap();
    let mut auth = Auth::load(dir.path()).unwrap();
    let qr = auth.new_pair();
    let (first_token, first) = auth.pair(&qr, "one", "100.64.0.1").unwrap();
    let qr = auth.new_pair();
    let (second_token, _) = auth.pair(&qr, "two", "100.64.0.2").unwrap();
    auth.revoke(&first.id).unwrap();
    assert!(auth.authenticate(&first_token).is_none());
    assert!(auth.authenticate(&second_token).is_some());
    auth.store.devices[0].expires_at = now();
    assert!(auth.authenticate(&second_token).is_none());
}

#[test]
fn persisted_store_contains_neither_bearer_nor_pem() {
    let dir = tempfile::tempdir().unwrap();
    let mut auth = Auth::load(dir.path()).unwrap();
    let qr = auth.new_pair();
    let (token, _) = auth.pair(&qr, "phone", "100.64.0.1").unwrap();
    let bytes = std::fs::read(dir.path().join("identity.dpapi")).unwrap();
    assert!(!bytes.windows(token.len()).any(|w| w == token.as_bytes()));
    assert!(!bytes.windows(10).any(|w| w == b"-----BEGIN"));
    let loaded = Auth::load(dir.path()).unwrap();
    assert_eq!(auth.store.server_id, loaded.store.server_id);
    assert_eq!(auth.store.fingerprint, loaded.store.fingerprint);
    assert!(loaded.authenticate(&token).is_some());
}

#[test]
fn pin_is_validated_and_survives_restart() {
    let dir = tempfile::tempdir().unwrap();
    let mut auth = Auth::load(dir.path()).unwrap();
    for invalid in ["", "123", "123456789", "abcd", "12 4"] {
        assert!(auth.set_pin(invalid).is_err());
    }
    auth.set_pin("192837").unwrap();
    let mut loaded = Auth::load(dir.path()).unwrap();
    assert!(loaded.login("wrong", "phone", "100.64.0.1").is_err());
    assert!(loaded.login("192837", "phone", "100.64.0.1").is_ok());
}
