## Focus: Security

Review the diff for **exploitable vulnerabilities** — code an attacker could use to compromise confidentiality, integrity, or availability.

**Flag:**
- **Injection** — user-controlled input concatenated into SQL, shell commands, OS exec args, or template strings without sanitization or parameterization
- **Secrets in code** — API keys, passwords, tokens hardcoded or logged; credentials committed to version control
- **Path traversal** — file paths constructed from user input without cleaning (`../` escape)
- **Insecure crypto** — MD5/SHA1 for integrity, ECB mode, hardcoded IVs, custom crypto schemes
- **Missing auth checks** — endpoints or operations accessible without verifying identity or permissions
- **SSRF / open redirect** — URLs constructed from user input and fetched server-side without allowlist
- **Race conditions with security impact** — TOCTOU on file permissions, double-spend on balance checks

**Do not flag:**
- Missing rate limiting or CSRF tokens unless the diff introduces a new endpoint that clearly needs them
- Theoretical timing attacks on non-cryptographic operations
