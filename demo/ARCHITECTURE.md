# Merkle Tree Certificates (MTC) Demo — Architecture

## What are Merkle Tree Certificates?

Standard post-quantum algorithms (e.g. ML-DSA) produce signatures and public keys significantly larger than traditional algorithms. Embedding them in traditional X.509 certificate chains inflates TLS handshake size, causing fragmentation and latency.

**Merkle Tree Certificates** solve this via batch signing and **Merkle inclusion proofs**: instead of sending an entire signed certificate, the server transmits a compact hash path proving the certificate belongs to a batch already signed by the CA.

---

## Demonstration Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                       Browser (TLS 1.3)                     │
│  ┌────────────────────┐   ┌──────────────────────────────┐  │
│  │   Left Panel       │   │   Right Panel (tabbed)       │  │
│  │  ─ Connection Info │   │  ─ Merkle Proof Chain        │  │
│  │  ─ Artifact Sizes  │   │  ─ Payload Decoder           │  │
│  │  ─ Live Verify     │   │                              │  │
│  └────────────────────┘   └──────────────────────────────┘  │
└──────────────────────────┬──────────────────────────────────┘
                           │ HTTPS (TLS 1.3 only, :8443)
┌──────────────────────────▼──────────────────────────────────┐
│               Go Web Server (demo/main.go)                  │
│                                                             │
│  GET /              → static/index.html                     │
│  GET /api/verify           → mtc-cli verify (Root CA)       │
│  GET /api/verify/landmark  → mtc-cli verify (Landmark CA)   │
│  GET /api/inspect?target=  → mtc-cli inspect (truncated)    │
│  GET /api/file-sizes       → os.Stat() on artifacts         │
│  GET /api/proof-chain      → 6-step chain walk              │
│  GET /api/files/*          → raw artifact download          │
└────────────────────┬────────────────────────────────────────┘
                     │ exec.Command
┌────────────────────▼────────────────────────────────────────┐
│                    mtc-cli (binary)                         │
│   verify  │  inspect cert/validity-window/ca-params         │
│   ca new  │  ca queue  │  ca issue  │  cert  │  new-asr     │
└────────────────────┬────────────────────────────────────────┘
                     │
        ┌────────────┴────────────┐
        │                        │
┌───────▼───────┐       ┌────────▼──────┐
│   Root CA     │       │  Landmark CA  │
│  ca/          │       │  landmark-ca/ │
│  website.mtc  │       │  landmark-    │
│  website.vw   │       │  website.mtc  │
│  ca-params    │       │  landmark.vw  │
└───────────────┘       └───────────────┘
```

---

## Component Detail

### 1. PKI — Root CA & Landmark CA

Both CAs are initialised by `demo/setup.sh` using `mtc-cli ca new`:

| Parameter        | Value              |
|------------------|--------------------|
| Batch duration   | 2 seconds          |
| Lifetime         | 7 days (168h)      |
| Storage duration | 14 days (336h)     |
| Window size      | ~302,400 tree heads |
| Window size (MB) | ~9.2 MB            |

The **Root CA** (OID `62253.12.15`) issues certs for `localhost` / `127.0.0.1`.  
The **Landmark CA** (OID `62253.12.15.1`) acts as an intermediary, independently issuing the same website assertion — demonstrating a two-level trust hierarchy.

### 2. Certificate Artifacts

| File | Description |
|------|-------------|
| `website.mtc` | Leaf certificate: assertion + Merkle inclusion proof (128 B) |
| `website.vw` | Root CA signed validity window (~9.23 MB) |
| `landmark-website.mtc` | Website cert via Landmark CA path (161 B) |
| `landmark.vw` | Landmark CA signed validity window (~9.23 MB) |
| `ca/www/mtc/v04b/ca-params` | Root CA public parameters (2.6 KB) |
| `landmark-ca/www/mtc/v04b/ca-params` | Landmark CA public parameters |
| `website.pem` + `website.key` | Self-signed X.509 for TLS fallback (browsers don't speak MTC yet) |

### 3. Go Web Server (`demo/main.go`)

Key design decisions:

- **TLS 1.3 only** — `MinVersion` and `MaxVersion` both set to `tls.VersionTLS13`.
- **Validity window streaming** — The `inspect validity-window` command emits one line per batch slot (~302k lines for a 7-day window). The server streams and **stops reading after 30 lines**, kills the child process, and appends a truncation notice. This prevents the API from hanging the browser.
- **Landmark cert path** — Prefers `landmark-website.mtc`; falls back to `landmark.mtc` if absent, both in the inspect handler and the file-sizes handler.
- **CSP-safe frontend** — All DOM construction in `index.html` uses `createElement` + `textContent` rather than `innerHTML`, avoiding Trusted Types violations.

### 4. Frontend UI (`demo/static/index.html`)

Single-screen landscape layout fitting a 1800×940 viewport without scrolling:

```
┌─ Header (64px) ─────────────────────────────────── TLS 1.3 Active ─┐
├─ Left Panel (340px) ─┬─ Right Panel (flex) ──────────────────────────┤
│ Connection Security  │ [🌳 Proof Chain] [🔬 Payload Decoder]          │
│ Artifact Sizes       │                                               │
│ ── ── ── ── ── ──    │  (tab content scrolls internally)             │
│ Live Verification    │                                               │
│  [Verify Root CA]    │                                               │
│  [Verify Landmark]   │                                               │
└──────────────────────┴───────────────────────────────────────────────┘
```

### 5. Secure Containerisation & CI/CD

- **Build stage:** `cgr.dev/chainguard/wolfi-base` — compiles Go backend and invokes PKI setup.
- **Runtime stage:** `cgr.dev/chainguard/static` — zero-CVE distroless image; only binaries and crypto artifacts.
- **CI/CD:** GitHub Actions multi-arch build for `linux/amd64` and `linux/arm64` via QEMU + Buildx.

---

## Live Screenshots

### Main View — Landscape Layout
![Main view](./screenshots/01_main_view.png)

### Merkle Proof Chain — All Steps Verified
![Proof chain](./screenshots/04_proof_chain.png)

### Root CA Verification
![Root verify](./screenshots/02_verify_root.png)

### Landmark CA Verification
![Landmark verify](./screenshots/03_verify_landmark.png)

### Payload Decoder — Leaf Certificate
![Payload decoder](./screenshots/05_payload_decoder.png)

### Payload Decoder — Landmark Certificate (161 B)
![Landmark cert](./screenshots/06_landmark_cert.png)

### Payload Decoder — Landmark Validity Window (Truncated)
![Landmark VW](./screenshots/07_landmark_vw.png)

---

## Running Locally

```bash
cd demo
./setup.sh          # Generate PKI artifacts (~30s)
./website-server    # Start TLS 1.3 server on :8443
```

Open **`https://localhost:8443`** and accept the self-signed certificate warning.

Or via Docker:
```bash
docker build -t mtc-demo:latest .
docker run -p 8443:8443 mtc-demo:latest
```
