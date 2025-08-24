use clap::{Parser, Subcommand};
use std::io::{self, Write};

mod crypto;
mod storage;

use crate::storage::SecretsStorage;

#[derive(Parser)]
#[command(name = "sm")]
#[command(about = "A CLI secrets manager")]
struct Cli {
    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand)]
enum Commands {
    Get {
        name: String,
    },
    Add {
        name: String,
        #[arg(short, long)]
        value: Option<String>,
    },
    List,
    Remove {
        name: String,
    },
    Password,
}

fn main() -> Result<(), Box<dyn std::error::Error>> {
    let cli = Cli::parse();
    let mut storage = SecretsStorage::new()?;

    match cli.command {
        Commands::Get { name } => {
            match storage.get_secret(&name)? {
                Some(secret) => {
                    println!("{}", secret);
                }
                None => {
                    println!("Secret '{}' not found", name);
                }
            }
        }
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
            println!("Secret '{}' added successfully!", name);
        }
        Commands::List => {
            let secrets = storage.list_secrets()?;
            if secrets.is_empty() {
                println!("No secrets stored");
            } else {
                println!("Stored secrets:");
                for secret_name in secrets {
                    println!("  - {}", secret_name);
                }
            }
        }
        Commands::Remove { name } => {
            if storage.remove_secret(&name)? {
                println!("Secret '{}' removed successfully!", name);
            } else {
                println!("Secret '{}' not found", name);
            }
        }
        Commands::Password => {
            storage.change_password()?;
            println!("Master password changed successfully!");
        }
    }

    Ok(())
}
