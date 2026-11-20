# Build & Functional Verification Report

- **Verification Date:** `2026-11-20`
- **Verified By:** **Jeffs' Automated build and deploy agent**
- **Repository:** `jsoehner/merkle-demo` (branch `main`, standalone repository — fork guardrail passed)
- **Overall Status:** `VERIFIED`

---

## 1. Trusted OSS Split-Horizon Configuration

Configured via `setup_trusted_oss_repo.py` (`--mode virtual`):

| Scope | Setting | Value |
| :--- | :--- | :--- |
| Local LAN (`.env`, excluded via `.git/info/exclude`) | `TRUSTED_OSS_CACHE_URL` | `http://oss.jsigroup.local:8080` |
| GitHub Actions Variable | `TRUSTED_OSS_CACHE_URL` | `https://jsigroup-oss.duckdns.org` |
| GitHub Actions Variable | `TRUSTED_OSS_MODE` | `virtual` |
| GitHub Actions Variable | `TRUSTED_OSS_GPG_EMAIL` | `oss-signing@the-techy.com` |
| GitHub Actions Secrets | `OSS_MTLS_CLIENT_CERT`, `OSS_MTLS_CLIENT_KEY` | Provisioned from local mTLS PKI |

> **Registry Note:** The multi-stage Chainguard Wolfi `Dockerfile` compiles the Go backends (`demo/main.go`, `playground/main.go`), upstream `bwesterb/mtc` CLI, and `digicert/ca-extension-mtc-playground` tools inside the container image using Go modules and Wolfi APK repositories. Local `.env` and package-manager LAN configs are excluded from git via `.git/info/exclude`.

---

## 2. Build & Runtime Commands

```bash
# Build both Chainguard Wolfi multi-stage images and launch the TLS 1.3 stack
docker compose up --build -d

# Clean teardown after verification
docker compose down -v
```

- **`mtc-demo-website`** (`https://localhost:8443`): Built cleanly (`demo-runtime`), generated Root CA + Landmark CA MTC artifacts via `setup.sh`, and reached Docker `healthy` status.
- **`mtc-playground`** (`https://localhost:8444`): Built cleanly (`playground-runtime`), mounted shared TLS certificates from `shared-artifacts`, and reached Docker `healthy` status.

---

## 3. Functional Verification Results

### 3.1 Guided Headless Chrome (CDP) Verification

Executed with `browser_verify.py` against both TLS 1.3 web interfaces:

| Service | URL | HTTP | Page Title | Expected Text | Interactive Clicks | Console Errors | JS Exceptions |
| :--- | :--- | :---: | :--- | :---: | :--- | :---: | :---: |
| MTC Demo Website | `https://localhost:8443/` | `200` | `Post-Quantum TLS Ready - Merkle Tree Certificates` | Found (`Next-Gen WebPKI`) | `#verifyBtn` (`CLICKED`), `#verifyLmBtn` (`CLICKED`) | `0` | `0` |
| MTC Playground | `https://localhost:8444/` | `200` | `DigiCert MTC Playground — CA Extension Demo` | Found (`DigiCert MTC Playground`) | `#runDemoBtn` (`CLICKED`), `#runVerifyBtn` (`CLICKED`) | `0` | `0` |

### 3.2 `curl` API & Security Verification

Executed with `api_verify.py` (13 endpoints total — all passed):

| Service | Endpoint | Method | Expected | Actual | Result Summary |
| :--- | :--- | :---: | :---: | :---: | :--- |
| Demo (`:8443`) | `/` | `GET` | `2xx` | `200` | HTML UI (`32,094` bytes) |
| Demo (`:8443`) | `/api/verify` | `GET` | `2xx` | `200` | JSON `{"success":true,"output":""}` (Root CA MTC verification) |
| Demo (`:8443`) | `/api/verify/landmark` | `GET` | `2xx` | `200` | JSON `{"success":true,"output":""}` (Landmark CA MTC verification) |
| Demo (`:8443`) | `/api/file-sizes` | `GET` | `2xx` | `200` | JSON artifact sizes (`website.mtc` 128 B, `landmark-website.mtc` 161 B, `website.vw` ~9.2 MB) |
| Demo (`:8443`) | `/api/proof-chain` | `GET` | `2xx` | `200` | JSON `{"success":true,...}` full Root -> Landmark -> Leaf proof chain (`5,595` bytes) |
| Demo (`:8443`) | `/api/inspect?target=cert&ca=root` | `GET` | `2xx` | `200` | JSON decoded MTC leaf payload (`subject_type TLS`, `p256`, `merkle_tree_sha256`) |
| Demo (`:8443`) | `/api/files/cert` | `GET` | `2xx` | `200` | Binary `application/octet-stream` (`128` bytes) |
| Playground (`:8444`) | `/` | `GET` | `2xx` | `200` | HTML UI (`29,223` bytes) |
| Playground (`:8444`) | `/api/health` | `GET` | `2xx` | `200` | JSON `{"ok":true,"playground_ok":true,...}` |
| Playground (`:8444`) | `/api/conformance-info` | `GET` | `2xx` | `200` | JSON spec metadata (`draft-ietf-plants-merkle-tree-certs-01`) |
| Playground (`:8444`) | `/api/demo?domain=demo.example.com&cert_mode=mtc` | `GET` | `2xx` | `200` | JSON `{"success":true,...}` MTC-Spec certificate generation |
| Playground (`:8444`) | `/api/verify?domain=demo.example.com&cert_mode=mtc` | `GET` | `2xx` | `200` | JSON `{"success":true,...}` end-to-end MTC generation + verification |
| Playground (`:8444`) | `/api/verify?domain=bad_domain!&cert_mode=mtc` | `GET` | `400` | `400` | Negative validation test: `{"error":"Invalid domain name"}` |

### 3.3 TLS 1.3 Enforcement Check

- `curl -k --tls-max 1.2 -sS https://localhost:8443/` was rejected at the TLS handshake layer, confirming strict TLS 1.3 enforcement (`MinVersion: tls.VersionTLS13`).
