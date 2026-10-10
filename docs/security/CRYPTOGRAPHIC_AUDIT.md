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

### ⚠️ Quantum-Vulnerable Assets & Remediation Plan

| Component / Algorithm | Type / Primitive | Key Length / Curve | Recommended Target | Source Location(s) & Code Context |
|---|---|---|---|---|
| **`Ed25519`**<br><sub>Ed25519</sub> | algorithm / signature | 25519 | **ML-DSA-65 / Dilithium (FIPS 204)** | `playground/main.go:324`<br><sub><code>"verification": "signatureless (landmark root hash) or signed (Ed25519 / ML-DSA cosigners)",</code></sub><br><br>`playground/main.go:332`<br><sub><code>"algorithms": []string{"ECDSA P-256 (key generation)", "SHA-256 (Merkle hashing)", "Ed25519 (cosigner)", "ML-DSA-44/65/8</code></sub> |
| **`ECDSA-P256`**<br><sub>ECDSA-P256</sub> | algorithm / signature | secp256r1 | **ML-DSA-65 / Dilithium (FIPS 204)** | `playground/main.go:332`<br><sub><code>"algorithms": []string{"ECDSA P-256 (key generation)", "SHA-256 (Merkle hashing)", "Ed25519 (cosigner)", "ML-DSA-44/65/8</code></sub> |

### 🔒 Classical Symmetric & Digest Assets

| Component Name | Primitive | Key Length | Quantum Resistance Assessment | Location(s) |
|---|---|---|---|---|
| `SHA-256` | hash | 256 | Quantum-Resistant (Grover's proof) | `playground/main.go:332`<br>`playground/main.go:334`<br>`playground/main.go:335` |
