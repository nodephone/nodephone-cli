# Security Policy & Vulnerability Disclosure

## 🔒 Security Architecture

NodePhone CLI places paramount emphasis on securing user credentials, API keys, and connection tokens.

### Credential Encryption Specs
- **Algorithm**: AES-256-GCM (Authenticated Galois/Counter Mode encryption).
- **Key Generation**: SHA-256 derived from local machine salt and host hardware identity.
- **Storage Location**: `~/.nodephone/credentials.json` with strict user-only read/write permissions (`0600`).
- **Transport Layer**: All outbound communications to NodePhone servers mandate TLS (HTTPS/WSS).

---

## 🛡️ Reporting Vulnerabilities

If you discover a potential security vulnerability in NodePhone CLI or NodePhone Cloud endpoints, please report it privately:

1. **Email**: Send details to `security@nodephone.dev`.
2. **Details to include**:
   - Step-by-step description to reproduce the vulnerability.
   - Command flags or environment settings used.
   - Operating system and NodePhone CLI version (`nodephone version`).
3. **Response Time**: We acknowledge receipt of security reports within 24 hours and aim to release patches within 7 days for critical vulnerabilities.

Please **do not** report security vulnerabilities via public GitHub issues.
