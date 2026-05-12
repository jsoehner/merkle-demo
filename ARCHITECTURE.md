# Merkle Tree Certificates (MTC) Demo Architecture

## What are Merkle Tree Certificates?
Standard post-quantum algorithms produce signatures and public keys that are significantly larger than traditional algorithms. Using them in traditional X.509 certificate chains would increase the size of the TLS handshake, potentially leading to fragmentation and latency.

Merkle Tree Certificates (MTCs) solve this by using batch signing and **Merkle inclusion proofs**. Instead of sending an entire certificate with signatures, the server simply sends a compact sequence of hashes proving the certificate is part of a batch signed by the Certificate Authority (CA).

## Demonstration Architecture
In this demonstration, we've built the following components:

### 1. The PKI (Certificate Authority)
Located in `demo/ca`, the MTC Certificate Authority was created using the `bwesterb/mtc` Go CLI. 
* It issues batches every 2 seconds with a 1-hour lifetime.
* It signed an assertion request (subject identity + claim) for `localhost` and `127.0.0.1` using an ECDSA prime256v1 public key.
* The CA publishes the signed validity window and the Merkle tree containing our assertion.

### 2. The Certificate Artifacts
The CA issued the following specific artifacts for our website:
* `website.mtc`: The Merkle Tree Certificate containing the assertion and the authentication path (inclusion proof).
* `website.vw`: The signed validity window, providing the trusted checkpoint of tree heads.
* `ca-params`: The public parameters of the CA.

### 3. The TLS 1.3 Web Server
Located in `demo/main.go`, this is a Go backend running on standard TLS 1.3. Because MTC is currently an experimental IETF draft, mainstream browsers do not accept MTC directly in the TLS handshake natively yet. 
To demonstrate it, our server:
* Secures the connection over standard TLS 1.3.
* Exposes an API endpoint (`/api/verify`) that triggers a live verification of our `website.mtc` using the CA's validity window.
* Exposes a diagnostics API (`/api/inspect`) to parse and visualize the raw binary MTC payloads on the frontend.
* Exposes the MTC components statically at `/.well-known/mtc/` mimicking how an interoperable system would query them.
* Serves a premium, glassmorphism-inspired dark mode frontend where you can visually trigger and observe the certificate inclusion proof verification as well as deeply inspect the decoded structures.

### 4. Secure Containerization & CI/CD Pipeline
To ensure the demo runs securely and consistently without manual dependencies, the entire environment is containerized using **Chainguard** hardened images.
* **Build Stage:** Utilizes `cgr.dev/chainguard/wolfi-base` to compile the Go backend, the `mtc-cli`, and invoke the PKI generation script natively during the container build.
* **Runtime Stage:** Employs the zero-CVE `cgr.dev/chainguard/static` distroless image to host only the statically compiled binaries and cryptographic artifacts, completely removing the attack surface of a traditional OS environment.
* **Multi-Arch CI:** A GitHub Actions workflow securely builds this container natively for both `amd64` and `arm64` using QEMU emulation and Docker Buildx.
* **Playground Dashboard:** A second service running on port `8444` that wraps the `ca-extension-mtc-playground` standalone tools, providing an interactive environment for certificate generation and verification.

## Two-Dashboard Experience

The container hosts two distinct web applications:

1.  **Main MTC Demo (`:8443`):** Focused on the full end-to-end TLS 1.3 workflow with a persistent Landmark CA and live validity window monitoring.
2.  **MTC Playground (`:8444`):** Focused on interactive "on-demand" generation of various MTC formats (Spec-compliant vs. Embedded) for experimentation.

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
