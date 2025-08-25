use clap::{Parser, Subcommand};
use std::io::{self, Write};
use reqwest::blocking::Client;
use reqwest::StatusCode;
use serde::Deserialize;

mod storage;
mod crypto;
use storage::SecretsStorage;

#[derive(Parser)]
#[command(name = "sm")]
#[command(about = "A CLI secrets manager")]
struct Cli {
    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand)]
enum Commands {
    Get { name: String },
    Add { name: String, #[arg(short, long)] value: Option<String> },
    List,
    Remove { name: String },
    Password,
    CreateApi,
    ImportApi {
        #[arg(long)]
        api_key: Option<String>,
    },
}

fn main() -> Result<(), Box<dyn std::error::Error>> {
    let cli = Cli::parse();
    let mut storage = SecretsStorage::new()?;

    match cli.command {
        Commands::Get { name } => {
            match storage.get_secret(&name)? {
                Some(secret) => println!("{}", secret),
                None => println!("Secret '{}' not found", name),
            }
        },
        Commands::Add { name, value } => {
            let secret_value = match value {
                Some(v) => v,
                None => {
                    print!("Enter secret value: ");
                    io::stdout().flush()?;
                    rpassword::read_password()?
                }
            };
            storage.add_secret(&name, &secret_value)?;
            println!("Secret '{}' added.", name);
        },
        Commands::List => {
            let secrets = storage.list_secrets()?;
            if secrets.is_empty() {
                println!("No secrets stored");
            } else {
                println!("Stored secrets:");
                for name in secrets {
                    println!("  - {}", name);
                }
            }
        },
        Commands::Remove { name } => {
            if storage.remove_secret(&name)? {
                println!("Secret '{}' removed.", name);
            } else {
                println!("Secret '{}' not found", name);
            }
        },
        Commands::Password => {
            storage.change_password()?;
            println!("Master password changed.");
        },
        Commands::CreateApi => {
            create_api_key(&mut storage)?;
        },
        Commands::ImportApi { api_key } => {
            import_from_api(&mut storage, &api_key)?;
        },
    }
    Ok(())
}

fn create_api_key(storage: &mut SecretsStorage) -> Result<(), Box<dyn std::error::Error>> {
    print!("Enter your email: ");
    io::stdout().flush()?;
    let mut email = String::new();
    io::stdin().read_line(&mut email)?;
    let email = email.trim();

    print!("Enter your password: ");
    io::stdout().flush()?;
    let password = rpassword::read_password()?;

    print!("Enter API key name: ");
    io::stdout().flush()?;
    let mut key_name = String::new();
    io::stdin().read_line(&mut key_name)?;
    let key_name = key_name.trim();

    let client = Client::new();
    let login_payload = serde_json::json!({
        "email": email,
        "password": password,
        "master_password": password
    });

    let resp = client
        .post("http://localhost:8080/api/v1/auth/login")
        .header("Content-Type", "application/json")
        .json(&login_payload)
        .send()?;

    if resp.status() == StatusCode::UNAUTHORIZED {
        let error = resp.text().unwrap_or_default().to_lowercase();
        if error.contains("user not found") || error.contains("no user found") {
            println!("User not found. Please sign up at https://smvault.tech/signup");
            return Ok(());
        } else if error.contains("invalid password") {
            println!("Invalid password. Please check your credentials.");
            return Ok(());
        } else {
            println!("Authentication failed.");
            return Ok(());
        }
    }
    if !resp.status().is_success() {
        println!("Login failed: {}", resp.status());
        return Ok(());
    }
    #[derive(Deserialize)]
    struct LoginResponse {
        token: String,
    }
    let login_data: LoginResponse = resp.json()?;

    let api_key_resp = client
        .post("http://localhost:8080/api/v1/apikeys")
        .header("Authorization", format!("Bearer {}", login_data.token))
        .header("Content-Type", "application/json")
        .json(&serde_json::json!({ "name": key_name }))
        .send()?;

    if !api_key_resp.status().is_success() {
        let error = api_key_resp.text().unwrap_or_default();
        println!("Failed to create API key: {}", error);
        return Ok(());
    }

    #[derive(Deserialize)]
    struct ApiKeyResponse {
        api_key: ApiKey,
    }
    #[derive(Deserialize)]
    struct ApiKey {
        key: String,
        name: String,
    }
    let api_response: ApiKeyResponse = api_key_resp.json()?;

    storage.set_api_key(&api_response.api_key.key)?;

    println!("API key '{}' created and stored.", key_name);
    println!("Key: {}", api_response.api_key.key);
    println!("This is the only time the key will be shown. Store it safely.");
    Ok(())
}

fn import_from_api(storage: &mut SecretsStorage, api_key_option: &Option<String>) -> Result<(), Box<dyn std::error::Error>> {
    let api_key = match api_key_option {
        Some(key) => key.clone(),
        None => match storage.get_api_key() {
            Some(stored_key) => stored_key,
            None => {
                println!("No API key found. Use --api-key or 'sm create-api'");
                return Ok(());
            }
        }
    };

    println!("Fetching secrets from server...");
    let client = Client::new();
    let resp = client
        .get("http://localhost:8080/api/v1/secrets")
        .header("X-API-Key", &api_key)
        .send()?;

    if resp.status() == StatusCode::UNAUTHORIZED {
        println!("Failed to fetch secrets: 401 Unauthorized");
        println!("API key might be expired. Try 'sm create-api' again.");
        return Ok(());
    }
    if !resp.status().is_success() {
        println!("Failed to fetch secrets: {}", resp.status());
        return Ok(());
    }
    #[derive(Deserialize)]
    struct ApiSecret {
        name: String,
        value: String,
    }
    #[derive(Deserialize)]
    struct SecretsResponse {
        secrets: Vec<ApiSecret>,
    }
    let api_response: SecretsResponse = resp.json()?;
    let mut imported = 0;
    for secret in api_response.secrets {
        if storage.get_secret(&secret.name)?.is_some() {
            print!("Secret '{}' already exists. Overwrite? (y/N): ", secret.name);
            io::stdout().flush()?;
            let mut input = String::new();
            io::stdin().read_line(&mut input)?;
            if !input.trim().to_lowercase().starts_with('y') {
                println!("Skipping '{}'", secret.name);
                continue;
            }
        }
        storage.add_secret(&secret.name, &secret.value)?;
        println!("Imported: {}", secret.name);
        imported += 1;
    }
    println!("Successfully imported {} secrets.", imported);
    Ok(())
}
