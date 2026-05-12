package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

// ────────────────────────────────────────────────────────────────────────────
// Response types
// ────────────────────────────────────────────────────────────────────────────

type RunResult struct {
	Success  bool   `json:"success"`
	Output   string `json:"output"`
	Duration string `json:"duration_ms"`
	Mode     string `json:"mode,omitempty"`
}

type ParsedStep struct {
	Step   int    `json:"step"`
	Label  string `json:"label"`
	Detail string `json:"detail"`
	Pass   bool   `json:"pass"`
}

type DemoResult struct {
	Success  bool         `json:"success"`
	Output   string       `json:"output"`
	Duration string       `json:"duration_ms"`
	Mode     string       `json:"mode"`
	Steps    []ParsedStep `json:"steps"`
	Domain   string       `json:"domain"`
}

type VerifyResult struct {
	Success  bool         `json:"success"`
	Output   string       `json:"output"`
	Duration string       `json:"duration_ms"`
	Steps    []ParsedStep `json:"steps"`
	CertPath string       `json:"cert_path,omitempty"`
}

type HealthResponse struct {
	OK           bool   `json:"ok"`
	PlaygroundOK bool   `json:"playground_ok"`
	BinPath      string `json:"bin_path"`
	Timestamp    string `json:"timestamp"`
}

// ────────────────────────────────────────────────────────────────────────────
// Helpers
// ────────────────────────────────────────────────────────────────────────────

func binPath(name string) string {
	p := os.Getenv("PLAYGROUND_BIN_DIR")
	if p == "" {
		p = "/usr/local/bin"
	}
	return p + "/" + name
}

func corsHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
}

func jsonErr(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

// parseSteps extracts Step N: lines and [PASS]/[FAIL] lines from output.
func parseSteps(output string) []ParsedStep {
	lines := strings.Split(output, "\n")
	stepRe := regexp.MustCompile(`(?i)^step\s+(\d+)[:\.]?\s+(.+)`)
	passRe := regexp.MustCompile(`(?i)\[PASS\]\s*(.*)`)
	failRe := regexp.MustCompile(`(?i)\[FAIL\]\s*(.*)`)

	var steps []ParsedStep
	var current *ParsedStep
	var detailBuf strings.Builder

	flush := func() {
		if current != nil {
			current.Detail = strings.TrimSpace(detailBuf.String())
			steps = append(steps, *current)
			current = nil
			detailBuf.Reset()
		}
	}

	for _, line := range lines {
		if m := stepRe.FindStringSubmatch(strings.TrimSpace(line)); m != nil {
			flush()
			stepNum := 0
			fmt.Sscanf(m[1], "%d", &stepNum)
			current = &ParsedStep{Step: stepNum, Label: strings.TrimSpace(m[2]), Pass: true}
			continue
		}
		if m := passRe.FindStringSubmatch(strings.TrimSpace(line)); m != nil {
			detailBuf.WriteString("✅ " + strings.TrimSpace(m[1]) + "\n")
			continue
		}
		if m := failRe.FindStringSubmatch(strings.TrimSpace(line)); m != nil {
			if current != nil {
				current.Pass = false
			}
			detailBuf.WriteString("❌ " + strings.TrimSpace(m[1]) + "\n")
			continue
		}
		// Detail lines (indented)
		if strings.HasPrefix(line, "        ") || strings.HasPrefix(line, "\t\t") {
			detailBuf.WriteString(strings.TrimSpace(line) + "\n")
		}
	}
	flush()
	return steps
}

// ────────────────────────────────────────────────────────────────────────────
// Handlers
// ────────────────────────────────────────────────────────────────────────────

// GET /api/health
func healthHandler(w http.ResponseWriter, r *http.Request) {
	corsHeaders(w)
	binOK := false
	if _, err := os.Stat(binPath("demo-embedded-cert")); err == nil {
		binOK = true
	}
	resp := HealthResponse{
		OK:           true,
		PlaygroundOK: binOK,
		BinPath:      binPath("demo-embedded-cert"),
		Timestamp:    time.Now().UTC().Format(time.RFC3339),
	}
	json.NewEncoder(w).Encode(resp)
}

// POST /api/demo?mode=embedded|mtc&domain=<domain>
func demoHandler(w http.ResponseWriter, r *http.Request) {
	corsHeaders(w)

	mode := r.URL.Query().Get("mode")
	if mode == "" {
		mode = "mtc"
	}
	domain := r.URL.Query().Get("domain")
	if domain == "" {
		domain = "demo.example.com"
	}

	// Basic domain validation
	domainRe := regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9\-\.]{1,61}[a-zA-Z0-9]\.[a-zA-Z]{2,}$`)
	if !domainRe.MatchString(domain) {
		jsonErr(w, "Invalid domain name", http.StatusBadRequest)
		return
	}

	var args []string
	certMode := "MTC-Spec (id-alg-mtcProof)"

	switch mode {
	case "mtc":
		args = []string{"-mtc-mode", "-domain", domain}
	case "embedded":
		args = []string{"-domain", domain}
		certMode = "Legacy Embedded Proof (X.509 extension)"
	default:
		jsonErr(w, "mode must be 'mtc' or 'embedded'", http.StatusBadRequest)
		return
	}

	start := time.Now()
	cmd := exec.Command(binPath("demo-embedded-cert"), args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	// stdout is the PEM (we don't need it for the dashboard)
	cmd.Stdout = &bytes.Buffer{}

	err := cmd.Run()
	elapsed := time.Since(start)

	output := stderr.String()
	success := err == nil

	steps := parseSteps(output)
	if len(steps) == 0 && output != "" {
		// Fallback: treat entire output as a single step
		steps = []ParsedStep{{Step: 1, Label: "Run Demo", Detail: output, Pass: success}}
	}

	resp := DemoResult{
		Success:  success,
		Output:   output,
		Duration: fmt.Sprintf("%d", elapsed.Milliseconds()),
		Mode:     certMode,
		Steps:    steps,
		Domain:   domain,
	}
	json.NewEncoder(w).Encode(resp)
}

// POST /api/verify?domain=<domain>&cert_mode=embedded|mtc
// Generates a cert into a temp file, then runs mtc-verify-cert on it.
func verifyHandler(w http.ResponseWriter, r *http.Request) {
	corsHeaders(w)

	domain := r.URL.Query().Get("domain")
	if domain == "" {
		domain = "demo.example.com"
	}
	certMode := r.URL.Query().Get("cert_mode")
	if certMode == "" {
		certMode = "mtc"
	}

	domainRe := regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9\-\.]{1,61}[a-zA-Z0-9]\.[a-zA-Z]{2,}$`)
	if !domainRe.MatchString(domain) {
		jsonErr(w, "Invalid domain name", http.StatusBadRequest)
		return
	}

	// Write cert to temp file
	tmpFile, err := os.CreateTemp("", "playground-cert-*.pem")
	if err != nil {
		jsonErr(w, "Failed to create temp file", http.StatusInternalServerError)
		return
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	// Generate cert
	var genArgs []string
	if certMode == "mtc" {
		genArgs = []string{"-mtc-mode", "-domain", domain, "-output", tmpFile.Name()}
	} else {
		genArgs = []string{"-domain", domain, "-output", tmpFile.Name()}
	}

	genCmd := exec.Command(binPath("demo-embedded-cert"), genArgs...)
	var genStderr bytes.Buffer
	genCmd.Stderr = &genStderr
	genCmd.Stdout = &bytes.Buffer{}

	start := time.Now()
	if err := genCmd.Run(); err != nil {
		resp := VerifyResult{
			Success:  false,
			Output:   genStderr.String(),
			Duration: fmt.Sprintf("%d", time.Since(start).Milliseconds()),
		}
		json.NewEncoder(w).Encode(resp)
		return
	}

	// Run verify
	verifyCmd := exec.Command(binPath("mtc-verify-cert"), "-cert", tmpFile.Name())
	var verifyOut bytes.Buffer
	verifyCmd.Stdout = &verifyOut
	verifyCmd.Stderr = &verifyOut

	if err := verifyCmd.Run(); err != nil {
		elapsed := time.Since(start)
		resp := VerifyResult{
			Success:  false,
			Output:   genStderr.String() + "\n\n--- Verify ---\n" + verifyOut.String(),
			Duration: fmt.Sprintf("%d", elapsed.Milliseconds()),
			CertPath: tmpFile.Name(),
		}
		json.NewEncoder(w).Encode(resp)
		return
	}

	elapsed := time.Since(start)
	fullOutput := genStderr.String() + "\n\n=== Certificate Verification ===\n" + verifyOut.String()
	steps := parseSteps(fullOutput)

	resp := VerifyResult{
		Success:  true,
		Output:   fullOutput,
		Duration: fmt.Sprintf("%d", elapsed.Milliseconds()),
		Steps:    steps,
		CertPath: tmpFile.Name(),
	}
	json.NewEncoder(w).Encode(resp)
}

