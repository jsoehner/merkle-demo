# Merkle Tree Certificates (MTC) Demonstration

[![Post-Quantum TLS Ready](https://img.shields.io/badge/Security-Post--Quantum_Ready-4f46e5.svg)](#)
[![TLS 1.3](https://img.shields.io/badge/Protocol-TLS_1.3-10b981.svg)](#)
[![Go Web Server](https://img.shields.io/badge/Backend-Go_1.21+-00add8.svg)](#)
[![Docker](https://img.shields.io/badge/Container-Chainguard_Hardened-2496ED.svg)](#)
[![Multi-Arch](https://img.shields.io/badge/Architecture-amd64_%7C_arm64-ff69b4.svg)](#)

This project provides a fully self-contained, automated demonstration of **Merkle Tree Certificates (MTC)**—an experimental Internet Engineering Task Force (IETF) architectural proposal designed to fix the severe performance bottlenecks caused by massive Post-Quantum Cryptography (PQC) signature sizes in standard TLS handshakes.

![Main UI showing Connection Security](./demo/screenshots/main_ui.png)

## What are Merkle Tree Certificates?

As the web transitions to quantum-resistant algorithms, standard certificates equipped with PQC signatures (like ML-DSA) balloon in size. This massive data overhead leads to TLS handshake fragmentation, high latency, and severe middlebox compatibility issues.

Instead of bundling multiple massive signatures inside every certificate, MTCs introduce an elegant solution:
1. **Batch Signing:** A Certificate Authority (CA) aggregates hundreds of certificates into a Merkle tree and signs only the root hash.
2. **Inclusion Proofs:** Servers simply transmit a compact Merkle authentication path (an inclusion proof) to prove they are part of the CA's signed tree.
3. **Validity Windows:** Clients verify the proofs against trusted checkpoints (signed validity windows) fetched out-of-band.

The result? The internet gets quantum-level security without sacrificing the speed of modern TLS.

## Features

* **Automated PKI Generation:** Spin up a local Merkle Tree CA using Cloudflare's reference implementation in a single command.
* **Certificate Issuance Simulation:** Automates the queueing and batch issuance of assertions into valid MTC artifacts (`.mtc`, `.vw`, and `ca-params`).
* **TLS 1.3 Web Server:** A Go-based backend that enforces TLS 1.3 and statically exposes the generated MTC artifacts in a `/.well-known/mtc/` style structure.
* **Live In-Browser Verification:** Real-time backend verification of the certificate's inclusion proof, mimicking an interoperable TLS client.
* **Payload Decoder & Diagnostics:** Explore the internals of the experimental binary structures (CA Params, Validity Window, Certificate payload) right from the browser.
* **Secure Containerization:** Containerized using an ultra-hardened, zero-CVE Chainguard distroless base image for the absolute minimum attack surface.
* **Multi-Arch CI Pipeline:** Fully automated GitHub Actions workflow to build the secure container concurrently across `linux/amd64` and `linux/arm64` via QEMU and Buildx.

![Payload Diagnostics & Decoded Views](./demo/screenshots/diagnostics.png)

## Running with Docker (Recommended)

You can build and run this entire demonstration securely inside a hardened Chainguard container without manually compiling the Go binaries on your host machine.

```bash
# Build the multi-arch image locally
docker build -t mtc-demo:latest .

# Run the container mapping the secure port
docker run -p 8443:8443 mtc-demo:latest
```
Then simply open `https://localhost:8443` in your browser.

## Quick Start

### 1. Requirements
* `go` (1.21+)
* `openssl`
* macOS or Linux environment

### 2. Setup the PKI & Generate Payloads
Navigate into the `demo` directory and run the fully automated setup script. This will compile the CLI, initialize the CA, issue a batch containing our `localhost` certificate, and extract the generated payloads.

```bash
chmod +x setup.sh
./setup.sh
```

During setup, you will see a dump of the decoded diagnostic structures printed directly to your terminal.

### 3. Launch the Demo Web Server
Start the Go web server to serve the beautiful UI and diagnostic APIs.

```bash
go run main.go
```

The server will bind to `8443` enforcing **TLS 1.3**.

### 4. Explore the Interface
Open your browser and navigate to:
**`https://localhost:8443`**

*(Note: You will receive a standard browser security warning because the server uses a self-signed X.509 certificate as a fallback, given that mainstream browsers do not natively support MTCs in the handshake just yet. Click `Advanced -> Proceed to localhost`.)*

Once loaded, click **"Verify Merkle Tree Certificate"** to perform a live inclusion proof verification, and use the **Diagnostics Buttons** to decode the various binary payloads.

## Project Structure

```text
/
├── Dockerfile           # Secure multi-stage Chainguard container
├── .github/workflows/   # Multi-arch CI pipeline via QEMU/Buildx
└── demo/
    ├── setup.sh             # Automates CA creation and certificate issuance
    ├── main.go              # The Go TLS 1.3 Web Server and verification API
    ├── static/              # The premium frontend UI
    └── screenshots/         # Screenshots for documentation
```

## Behind the Scenes (How Verification Works)
When you click the "Verify" button on the UI, the Go backend executes the reference verification logic:
```bash
mtc-cli verify -ca-params <ca-params> -validity-window <website.vw> <website.mtc>
```
If the certificate's authentication path successfully hashes up to the tree head checkpoint listed in the validity window, the certificate is implicitly trusted without requiring its own bulky PQC signature!
