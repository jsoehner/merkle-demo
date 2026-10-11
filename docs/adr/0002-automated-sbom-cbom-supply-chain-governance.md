# ADR-0002: Automated SBOM and CBOM Supply Chain Governance Pipeline

## Metadata

| Field | Value |
|---|---|
| ADR ID | ADR-0002 |
| Status | Accepted |
| Created Date | 2026-10-11 |
| Decision Date | 2026-10-11 |
| Effective Date | 2026-10-11 |
| Architecture Domain | Security / Supply Chain Governance / Post-Quantum Cryptography |
| Scope | Software & Cryptographic Supply Chain Security (Repository & CI/CD) |
| Decision Owner | Lead DevSecOps Architect |
| Named Decision Maker | Merkle Demo Core Engineering & Security Architecture Team |
| Approval Authority | Architecture Review Board (ARB) / Security Governance Committee |
| Approval Evidence ID | ARB-2026-10-SEC-MTC-002 |
| APM Code(s) | SEC-SUPPLYCHAIN-02 |
| NIRA Reference(s) | NIRA-SEC-2026-043 |
| Related Records | [ADR-0001](0001-security-testing-audit-mode.md) |
| Tags | sbom, cbom, cyclonedx, spdx, supply-chain, pqc, ast-analysis, devsecops, mtc |

---

## Business and Architecture Context

### Problem Statement

The Merkle Tree Certificates (MTC) demonstration project operates at the leading edge of Post-Quantum Cryptography (PQC) and Internet Engineering Task Force (IETF) PKI architectures. To ensure production trustworthiness, zero-CVE posture, and alignment with federal directives (Executive Order 14028, NIST SP 800-218 SSDF, and NIST IR 8547 PQC transition guidelines), the codebase requires continuous, automated visibility into both third-party software dependencies and cryptographic assets.

Without automated, repeatable generation and verification embedded into Continuous Integration (CI):
1. Third-party software dependencies cannot be audited against open-source vulnerability databases via standard machine-readable formats.
2. Cryptographic algorithms, certificates, and keys implemented across the application and container layers lack structured inventory, preventing accurate assessment of PQC migration readiness.
3. First-party source code cryptographic call sites risk diverging from external bills of materials without Abstract Syntax Tree (AST) semantic reconciliation.

### Current Situation

Prior to this decision, the repository maintained Docker multi-arch publishing (`.github/workflows/docker-publish.yml`) and Trivy vulnerability auditing (`.github/workflows/security-testing.yml` configured in reporting mode per ADR-0001). However:
1. No automated CI pipeline existed to generate and publish standardized Software Bills of Materials (SBOMs) in both CycloneDX and SPDX formats.
2. No Cryptographic Bill of Materials (CBOM) pipeline existed to enumerate cryptographic primitives, keys, certificates, or protocol configurations.
3. Cryptographic call sites in Go source files were not validated against CBOM inventories.
4. No automated verification test harness guarded against corrupt or empty BOM artifacts.

### Decision Drivers

- **Cryptographic & Post-Quantum Posture:** Comprehensive automated visibility into cryptographic primitives, public-key algorithms, and quantum-vulnerable assets (e.g., classical RSA/ECC vs. post-quantum hash-based Merkle structures).
- **Compliance & Interoperability:** Support for both SPDX (v2.3) and CycloneDX (v1.5/1.6) formats to satisfy enterprise procurement and compliance specifications (EO 14028, NTIA minimum elements).
- **Dual-Engine Source & Artifact Reconciliation:** Combining container/filesystem scanning with dual-engine AST semantic scanning (`scan_crypto_ast.py`) to reconcile internal call sites with declared components.
- **Continuous CI Verification & Quality Gates:** Automated test harness (`test_boms.sh`) running in CI to enforce valid JSON schema, required CycloneDX attributes, and non-empty asset lists.
- **Developer Experience Parity:** Portable local execution scripts (`scripts/`) allowing engineers to generate, analyze, and test BOMs locally prior to pushing code.

---

## Assumptions

- The repository codebase contains Go source code, multi-stage Dockerfiles, and containerized runtime environments.
- Continuous Integration runners support Python 3.11+ and GitHub Actions runner environments executing Node 24 compliant actions.
- PQC migration tracking requires active inventory of cryptographic call sites and protocol parameters over the transition horizon.

---

## Constraints

- GitHub Actions execution limits: CI workflows must run within standard runner resource constraints without adding prohibitive build latency (< 2 minutes).
- Zero external paid SaaS dependencies: Tooling must execute locally and in CI via open-source tools (Syft, cbomkit, Python AST engine) without external proprietary scanners.
- Security immutability: All workflow action steps must adhere to immutable commit SHA pinning on Node 24 runtimes.

---

## Alternatives Considered