// GET /api/conformance-info  — static info about the spec
func conformanceInfoHandler(w http.ResponseWriter, r *http.Request) {
	corsHeaders(w)
	info := map[string]interface{}{
		"spec":              "draft-ietf-plants-merkle-tree-certs-01",
		"oid_id_alg_mtcProof": "1.3.6.1.4.1.44363.47.0",
		"modes": []map[string]string{
			{
				"name":        "MTC-Spec (Primary)",
				"description": "signatureAlgorithm = id-alg-mtcProof; signatureValue carries binary MTCProof",
				"verification": "signatureless (landmark root hash) or signed (Ed25519 / ML-DSA cosigners)",
			},
			{
				"name":        "Legacy Embedded (Compatibility)",
				"description": "Standard X.509 cert with MTC inclusion proof in custom extension OID 1.3.6.1.4.1.99999.1.1",
				"verification": "backward-compatible with X.509 parsers",
			},
		},
		"algorithms": []string{"ECDSA P-256 (key generation)", "SHA-256 (Merkle hashing)", "Ed25519 (cosigner)", "ML-DSA-44/65/87 (post-quantum cosigner)"},
		"tree_construction": map[string]string{
			"leaf_hash": "SHA-256(0x00 ∥ data)",
			"node_hash": "SHA-256(0x01 ∥ left ∥ right)",
		},
		"tlog_protocol": "C2SP tlog-tiles",
		"acme_support":  "RFC 8555",
	}
	json.NewEncoder(w).Encode(info)
}

// ────────────────────────────────────────────────────────────────────────────
// Main
// ────────────────────────────────────────────────────────────────────────────

func main() {
	mux := http.NewServeMux()

	// Serve static UI
	mux.Handle("/", http.FileServer(http.Dir("static")))

	// API endpoints
	mux.HandleFunc("/api/health", healthHandler)
	mux.HandleFunc("/api/demo", demoHandler)
	mux.HandleFunc("/api/verify", verifyHandler)
	mux.HandleFunc("/api/conformance-info", conformanceInfoHandler)

	addr := ":8444"
	fmt.Println("🎮 MTC Playground Dashboard running at http://localhost" + addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
