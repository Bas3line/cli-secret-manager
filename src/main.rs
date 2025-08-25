use clap::{Parser, Subcommand};
use std::io::{self, Write};

mod crypto;
mod storage;

use crate::storage::SecretsStorage;

use owo_colors::OwoColorize;
use std::env;
use unicode_width::UnicodeWidthStr;

fn print_pretty_help() {
    let indent = "    ";

    let title = "⌬ ".to_string();
    let description = "secrets-manager (sm) is a tool to securely store secrets locally using AES-256 encryption (https://github.com/sm-vault/sm-cli)";

    println!();
    println!("{}{} {}\n", indent, title.green().bold(), description.dimmed());
    println!("{}{}\n  {} <command> [options]\n", indent, "USAGE".bold(), "  sm".cyan());

    // commands as a neat aligned tree
    println!("{}{}", indent, "COMMANDS".bold());
    let commands = vec![
        ("get", "Retrieve a secret by name"),
        ("add", "Add a new secret (prompts if value omitted)"),
        ("list", "List stored secret names"),
        ("remove", "Remove a secret permanently"),
        ("password", "Change the master password"),
    ];
    let max_cmd_len = commands.iter().map(|(c, _)| UnicodeWidthStr::width(*c)).max().unwrap_or(0);

    for (i, (cmd, cmd_desc)) in commands.iter().enumerate() {
        let is_last = i == commands.len() - 1;
        let branch = if is_last { "└──" } else { "├──" };
        let padded_cmd = format!("{:<width$}", cmd, width = max_cmd_len);
        println!("{}  {} {}    {}", indent, branch.dimmed(), padded_cmd.green().bold(), cmd_desc.dimmed());
    }

    println!("\n{}{}", indent, "OPTIONS".bold());
    println!("{}  -h, --help    Print help\n", indent);

    println!("{}For detailed help on a command: {}\n", indent, format!("{} help <command>", "sm").dimmed());
}

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
    let args: Vec<String> = env::args().collect();

    if args.len() == 1 || (args.len() == 2 && (args[1] == "-h" || args[1] == "--help" || args[1] == "help")) {
        print_pretty_help();
        return Ok(());
    }

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