| Option | Advantages | Disadvantages / Risks | Disposition |
|---|---|---|---|
| **Option 1: Status Quo (Trivy Container Scanning Only)** | No additional CI steps; zero pipeline latency overhead. | Zero machine-readable SBOM or CBOM artifacts; no PQC algorithm tracking; fails EO 14028 requirements. | **Rejected** |
| **Option 2: Single-Tool SBOM Generation (Syft Only)** | Simple workflow; generates CycloneDX and SPDX SBOMs quickly. | Completely ignores cryptographic assets; no CBOM generation; no AST call-site tracking; no PQC migration metrics. | **Rejected** |
| **Option 3: cbomkit-only Generation** | Generates CycloneDX CBOM for cryptographic algorithms. | Does not produce SPDX SBOM required by enterprise governance; lacks source code AST call-site reconciliation. | **Rejected** |
| **Option 4: Integrated Dual-Engine SBOM/CBOM + AST Analysis + Automated CI Harness (Selected)** | Comprehensive coverage: dual-engine SBOM (CycloneDX & SPDX via Anchore/Syft) + CBOM (cbomkit); deep AST semantic crypto call-site discovery (`scan_crypto_ast.py`); automated schema and gate verification (`test_boms.sh`); automatic PQC readiness dashboard generation (`analyze_cbom.py`); fully mirrored local scripts. | Modest CI runtime addition (~60-90s); requires Python runtime setup in CI runner; generates artifacts that must be archived. | **Selected** |

---

## Decision

### Selected Option

Adopt **Option 4: Integrated Dual-Engine SBOM/CBOM + AST Analysis + Automated CI Harness** across the repository.

Specifically:
1. **GitHub Actions Workflow (`.github/workflows/sbom-cbom.yml`):**
   - Configured with Node 24 pinned actions (`actions/checkout@v7.0.1`, `anchore/sbom-action@v0.24.2`, `cbomkit/cbomkit-action@v2.3.0`, `actions/setup-python@v5.6.0`, `actions/upload-artifact@v4.6.2`).
   - Generates both CycloneDX JSON (`oss/sbom.cyclonedx.json`) and SPDX JSON (`oss/sbom.spdx.json`) SBOMs via Anchore / Syft.
   - Generates CycloneDX CBOM (`oss/cbom.json`) via `cbomkit`.
   - Executes semantic AST cryptographic analysis via `scripts/scan_crypto_ast.py` to reconcile source code call sites into the CBOM inventory.
   - Executes validation and test harness via `scripts/test_boms.sh`.
   - Assesses Post-Quantum Cryptography (PQC) readiness and outputs dashboard summaries via `scripts/analyze_cbom.py`.
   - Automatically publishes updated audit reports to `docs/security/CRYPTOGRAPHIC_AUDIT.md` on `main` branch pushes.
   - Archives all BOM artifacts as workflow artifacts with 90-day retention.
2. **Local Tooling Suite (`scripts/`):**
   - `scripts/generate_boms.sh`: Portable CLI script for local generation across source trees or Docker images.
   - `scripts/test_boms.sh`: Automated validation test harness asserting JSON syntax, CycloneDX schema conformance, and component counts.
   - `scripts/scan_crypto_ast.py`: Multi-language AST walker identifying cryptographic call-sites, key generation, and cipher instantiations.
   - `scripts/analyze_cbom.py`: PQC readiness analysis, algorithm categorization, and Markdown/JSON report generator.
   - `scripts/attach_oci_bom.sh` & `scripts/fetch_oci_cbom.py`: OCI artifact publishing and retrieval utilities.

### Rationale

This architecture provides comprehensive supply chain and cryptographic visibility tailored to the Merkle Tree Certificates demonstration. Merging dependency-level SBOM tracking with cryptographic-level CBOM asset inventory and source AST call-site mapping provides defense-in-depth assurance. Automated testing in CI guarantees that corrupt, incomplete, or drifting BOMs are caught immediately.

---

## Consequences

### Benefits

- **Continuous Compliance & Audit Readiness:** Automated generation of both CycloneDX and SPDX formats satisfying federal and industry supply chain mandates.
- **PQC Cryptographic Transparency:** Real-time classification of quantum-vulnerable primitives and hybrid/post-quantum mechanisms directly aligned with MTC project goals.
- **Automated Quality Gating:** The `test_boms.sh` verification harness prevents publishing invalid or empty bills of materials.
- **Developer Parity:** Developers can generate and audit BOMs locally with zero friction using identical tooling.

### Trade-offs and Negative Consequences

- **CI Pipeline Latency:** Adds approximately 60-90 seconds to workflow execution time.
- **Catalog Maintenance:** The static cryptographic catalog in `scripts/known_cbom_catalog.json` must be updated periodically as cryptographic standards evolve.
- **Artifact Storage:** Retention of BOM artifacts in GitHub Actions consumes storage quota (controlled with 90-day retention).

### New Constraints

- Repository CI runners must supply Python 3.11+ for AST scanning and analysis scripts.
- GitHub Actions token permissions require `contents: write` to synchronize `docs/security/CRYPTOGRAPHIC_AUDIT.md` on `main` branch pushes.

### Follow-on Decisions and Activities

- Establish automated pull request checks that alert when non-quantum-resistant primitives are newly introduced.
- Integrate OCI cosign / notation attestation signing for generated BOM artifacts in container release pipelines.

---

## Security, Privacy, and Risk Considerations

- **Integrity & Action Pinning:** All workflow dependencies use immutable SHA-256 commit pins to mitigate supply chain tampering of CI actions.
- **Confidentiality:** Scanning inspects public dependencies, configurations, and cryptographic primitives; no private keys, passwords, or secrets are exposed or embedded in generated BOM artifacts.
- **Traceability:** BOM artifacts link directly to commit SHAs and repository build metadata.

---

## Technical Debt Assessment

- **Catalog Sync:** The cryptographic catalog in `scripts/known_cbom_catalog.json` requires review alongside NIST PQC standardization updates.

---

## Review and Lifecycle

- **Review Schedule:** Annual or upon major NIST PQC milestone updates.
- **Supersedes:** None.
- **Superseded By:** None.
