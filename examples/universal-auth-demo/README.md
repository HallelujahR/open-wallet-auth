# Universal Auth Demo

This is a browser-only developer demo for Open Wallet Auth.

It is designed to show how a product frontend can wire multiple login methods into one page.

## Supported Now

- Email registration
- Email login
- EVM wallet login
- Wallet provider selection through EIP-6963 when the browser exposes multiple wallets

## UI Placeholders

These tabs are included to document the intended frontend contract, but the backend endpoints are not implemented yet:

- Phone login
- Google login
- GitHub login

## Run

Start Open Wallet Auth:

```bash
OWA_HTTP_PORT=8081 go run ./cmd/server
```

Serve the examples directory:

```bash
python3 -m http.server 5173
```

Open in Chrome or another browser with a wallet extension:

```text
http://localhost:5173/examples/universal-auth-demo/
```

Use this Auth Base URL:

```text
http://localhost:8081
```

The Codex in-app browser cannot load wallet extensions, so wallet login must be tested in a normal browser with MetaMask or another EIP-1193 wallet installed.
