## 🛡️ Cryptographic Bill of Materials (CBOM) & PQC Migration Assessment

**Format**: CycloneDX (v1.6) | **Total Components**: 4 | **Crypto Assets**: 4

### 📊 Post-Quantum Migration Scorecard

| Metric | Count | Migration Status |
|---|---|---|
| **Post-Quantum Ready (PQC)** | **1** | 🟢 Quantum-Resistant (NIST FIPS 203/204/205) |
| **Quantum-Vulnerable (Backlog)** | **2** | 🔴 At Risk of 'Harvest Now, Decrypt Later' |
| **Classical Symmetric / Hashing** | **1** | 🟡 Classical Security (Requires AES-256 / SHA-256+) |
| **Asymmetric PQC Migration Progress** | **33.3%** | (1 of 3 asymmetric primitives migrated) |

### ✅ Post-Quantum Cryptography Migrated Assets

| Component Name | Primitive | Key/Parameter Set | PQC Standard | Location(s) |
|---|---|---|---|---|
| `ML-DSA-65` | signature | N/A | NIST FIPS 204 (ML-DSA) | `playground/main.go:324`<br>`playground/main.go:332` |

### ⚠️ Quantum-Vulnerable Assets (Action Required)

| Component / Asset Name | Asset Type | Primitive / Algorithm | Key Length / Curve | Recommended PQC Replacement | Location(s) |
|---|---|---|---|---|---|
| `Ed25519` | algorithm | Ed25519 | 25519 | **ML-DSA-65 / Dilithium (FIPS 204)** | `playground/main.go:324`<br>`playground/main.go:332` |
| `ECDSA-P256` | algorithm | ECDSA-P256 | secp256r1 | **ML-DSA-65 / Dilithium (FIPS 204)** | `playground/main.go:332` |

### 🛠️ Developer & Security Remediation Guide

The following source code locations require cryptographic migration before quantum computing milestones:

#### `Ed25519` (signature)
- **Current Algorithm**: `Ed25519` (Key/Curve: `25519`)
- **Recommended Target**: **ML-DSA-65 / Dilithium (NIST FIPS 204)**
- **Call Sites / Instantiations**:
  - `playground/main.go:324`:
    ```
    "verification": "signatureless (landmark root hash) or signed (Ed25519 / ML-DSA cosigners)",
    ```
  - `playground/main.go:332`:
    ```
    "algorithms": []string{"ECDSA P-256 (key generation)", "SHA-256 (Merkle hashing)", "Ed25519 (cosigner)", "ML-DSA-44/65/8
    ```

#### `ECDSA-P256` (signature)
- **Current Algorithm**: `ECDSA-P256` (Key/Curve: `secp256r1`)
- **Recommended Target**: **ML-DSA-65 / Dilithium (NIST FIPS 204)**
- **Call Sites / Instantiations**:
  - `playground/main.go:332`:
    ```
    "algorithms": []string{"ECDSA P-256 (key generation)", "SHA-256 (Merkle hashing)", "Ed25519 (cosigner)", "ML-DSA-44/65/8
    ```

### 🔒 Classical Symmetric & Digest Assets

| Component Name | Primitive | Key Length | Quantum Resistance Assessment | Location(s) |
|---|---|---|---|---|
| `SHA-256` | hash | 256 | Quantum-Resistant (Grover's proof) | `playground/main.go:332`<br>`playground/main.go:334`<br>`playground/main.go:335` |
