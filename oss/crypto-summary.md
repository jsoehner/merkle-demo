## 🛡️ Cryptographic Bill of Materials (CBOM) & PQC Migration Assessment

**Format**: CycloneDX (v1.6) | **First-Party Code Crypto Assets**: 8 | **Total Tracked Crypto Assets**: 8

### 📊 Post-Quantum Migration Scorecard

| Metric | Count | Migration Status |
|---|---|---|
| **Post-Quantum Ready (PQC)** | **1** | 🟢 Quantum-Resistant (NIST FIPS 203/204/205) |
| **Quantum-Vulnerable (Backlog)** | **3** | 🔴 At Risk of 'Harvest Now, Decrypt Later' |
| **Classical Symmetric / Digest** | **4** | 🟡 Classical Security (Requires AES-256 / SHA-256+) |
| **Asymmetric PQC Migration Progress** | **25.0%** | (1 of 4 asymmetric primitives migrated) |

### 🎯 Cryptographic Supply Chain Coverage & Confidence

| Evaluation Layer | Coverage / Status | Audit Confidence Assessment |
|---|---|---|
| **First-Party Code (`src/`)** | **100% Audited** (0 Custom Primitives) | 🟢 **HIGH** (Direct AST & SAST verified clean) |
| **Third-Party Supply Chain** | **No external dependencies in SBOM** | 🟢 **HIGH** (Self-contained scope) |

| **System OS Inventory** | **Pending Update** | 🟡 **Warning** (Local OS crypto components not yet inventoried) |

### ✅ Post-Quantum Ready Assets

| Component Name | Primitive | Key/Parameter Set | PQC Standard | Provenance / Location(s) |
|---|---|---|---|---|
| `ML-DSA-65` | signature | N/A | ML-DSA-65 (FIPS 204) | First-Party Code (SAST/AST)<br>`playground/main.go:324`<br>`playground/main.go:332`<br>`mtc/signature.go:14`<br>`mtc/signature.go:77`<br>`mtc/signature.go:80`<br>`mtc/signature.go:81`<br>`mtc/signature.go:86`<br>`mtc/signature.go:90`<br>`mtc/signature.go:151`<br>`mtc/signature.go:153`<br>`mtc/signature.go:199`<br>`mtc/signature.go:200`<br>`mtc/signature.go:202`<br>`mtc/signature.go:203`<br>`mtc/signature.go:220`<br>`mtc/signature.go:223`<br>`mtc/signature.go:224`<br>`mtc/signature.go:229`<br>`mtc/signature.go:230`<br>`mtc/signature.go:247`<br>`mtc/signature.go:248`<br>`mtc/signature.go:250`<br>`mtc/signature.go:251`<br>`mtc/signature.go:269`<br>`mtc/signature.go:296`<br>`mtc/signature.go:315`<br>`mtc/signature.go:344` |

### ⚠️ Quantum-Vulnerable Assets & Remediation Plan

