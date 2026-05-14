# Merkle Tree Certificates (MTC) Demo Architecture

## What are Merkle Tree Certificates?
Standard post-quantum algorithms produce signatures and public keys that are significantly larger than traditional algorithms. Using them in traditional X.509 certificate chains would increase the size of the TLS handshake, potentially leading to fragmentation and latency.

Merkle Tree Certificates (MTCs) solve this by using batch signing and **Merkle inclusion proofs**. Instead of sending an entire certificate with signatures, the server simply sends a compact sequence of hashes proving the certificate is part of a batch signed by the Certificate Authority (CA).

## Demonstration Architecture
In this demonstration, we've built the following components:

### 1. The Multi-CA Infrastructure
The demo features a dual-CA hierarchy to simulate a real-world Landmark CA cross-verification path:
* **Root CA:** Located in `demo/ca`, issues the primary batch every 2 seconds.
* **Landmark CA:** Located in `demo/landmark-ca`, acts as a trusted cross-signer, providing a secondary verification path for the same assertion.
* Both CAs maintain live **Validity Windows** with ~300k tree heads, simulating a full storage window of historical checkpoints.

### 2. The Certificate Artifacts
The environment generates several cryptographic artifacts for the `localhost` identity:
* `website.mtc`: The primary Merkle Tree Certificate.
* `website.vw`: The Root CA's signed validity window (~9.2 MB).
* `landmark.vw`: The Landmark CA's signed validity window.
* `landmark-website.mtc`: A certificate issued specifically through the Landmark CA's path.

### 3. The TLS 1.3 Web Server & UI
Located in `demo/main.go`, the server provides the following capabilities:
* **TLS 1.3 Enforcement:** Secures all demo traffic using modern TLS 1.3.
* **Proof Chain API:** A specialized endpoint (`/api/proof-chain`) that performs a multi-stage verification across both CAs, returning a structured trace of the cryptographic validation.
* **Truncated Streaming Inspect:** A performance-optimized diagnostics API that streams large validity window outputs while truncating them for UI performance.
* **Interactive Explorer:** A premium glassmorphism-inspired UI at `https://localhost:8443` featuring:
    * **Live Proof Path:** Visualizes the journey from CA parameters to the leaf assertion.
    * **Multi-Path Verification:** Compare verification against Root vs. Landmark CA paths.
    * **Payload Decoder:** Real-time parsing of binary MTC structures.

### 4. Secure Containerization & CI/CD Pipeline
To ensure the demo runs securely and consistently, the environment is containerized using **Chainguard** hardened images.
* **Distroless Runtime:** Uses `cgr.dev/chainguard/static` to minimize attack surface.
* **Multi-Arch Native Build:** The `mtc-cli` and Go backend are built natively for the host architecture during the container build stage.

## Two-Dashboard Experience

The container hosts two distinct web applications:

1.  **Main MTC Demo (`:8443`):** Focused on the full end-to-end TLS 1.3 workflow with a persistent Landmark CA and live validity window monitoring.
2.  **MTC Playground (`:8444`):** Focused on interactive "on-demand" generation of various MTC formats (Spec-compliant vs. Embedded) for experimentation. Served over TLS 1.3.

## How to Run the Demo

To launch the web server, simply navigate to the `demo/` directory and run:

```bash
cd demo
go run main.go
```

Then, open your browser and navigate to `https://localhost:8443` (accept the self-signed X.509 warning, which acts as the TLS fallback). 
Click the **Verify Merkle Tree Certificate** button to execute a live backend verification of the Inclusion Proof against the CA parameters.

Alternatively, to run the secure Chainguard container directly without requiring Go on your host:
```bash
docker build -t mtc-demo:latest .
docker run -p 8443:8443 mtc-demo:latest
```
