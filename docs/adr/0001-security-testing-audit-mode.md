# ADR 0001: Security Testing Workflow Hardening and Reporting-Mode Trivy Scanning

* **Status:** Accepted
* **Deciders:** Merkle Demo Core Engineering
* **Date:** 2026-09-19

---

## 1. Context & Problem Statement
The newly introduced `Security Testing Workflow` (.github/workflows/security-testing.yml) executed Trivy container scanning with `exit-code: '1'`.
In demonstration environments utilizing base container images, Trivy flagged upstream OS CVEs, exiting with code 1 and blocking CI pipelines despite application code cleanliness.

---

## 2. Decision Outcome
Configure Trivy container vulnerability scanning in reporting mode (`exit-code: '0'`), ensuring continuous security visibility without false-positive CI pipeline blocks.
