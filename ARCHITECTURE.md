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

## How to Run the Demo

To launch the web server, simply navigate to the `demo/` directory and run:

```bash
cd demo
go run main.go
```

Then, open your browser and navigate to `https://localhost:8443` (accept the self-signed X.509 warning, which acts as the TLS fallback). 
Click the **Verify Merkle Tree Certificate** button to execute a live backend verification of the Inclusion Proof against the CA parameters.
