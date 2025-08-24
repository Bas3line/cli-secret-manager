# CLI Secret Manager (sm)

Simple CLI tool to securely store API keys and secrets locally with **AES-256 encryption**.

## Tech Used

![Rust](https://img.shields.io/badge/Rust-000000?style=for-the-badge&logo=rust&logoColor=white)
![Serde](https://img.shields.io/badge/Serde-000000?style=for-the-badge&logo=rust&logoColor=white)
![Clap](https://img.shields.io/badge/Clap-000000?style=for-the-badge&logo=rust&logoColor=white)
![AES](https://img.shields.io/badge/AES--GCM-0A0A0A?style=for-the-badge&logo=openssl&logoColor=white)
![SHA2](https://img.shields.io/badge/SHA2-333333?style=for-the-badge&logo=gnuprivacyguard&logoColor=white)
![Rand](https://img.shields.io/badge/Rand-222222?style=for-the-badge&logo=rust&logoColor=white)

## Installation

```bash
git clone https://github.com/Bas3line/cli-secret-manager.git
cd cli-secret-manager
cargo build --release
sudo cp target/release/sm /usr/local/bin/
````

## Usage

```bash
# Add secret (secure prompt)
sm add github-token

# Add secret with value
sm add api-key your-secret-here

# Get secret
sm get github-token

# List all secrets
sm list

# Remove secret
sm remove old-key

# Change master password
sm password
```

## Security

* AES-256-GCM encryption
* SHA-256 key derivation
* Random nonces for each entry
* Local storage only (no external servers)

## Storage Location

Secrets are stored securely at:

```
# This is Encrypted
~/.secrets-manager/secrets.enc
```

## Tested Environment

This project has been tested on:

* **OS:** Arch Linux (x86\_64)
* **Kernel:** Linux 6.16.1-arch1-1
* **WM/DE:** KDE Plasma 6.4.4 (Wayland)

## License

[MIT](LICENSE)