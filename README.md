Secrets Vault CLI

This is a terminal-based UI client for the Secrets Vault backend. It's a TUI built with tcell and configured to talk to the backend at http://localhost:8080 by default.

Build

From the `cli` directory:

```bash
cd cli
go mod download
cd cmd/cli
go build -o sm-cli
```

Run

```bash
./sm-cli
```

Implemented features (skeleton):
- tcell-based UI bootstrap
- Health check, main menu, placeholders for Login/Signup
- HTTP client helpers for backend endpoints (health, signup, login)

Planned enhancements:
- Full TUI workflows for login/signup, listing secrets, creating/updating secrets, managing API keys.
- Better keyboard navigation, non-blocking IO, input forms and masked fields.
