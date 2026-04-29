# Browser Wallet Login Example

This is a minimal browser-only example for EVM wallet login.

It does not use Next.js, Vite, or any build tool. Open `index.html` in a browser with MetaMask or another EIP-1193 wallet installed.

## Run

Start Open Wallet Auth first:

```bash
go run ./cmd/server
```

Then open:

```text
examples/browser-wallet-login/index.html
```

Default settings:

- Auth base URL: `http://localhost:8080`
- Client ID: `default`
- Chain ID: `1`

## Flow

1. Request wallet accounts with `eth_requestAccounts`.
2. Create a nonce with `POST /api/v1/wallet/nonce`.
3. Ask the wallet to sign the returned message with `personal_sign`.
4. Verify the signature with `POST /api/v1/wallet/verify`.
5. Store the returned token pair in `localStorage`.