| Component / Algorithm | Type / Primitive | Key Length / Curve | Recommended Target | Provenance / Context |
|---|---|---|---|---|
| **`Ed25519`**<br><sub>Ed25519</sub> | algorithm / signature | 25519 | **ML-DSA-65 / Dilithium (FIPS 204)** | First-Party Code (SAST/AST)<br>`playground/main.go:324`<br><sub><code>"verification": "signatureless (landmark root hash) or signed (Ed25519 / ML-DSA cosigners)",</code></sub><br><br>`playground/main.go:332`<br><sub><code>"algorithms": []string{"ECDSA P-256 (key generation)", "SHA-256 (Merkle hashing)", "Ed25519 (cosigner)", "ML-DSA-44/65/8</code></sub><br><br>`mtc/mtc_test.go:5`<br><sub><code>"crypto/ed25519"</code></sub><br><br>`mtc/mtc_test.go:59`<br><sub><code>var seed [ed25519.SeedSize]byte</code></sub><br><br>`mtc/mtc_test.go:65`<br><sub><code>privEd := ed25519.NewKeyFromSeed(seed[:])</code></sub><br><br>`mtc/signature.go:6`<br><sub><code>"crypto/ed25519"</code></sub><br><br>`mtc/signature.go:42`<br><sub><code>type ed25519Verifier ed25519.PublicKey</code></sub><br><br>`mtc/signature.go:45`<br><sub><code>ret := make([]byte, ed25519.PublicKeySize)</code></sub><br><br>`mtc/signature.go:51`<br><sub><code>if ed25519.Verify((ed25519.PublicKey)(v), msg, sig) {</code></sub><br><br>`mtc/signature.go:54`<br><sub><code>return errors.New("ed25519 verification failed")</code></sub><br><br>`mtc/signature.go:135`<br><sub><code>epk, ok := pk.(ed25519.PublicKey)</code></sub><br><br>`mtc/signature.go:136`<br><sub><code>if !ok \|\| len(epk) != ed25519.PublicKeySize {</code></sub><br><br>`mtc/signature.go:137`<br><sub><code>return nil, errors.New("Expected ed25519.PublicKey")</code></sub><br><br>`mtc/signature.go:176`<br><sub><code>if len(data) != ed25519.PublicKeySize {</code></sub><br><br>`mtc/signature.go:177`<br><sub><code>return nil, errors.New("Wrong length for ed25519 public key")</code></sub><br><br>`mtc/signature.go:179`<br><sub><code>ret := make([]byte, ed25519.PublicKeySize)</code></sub><br><br>`mtc/signature.go:294`<br><sub><code>return "ed25519"</code></sub><br><br>`mtc/signature.go:317`<br><sub><code>case "ed25519":</code></sub><br><br>`mtc/signature.go:342`<br><sub><code>case ed25519.PublicKey:</code></sub> |
| **`ECDSA-P256`**<br><sub>ECDSA-P256</sub> | algorithm / signature | secp256r1 | **ML-DSA-65 / Dilithium (FIPS 204)** | First-Party Code (SAST/AST)<br>`playground/main.go:332`<br><sub><code>"algorithms": []string{"ECDSA P-256 (key generation)", "SHA-256 (Merkle hashing)", "Ed25519 (cosigner)", "ML-DSA-44/65/8</code></sub><br><br>`mtc/signature.go:5`<br><sub><code>"crypto/ecdsa"</code></sub><br><br>`mtc/signature.go:58`<br><sub><code>pk     *ecdsa.PublicKey</code></sub><br><br>`mtc/signature.go:71`<br><sub><code>if ecdsa.VerifyASN1(v.pk, hashed, sig) {</code></sub><br><br>`mtc/signature.go:74`<br><sub><code>return errors.New("ecdsa verification failed")</code></sub><br><br>`mtc/signature.go:141`<br><sub><code>epk, ok := pk.(*ecdsa.PublicKey)</code></sub><br><br>`mtc/signature.go:143`<br><sub><code>return nil, errors.New("Expected *ecdsa.PublicKey")</code></sub><br><br>`mtc/signature.go:186`<br><sub><code>return nil, errors.New("Failed to unmarshal ecdsa public key")</code></sub><br><br>`mtc/signature.go:190`<br><sub><code>pk: &ecdsa.PublicKey{</code></sub><br><br>`mtc/signature.go:332`<br><sub><code>case *ecdsa.PublicKey:</code></sub><br><br>`mtc/signature.go:334`<br><sub><code>case "P-256":</code></sub><br><br>`mtc/ca/ca_test.go:4`<br><sub><code>"crypto/ecdsa"</code></sub><br><br>`mtc/ca/ca_test.go:71`<br><sub><code>sk, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)</code></sub> |
| **`RSA-2048`**<br><sub>RSA-2048</sub> | algorithm / signature | 2048 | **ML-KEM-768 / Kyber (FIPS 203)** | First-Party Code (SAST/AST)<br>`mtc/mtc_test.go:6`<br><sub><code>"crypto/rsa"</code></sub><br><br>`mtc/mtc_test.go:211`<br><sub><code>pubRSA := &rsa.PublicKey{</code></sub><br><br>`mtc/signature.go:8`<br><sub><code>"crypto/rsa"</code></sub><br><br>`mtc/signature.go:25`<br><sub><code>pk     *rsa.PublicKey</code></sub><br><br>`mtc/signature.go:38`<br><sub><code>signOpts := &rsa.PSSOptions{SaltLength: rsa.PSSSaltLengthEqualsHash}</code></sub><br><br>`mtc/signature.go:39`<br><sub><code>return rsa.VerifyPSS(v.pk, v.hash, hashed, sig, signOpts)</code></sub><br><br>`mtc/signature.go:128`<br><sub><code>rpk, ok := pk.(*rsa.PublicKey)</code></sub><br><br>`mtc/signature.go:130`<br><sub><code>return nil, errors.New("Expected *rsa.PublicKey")</code></sub><br><br>`mtc/signature.go:282`<br><sub><code>return "rsa-sha256"</code></sub><br><br>`mtc/signature.go:284`<br><sub><code>return "rsa-sha384"</code></sub><br><br>`mtc/signature.go:286`<br><sub><code>return "rsa-sha512"</code></sub><br><br>`mtc/signature.go:303`<br><sub><code>case "rsa-sha256":</code></sub><br><br>`mtc/signature.go:305`<br><sub><code>case "rsa-sha384":</code></sub><br><br>`mtc/signature.go:307`<br><sub><code>case "rsa-sha512":</code></sub><br><br>`mtc/signature.go:326`<br><sub><code>case *rsa.PublicKey:</code></sub> |

### 🔒 Classical Symmetric & Digest Assets

| Component Name | Primitive | Key Length | Quantum Resistance Assessment | Provenance / Location(s) |
|---|---|---|---|---|
| `SHA-256` | hash | 256 | Quantum-Resistant (Grover's proof) | First-Party Code (SAST/AST)<br>`playground/main.go:332`<br>`playground/main.go:334`<br>`playground/main.go:335`<br>`mtc/mtc.go:6`<br>`mtc/mtc.go:934`<br>`mtc/mtc.go:1171`<br>`mtc/mtc.go:1224`<br>`mtc/mtc.go:1380`<br>`mtc/mtc.go:1405`<br>`mtc/mtc.go:1534`<br>`mtc/mtc.go:1563`<br>`mtc/signature.go:9`<br>`mtc/signature.go:96`<br>`mtc/signature.go:282`<br>`mtc/signature.go:303`<br>`mtc/signature.go:353`<br>`mtc/ca/ca.go:6`<br>`mtc/ca/ca.go:651`<br>`mtc/ca/ca.go:677`<br>`mtc/umbilical/frozencas/frozencas.go:42`<br>`mtc/umbilical/frozencas/frozencas.go:271` |
| `SHA-3` | hash | 256 | Quantum-Resistant (Grover's proof) | First-Party Code (SAST/AST)<br>`mtc/mtc_test.go:16`<br>`mtc/mtc_test.go:61` |
| `SHA-384` | hash | 384 | Quantum-Resistant (Grover's proof) | First-Party Code (SAST/AST)<br>`mtc/signature.go:98`<br>`mtc/signature.go:284`<br>`mtc/signature.go:305` |
| `SHA-512` | hash | 512 | Quantum-Resistant (Grover's proof) | First-Party Code (SAST/AST)<br>`mtc/signature.go:100`<br>`mtc/signature.go:286`<br>`mtc/signature.go:307` |
