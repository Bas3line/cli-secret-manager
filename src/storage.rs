use crate::crypto::SecretsCrypto;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::fs;
use std::path::PathBuf;
use std::error::Error;

#[derive(Serialize, Deserialize, Default)]
struct SecretsData {
    secrets: HashMap<String, String>,
}

pub struct SecretsStorage {
    file_path: PathBuf,
    crypto: SecretsCrypto,
    data: SecretsData,
}

impl SecretsStorage {
    pub fn new() -> Result<Self, Box<dyn Error>> {
        let file_path = Self::get_storage_path()?;
        let password = if file_path.exists() {
            println!("Enter master password: ");
            rpassword::read_password()?
        } else {
            println!("Creating new secrets store.");
            println!("Enter master password: ");
            let pwd1 = rpassword::read_password()?;
            println!("Confirm master password: ");
            let pwd2 = rpassword::read_password()?;
            if pwd1 != pwd2 {
                return Err("Passwords don't match".into());
            }
            pwd1
        };
        let crypto = SecretsCrypto::new(&password);
        let data = Self::load_data(&file_path, &crypto)?;
        Ok(Self {
            file_path,
            crypto,
            data,
        })
    }

    fn get_storage_path() -> Result<PathBuf, Box<dyn Error>> {
        let mut path = dirs::home_dir().ok_or("Could not find home directory")?;
        path.push(".secrets-manager");
        fs::create_dir_all(&path)?;
        path.push("secrets.enc");
        Ok(path)
    }

    fn load_data(file_path: &PathBuf, crypto: &SecretsCrypto) -> Result<SecretsData, Box<dyn Error>> {
        if !file_path.exists() {
            return Ok(SecretsData::default());
        }
        let encrypted_content = fs::read_to_string(file_path)?;
        if encrypted_content.is_empty() {
            return Ok(SecretsData::default());
        }
        let decrypted_content = crypto.decrypt(&encrypted_content)?;
        let data: SecretsData = serde_json::from_str(&decrypted_content)?;
        Ok(data)
    }

    fn save_data(&self) -> Result<(), Box<dyn Error>> {
        let json_data = serde_json::to_string(&self.data)?;
        let encrypted_data = self.crypto.encrypt(&json_data)?;
        fs::write(&self.file_path, encrypted_data)?;
        Ok(())
    }

    pub fn add_secret(&mut self, name: &str, value: &str) -> Result<(), Box<dyn Error>> {
        self.data.secrets.insert(name.to_string(), value.to_string());
        self.save_data()?;
        Ok(())
    }

    pub fn get_secret(&self, name: &str) -> Result<Option<String>, Box<dyn Error>> {
        Ok(self.data.secrets.get(name).cloned())
    }

    pub fn list_secrets(&self) -> Result<Vec<String>, Box<dyn Error>> {
        let mut names: Vec<String> = self.data.secrets.keys().cloned().collect();
        names.sort();
        Ok(names)
    }

    pub fn remove_secret(&mut self, name: &str) -> Result<bool, Box<dyn Error>> {
        let removed = self.data.secrets.remove(name).is_some();
        if removed {
            self.save_data()?;
        }
        Ok(removed)
    }

    pub fn change_password(&mut self) -> Result<(), Box<dyn Error>> {
        println!("Enter new master password: ");
        let pwd1 = rpassword::read_password()?;
        println!("Confirm new master password: ");
        let pwd2 = rpassword::read_password()?;
        if pwd1 != pwd2 {
            return Err("Passwords don't match".into());
        }
        self.crypto = SecretsCrypto::new(&pwd1);
        self.save_data()?;
        Ok(())
    }
}
