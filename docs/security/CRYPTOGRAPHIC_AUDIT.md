## 🛡️ Cryptographic Bill of Materials (CBOM) & PQC Migration Assessment

**Format**: CycloneDX (v1.6) | **First-Party Code Crypto Assets**: 4 | **Total Tracked Crypto Assets**: 4

### 📊 Post-Quantum Migration Scorecard

| Metric | Count | Migration Status |
|---|---|---|
| **Post-Quantum Ready (PQC)** | **1** | 🟢 Quantum-Resistant (NIST FIPS 203/204/205) |
| **Quantum-Vulnerable (Backlog)** | **2** | 🔴 At Risk of 'Harvest Now, Decrypt Later' |
| **Classical Symmetric / Hashing** | **1** | 🟡 Classical Security (Requires AES-256 / SHA-256+) |
| **Asymmetric PQC Migration Progress** | **33.3%** | (1 of 3 asymmetric primitives migrated) |

### 🎯 Cryptographic Supply Chain Coverage & Confidence

| Evaluation Layer | Coverage / Status | Audit Confidence Assessment |
|---|---|---|
| **First-Party Code (`src/`)** | **100% Audited** (0 Custom Primitives) | 🟢 **HIGH** (Direct AST & SAST verified clean) |
| **Third-Party Supply Chain** | **0.0%** (0 of 105 dependencies cataloged) | 🔴 LOW (Known profiles assimilated) |
| **Overall Audit Confidence Score** | **0.9%** | **🔴 LOW** (105 unassimilated supply chain dependencies) |

### ✅ Post-Quantum Cryptography Migrated Assets

| Component Name | Primitive | Key/Parameter Set | PQC Standard | Provenance / Location(s) |
|---|---|---|---|---|
| `ML-DSA-65` | signature | N/A | NIST FIPS 204 (ML-DSA) | First-Party Code (SAST/AST)<br>`playground/main.go:332` |

### ⚠️ Quantum-Vulnerable Assets & Remediation Plan

| Component / Algorithm | Type / Primitive | Key Length / Curve | Recommended Target | Provenance / Context |
|---|---|---|---|---|
| **`ECDSA-P256`**<br><sub>ECDSA-P256</sub> | algorithm / signature | secp256r1 | **ML-DSA-65 / Dilithium (FIPS 204)** | First-Party Code (SAST/AST)<br>`playground/main.go:332`<br><sub><code>"algorithms": []string{"ECDSA P-256 (key generation)", "SHA-256 (Merkle hashing)", "Ed25519 (cosigner)", "ML-DSA-44/65/8</code></sub> |
| **`Ed25519`**<br><sub>Ed25519</sub> | algorithm / signature | 25519 | **ML-DSA-65 / Dilithium (FIPS 204)** | First-Party Code (SAST/AST)<br>`playground/main.go:332`<br><sub><code>"algorithms": []string{"ECDSA P-256 (key generation)", "SHA-256 (Merkle hashing)", "Ed25519 (cosigner)", "ML-DSA-44/65/8</code></sub> |

### 🔒 Classical Symmetric & Digest Assets

| Component Name | Primitive | Key Length | Quantum Resistance Assessment | Provenance / Location(s) |
|---|---|---|---|---|
| `SHA-256` | hash | 256 | Quantum-Resistant (Grover's proof) | First-Party Code (SAST/AST)<br>`playground/main.go:332` |

### ⚠️ Unassimilated Third-Party Binaries & Cryptographic Blind Spots

> ℹ️ *The following third-party dependencies do not have verified upstream CBOM attestations in the catalog. They lower the audit confidence score until explicit CBOMs or attestations are published.* 

| Dependency Name | Version | Package URL (purl) | Status |
|---|---|---|---|
| `actions/checkout` | v4 | `pkg:github/actions/checkout@v4` | 🟡 Unassimilated (No upstream CBOM) |
| `actions/checkout` | v7 | `pkg:github/actions/checkout@v7` | 🟡 Unassimilated (No upstream CBOM) |
| `actions/checkout` | v7 | `pkg:github/actions/checkout@v7` | 🟡 Unassimilated (No upstream CBOM) |
| `actions/checkout` | v7 | `pkg:github/actions/checkout@v7` | 🟡 Unassimilated (No upstream CBOM) |
| `actions/checkout` | v7.0.1 | `pkg:github/actions/checkout@v7.0.1` | 🟡 Unassimilated (No upstream CBOM) |
| `actions/setup-go` | v5 | `pkg:github/actions/setup-go@v5` | 🟡 Unassimilated (No upstream CBOM) |
| `actions/setup-python` | v5.6.0 | `pkg:github/actions/setup-python@v5.6.0` | 🟡 Unassimilated (No upstream CBOM) |
| `actions/upload-artifact` | v4.6.2 | `pkg:github/actions/upload-artifact@v4.6.2` | 🟡 Unassimilated (No upstream CBOM) |
| `anchore/sbom-action` | v0.24.2 | `pkg:github/anchore/sbom-action@v0.24.2` | 🟡 Unassimilated (No upstream CBOM) |
| `aquasecurity/trivy-action` | master | `pkg:github/aquasecurity/trivy-action@master` | 🟡 Unassimilated (No upstream CBOM) |
| `cbomkit/cbomkit-action` | v2.3.0 | `pkg:github/cbomkit/cbomkit-action@v2.3.0` | 🟡 Unassimilated (No upstream CBOM) |
| `demo` | UNKNOWN | `pkg:golang/demo` | 🟡 Unassimilated (No upstream CBOM) |
| `demo` | v0.0.0-20260512205513-da76a94bef6b | `pkg:golang/demo@v0.0.0-20260512205513-da76a94bef6b` | 🟡 Unassimilated (No upstream CBOM) |
| `docker/build-push-action` | v5 | `pkg:github/docker/build-push-action@v5` | 🟡 Unassimilated (No upstream CBOM) |
| `docker/build-push-action` | v6 | `pkg:github/docker/build-push-action@v6` | 🟡 Unassimilated (No upstream CBOM) |
| *... and 90 more unassimilated dependencies* | | | |
