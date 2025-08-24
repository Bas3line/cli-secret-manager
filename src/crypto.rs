use aes_gcm::{
    aead::{Aead, KeyInit, rand_core::RngCore},
    Aes256Gcm, Nonce
};
use sha2::{Digest, Sha256};
use std::error::Error;
use rand::rngs::OsRng;

pub struct SecretsCrypto {
    cipher: Aes256Gcm,
}

impl SecretsCrypto {
    pub fn new(password: &str) -> Self {
        let key = Self::derive_key(password);
        let cipher = Aes256Gcm::new(&key);
        Self { cipher }
    }

    fn derive_key(password: &str) -> aes_gcm::Key<Aes256Gcm> {
        let mut hasher = <Sha256 as Digest>::new();
        hasher.update(password.as_bytes());
        let hash = hasher.finalize();
        *aes_gcm::Key::<Aes256Gcm>::from_slice(&hash)
    }

    pub fn encrypt(&self, data: &str) -> Result<String, Box<dyn Error>> {
        let mut nonce_bytes = [0u8; 12];
        OsRng.fill_bytes(&mut nonce_bytes);
        let nonce = Nonce::from_slice(&nonce_bytes);

        let ciphertext = self.cipher.encrypt(nonce, data.as_bytes())
            .map_err(|e| format!("Encryption error: {}", e))?;

        let mut result = nonce_bytes.to_vec();
        result.extend_from_slice(&ciphertext);
        Ok(hex::encode(result))
    }

    pub fn decrypt(&self, encrypted_data: &str) -> Result<String, Box<dyn Error>> {
        let data = hex::decode(encrypted_data)?;
        if data.len() < 12 {
            return Err("Invalid encrypted data".into());
        }
        let (nonce_bytes, ciphertext) = data.split_at(12);
        let nonce = Nonce::from_slice(nonce_bytes);
        let plaintext = self.cipher.decrypt(nonce, ciphertext)
            .map_err(|e| format!("Decryption error: {}", e))?;
        Ok(String::from_utf8(plaintext)?)
    }
}
