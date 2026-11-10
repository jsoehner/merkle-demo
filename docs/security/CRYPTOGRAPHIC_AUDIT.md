## 🛡️ Cryptographic Bill of Materials (CBOM) & PQC Migration Assessment

**Format**: CycloneDX (v1.6) | **First-Party Code Crypto Assets**: 4 | **Total Tracked Crypto Assets**: 4

### 📊 Post-Quantum Migration Scorecard

| Metric | Count | Migration Status |
|---|---|---|
| **Post-Quantum Ready (PQC)** | **1** | 🟢 Quantum-Resistant (NIST FIPS 203/204/205) |
| **Quantum-Vulnerable (Backlog)** | **2** | 🔴 At Risk of 'Harvest Now, Decrypt Later' |
| **Classical Symmetric / Digest** | **1** | 🟡 Classical Security (Requires AES-256 / SHA-256+) |
| **Asymmetric PQC Migration Progress** | **33.3%** | (1 of 3 asymmetric primitives migrated) |

### 🎯 Cryptographic Supply Chain Coverage & Confidence

| Evaluation Layer | Coverage / Status | Audit Confidence Assessment |
|---|---|---|
| **First-Party Code (`src/`, `scripts/`)** | **100% Audited** (4 Crypto Primitive(s) Inventoried) | 🟢 **HIGH** (Direct AST & SAST verified) |
| **Third-Party Supply Chain** | **25.0%** (23 of 92 dependencies cataloged) | 🔴 LOW (Known profiles assimilated) |
| **Overall Audit Confidence Score** | **25.8%** | 🔴 LOW (69 unassimilated supply chain dependencies) |

| **System OS Inventory** | **Pending Update** | 🟡 **Warning** (Local OS crypto components not yet inventoried) |

### ✅ Post-Quantum Ready Assets

| Component Name | Primitive | Key/Parameter Set | PQC Standard | Provenance / Location(s) |
|---|---|---|---|---|
| `ML-DSA-65` | signature | N/A | ML-DSA-65 (FIPS 204) | First-Party Code (SAST/AST)<br>`playground/main.go:332` |

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
| `demo` | UNKNOWN | `pkg:golang/demo` | 🟡 Unassimilated (No upstream CBOM) |
| `demo` | v0.0.0-20260512205513-da76a94bef6b | `pkg:golang/demo@v0.0.0-20260512205513-da76a94bef6b` | 🟡 Unassimilated (No upstream CBOM) |
| `github.com/bwesterb/mtc` | v0.1.3-0.20251105140709-36baf0cc07ae | `pkg:golang/github.com/bwesterb/mtc@v0.1.3-0.20251105140709-36baf0cc07ae` | 🟡 Unassimilated (No upstream CBOM) |
| `github.com/bwesterb/mtc` | v0.1.3-0.20251105140709-36baf0cc07ae | `pkg:golang/github.com/bwesterb/mtc@v0.1.3-0.20251105140709-36baf0cc07ae` | 🟡 Unassimilated (No upstream CBOM) |
| `github.com/bwesterb/mtc` | v0.1.3-0.20251105140709-36baf0cc07ae | `pkg:golang/github.com/bwesterb/mtc@v0.1.3-0.20251105140709-36baf0cc07ae` | 🟡 Unassimilated (No upstream CBOM) |
| `github.com/bwesterb/mtc` | v0.1.3-0.20251105140709-36baf0cc07ae | `pkg:golang/github.com/bwesterb/mtc@v0.1.3-0.20251105140709-36baf0cc07ae` | 🟡 Unassimilated (No upstream CBOM) |
| `github.com/bwesterb/mtc` | v0.1.3-0.20251105140709-36baf0cc07ae | `pkg:golang/github.com/bwesterb/mtc@v0.1.3-0.20251105140709-36baf0cc07ae` | 🟡 Unassimilated (No upstream CBOM) |
| `github.com/cloudflare/circl` | v1.6.0 | `pkg:golang/github.com/cloudflare/circl@v1.6.0` | 🟡 Unassimilated (No upstream CBOM) |
| `github.com/cloudflare/circl` | v1.6.0 | `pkg:golang/github.com/cloudflare/circl@v1.6.0` | 🟡 Unassimilated (No upstream CBOM) |
| `github.com/cloudflare/circl` | v1.6.0 | `pkg:golang/github.com/cloudflare/circl@v1.6.0` | 🟡 Unassimilated (No upstream CBOM) |
| `github.com/cloudflare/circl` | v1.6.0 | `pkg:golang/github.com/cloudflare/circl@v1.6.0` | 🟡 Unassimilated (No upstream CBOM) |
| `github.com/cloudflare/circl` | v1.6.0 | `pkg:golang/github.com/cloudflare/circl@v1.6.0` | 🟡 Unassimilated (No upstream CBOM) |
| `github.com/cpuguy83/go-md2man/v2` | v2.0.2 | `pkg:golang/github.com/cpuguy83/go-md2man/v2@v2.0.2` | 🟡 Unassimilated (No upstream CBOM) |
| `github.com/cpuguy83/go-md2man/v2` | v2.0.2 | `pkg:golang/github.com/cpuguy83/go-md2man/v2@v2.0.2` | 🟡 Unassimilated (No upstream CBOM) |
| `github.com/cpuguy83/go-md2man/v2` | v2.0.2 | `pkg:golang/github.com/cpuguy83/go-md2man/v2@v2.0.2` | 🟡 Unassimilated (No upstream CBOM) |
| *... and 54 more unassimilated dependencies* | | | |
