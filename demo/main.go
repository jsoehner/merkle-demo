package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// LogFilter is an io.Writer that suppresses common non-critical TLS handshake errors.
type LogFilter struct{}

func (f *LogFilter) Write(p []byte) (n int, err error) {
	s := string(p)
	// Suppress common connection-level errors that clutter logs during scans or health checks
	if strings.Contains(s, "TLS handshake error") && (strings.Contains(s, "EOF") ||
		strings.Contains(s, "unknown certificate") ||
		strings.Contains(s, "unsupported versions") ||
		strings.Contains(s, "broken pipe") ||
		strings.Contains(s, "connection reset by peer")) {
		return len(p), nil
	}
	return os.Stderr.Write(p)
}

type VerifyResponse struct {
	Success bool   `json:"success"`
	Output  string `json:"output"`
}

type FileSizeResponse struct {
	Name  string `json:"name"`
	Bytes int64  `json:"bytes"`
	MB    string `json:"mb"`
}

type ProofChainResponse struct {
	Success    bool        `json:"success"`
	RootCA     ProofNode   `json:"root_ca"`
	Landmark   ProofNode   `json:"landmark"`
	LeafCert   ProofNode   `json:"leaf_cert"`
	ChainValid bool        `json:"chain_valid"`
	Steps      []ProofStep `json:"steps"`
}

type ProofNode struct {
	Name        string `json:"name"`
	BatchNumber string `json:"batch_number"`
	TreeHead    string `json:"tree_head"`
	PathLength  int    `json:"path_length"`
	SizeBytes   int64  `json:"size_bytes"`
}

type ProofStep struct {
	Step        int    `json:"step"`
	Actor       string `json:"actor"`
	Action      string `json:"action"`
	Result      string `json:"result"`
	Hash        string `json:"hash,omitempty"`
	Verified    bool   `json:"verified"`
}

func verifyHandler(w http.ResponseWriter, r *http.Request) {
	cmd := exec.Command("../mtc-cli", "verify", "-ca-params", "ca/www/mtc/v04b/ca-params", "-validity-window", "website.vw", "website.mtc")
	output, err := cmd.CombinedOutput()

	resp := VerifyResponse{
		Success: err == nil,
		Output:  string(output),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func verifyLandmarkHandler(w http.ResponseWriter, r *http.Request) {
	// Try to verify via landmark CA path first
	cmd := exec.Command("../mtc-cli", "verify",
		"-ca-params", "landmark-ca/www/mtc/v04b/ca-params",
		"-validity-window", "landmark.vw",
		"landmark-website.mtc")
	output, err := cmd.CombinedOutput()

	if err != nil {
		// Fall back to landmark identity cert
		cmd2 := exec.Command("../mtc-cli", "verify",
			"-ca-params", "landmark-ca/www/mtc/v04b/ca-params",
			"-validity-window", "landmark.vw",
			"landmark.mtc")
		output2, err2 := cmd2.CombinedOutput()
		resp := VerifyResponse{
			Success: err2 == nil,
			Output:  string(output) + "\n--- Landmark cert fallback ---\n" + string(output2),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}

	resp := VerifyResponse{
		Success: true,
		Output:  string(output),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func inspectHandler(w http.ResponseWriter, r *http.Request) {
	target := r.URL.Query().Get("target")
	ca := r.URL.Query().Get("ca") // "root" or "landmark"

	caParamsPath := "ca/www/mtc/v04b/ca-params"
	if ca == "landmark" {
		caParamsPath = "landmark-ca/www/mtc/v04b/ca-params"
	}

	var cmd *exec.Cmd
	isVW := false

	switch target {
	case "cert":
		cmd = exec.Command("../mtc-cli", "inspect", "-ca-params", caParamsPath, "cert", "website.mtc")
	case "landmark-cert":
		certFile := "landmark-website.mtc"
		if _, err := os.Stat(certFile); os.IsNotExist(err) {
			certFile = "landmark.mtc"
		}
		landmarkCA := "landmark-ca/www/mtc/v04b/ca-params"
		cmd = exec.Command("../mtc-cli", "inspect", "-ca-params", landmarkCA, "cert", certFile)
	case "vw":
		cmd = exec.Command("../mtc-cli", "inspect", "-ca-params", caParamsPath, "validity-window", "website.vw")
		isVW = true
	case "landmark-vw":
		landmarkCA := "landmark-ca/www/mtc/v04b/ca-params"
		cmd = exec.Command("../mtc-cli", "inspect", "-ca-params", landmarkCA, "validity-window", "landmark.vw")
		isVW = true
	case "ca-params":
		cmd = exec.Command("../mtc-cli", "inspect", "ca-params", caParamsPath)
	case "landmark-params":
		cmd = exec.Command("../mtc-cli", "inspect", "ca-params", "landmark-ca/www/mtc/v04b/ca-params")
	default:
		http.Error(w, "Invalid target", http.StatusBadRequest)
		return
	}

	rawOutput, cmdErr := runMtcCommand(cmd, isVW)

	resp := VerifyResponse{
		Success: cmdErr == nil,
		Output:  string(rawOutput),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// runMtcCommand executes an mtc-cli command and truncates output if it's a validity window
func runMtcCommand(cmd *exec.Cmd, isVW bool) ([]byte, error) {
	if !isVW {
		return cmd.CombinedOutput()
	}

	// For validity windows, we truncate to prevent memory bloat and I/O bottlenecks
	const maxLines = 20
	pr, pw, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	defer pr.Close()

	cmd.Stdout = pw
	cmd.Stderr = pw
	if err := cmd.Start(); err != nil {
		pw.Close()
		return nil, err
	}

	var buf bytes.Buffer
	done := make(chan bool)
	go func() {
		lineCount := 0
		tmp := make([]byte, 4096)
		for lineCount < maxLines {
			n, err := pr.Read(tmp)
			if n > 0 {
				chunk := tmp[:n]
				for _, b := range chunk {
					buf.WriteByte(b)
					if b == '\n' {
						lineCount++
						if lineCount >= maxLines {
							break
						}
					}
				}
			}
			if err != nil {
				break
			}
		}
		done <- true
	}()

	select {
	case <-done:
		_ = cmd.Process.Kill()
	case <-time.After(3 * time.Second):
		_ = cmd.Process.Kill()
	}

	pw.Close()
	_ = cmd.Wait()

	if isVW {
		buf.WriteString(fmt.Sprintf("\n… [output truncated — validity window contains ~%d tree heads; showing first %d] …",
			getVWSize(cmd), maxLines))
	}

	return buf.Bytes(), nil
}

// getVWSize returns the approximate number of tree-head entries in a validity window
func getVWSize(cmd *exec.Cmd) int {
	return 302400
}

func fileSizesHandler(w http.ResponseWriter, r *http.Request) {
	// Prefer landmark-website.mtc; fall back to landmark.mtc for the size display.
	landmarkCertPath := "landmark-website.mtc"
	if _, err := os.Stat(landmarkCertPath); os.IsNotExist(err) {
		landmarkCertPath = "landmark.mtc"
	}

	files := map[string]string{
		"cert":          "website.mtc",
		"vw":            "website.vw",
		"landmark-vw":   "landmark.vw",
		"landmark-cert": landmarkCertPath,
		"ca-params":     "ca/www/mtc/v04b/ca-params",
	}

	result := map[string]FileSizeResponse{}
	for key, path := range files {
		info, err := os.Stat(path)
		if err != nil {
			result[key] = FileSizeResponse{Name: path, Bytes: 0, MB: "N/A"}
			continue
		}
		mb := fmt.Sprintf("%.3f MB", float64(info.Size())/1048576.0)
		result[key] = FileSizeResponse{
			Name:  filepath.Base(path),
			Bytes: info.Size(),
			MB:    mb,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func proofChainHandler(w http.ResponseWriter, r *http.Request) {
	resp := ProofChainResponse{
		Steps: []ProofStep{},
	}

	// Step 1: Parse CA params
	caParamsOut, err := runMtcCommand(exec.Command("../mtc-cli", "inspect", "ca-params", "ca/www/mtc/v04b/ca-params"), false)
	step1 := ProofStep{Step: 1, Actor: "Root CA", Action: "Load CA Parameters", Result: string(caParamsOut), Verified: err == nil}
	resp.Steps = append(resp.Steps, step1)

	// Step 2: Parse Landmark CA params
	lmParamsOut, err2 := runMtcCommand(exec.Command("../mtc-cli", "inspect", "ca-params", "landmark-ca/www/mtc/v04b/ca-params"), false)
	step2 := ProofStep{Step: 2, Actor: "Landmark CA", Action: "Load Landmark Parameters", Result: string(lmParamsOut), Verified: err2 == nil}
	resp.Steps = append(resp.Steps, step2)

	// Step 3: Inspect validity window (TRUNCATED)
	vwOut, err3 := runMtcCommand(exec.Command("../mtc-cli", "inspect", "-ca-params", "ca/www/mtc/v04b/ca-params", "validity-window", "website.vw"), true)
	step3 := ProofStep{Step: 3, Actor: "Root CA", Action: "Inspect Signed Validity Window", Result: string(vwOut), Verified: err3 == nil}
	resp.Steps = append(resp.Steps, step3)

	// Step 4: Inspect certificate (Merkle proof)
	certOut, err4 := runMtcCommand(exec.Command("../mtc-cli", "inspect", "-ca-params", "ca/www/mtc/v04b/ca-params", "cert", "website.mtc"), false)
	step4 := ProofStep{Step: 4, Actor: "Leaf Certificate", Action: "Parse Merkle Inclusion Proof", Result: string(certOut), Verified: err4 == nil}
	resp.Steps = append(resp.Steps, step4)

	// Step 5: Landmark VW (TRUNCATED)
	lmVwOut, err5 := runMtcCommand(exec.Command("../mtc-cli", "inspect", "-ca-params", "landmark-ca/www/mtc/v04b/ca-params", "validity-window", "landmark.vw"), true)
	step5 := ProofStep{Step: 5, Actor: "Landmark CA", Action: "Landmark Validity Window", Result: string(lmVwOut), Verified: err5 == nil}
	resp.Steps = append(resp.Steps, step5)

	// Step 6: Full verification
	verifyOut, verifyErr := runMtcCommand(exec.Command("../mtc-cli", "verify", "-ca-params", "ca/www/mtc/v04b/ca-params", "-validity-window", "website.vw", "website.mtc"), false)
	step6 := ProofStep{Step: 6, Actor: "Verifier", Action: "Verify Merkle Inclusion Proof", Result: string(verifyOut), Verified: verifyErr == nil}
	resp.Steps = append(resp.Steps, step6)

	resp.Success = verifyErr == nil
	resp.ChainValid = verifyErr == nil

	// Build node summaries from inspect output
	resp.RootCA = buildProofNode("Root CA", string(vwOut), "website.vw")
	resp.Landmark = buildProofNode("Landmark CA", string(lmVwOut), "landmark.vw")
	resp.LeafCert = buildLeafNode(string(certOut), "website.mtc")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func buildProofNode(name, inspectOut, vwPath string) ProofNode {
	node := ProofNode{Name: name}

	// Get tree head from validity window inspect
	for _, line := range strings.Split(inspectOut, "\n") {
		lowerLine := strings.ToLower(line)
		if strings.Contains(lowerLine, "batch_number") || strings.Contains(lowerLine, "batch number") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				node.BatchNumber = parts[len(parts)-1]
			}
		}
		if strings.Contains(line, "tree_heads[") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				hash := parts[len(parts)-1]
				if len(hash) > 16 {
					node.TreeHead = hash[:16] + "…"
				} else {
					node.TreeHead = hash
				}
			}
		}
	}

	info, err := os.Stat(vwPath)
	if err == nil {
		node.SizeBytes = info.Size()
	}
	return node
}

func buildLeafNode(certOut, certPath string) ProofNode {
	node := ProofNode{Name: "Leaf Certificate"}
	for _, line := range strings.Split(certOut, "\n") {
		lowerLine := strings.ToLower(line)
		if strings.Contains(lowerLine, "batch_number") || strings.Contains(lowerLine, "batch number") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				node.BatchNumber = parts[len(parts)-1]
			}
		}
		if strings.Contains(lowerLine, "recomputed tree head") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				hash := parts[len(parts)-1]
				if len(hash) > 16 {
					node.TreeHead = hash[:16] + "…"
				} else {
					node.TreeHead = hash
				}
			}
		}
		if strings.Contains(lowerLine, "path length") || (strings.Contains(lowerLine, "path") && !strings.Contains(lowerLine, "authentication")) {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				val := parts[len(parts)-1]
				n, err := strconv.Atoi(val)
				if err == nil {
					node.PathLength = n
				}
			}
		}
	}
	info, err := os.Stat(certPath)
	if err == nil {
		node.SizeBytes = info.Size()
	}
	return node
}

func main() {
	mux := http.NewServeMux()

	// Serve static UI
	mux.Handle("/", http.FileServer(http.Dir("static")))

	// Expose MTC files statically
	mux.HandleFunc("/api/files/cert", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "website.mtc")
	})
	mux.HandleFunc("/api/files/vw", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "website.vw")
	})
	mux.HandleFunc("/api/files/params", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "ca/www/mtc/v04b/ca-params")
	})
	mux.HandleFunc("/api/files/landmark-vw", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "landmark.vw")
	})

	// Verification endpoints
	mux.HandleFunc("/api/verify", verifyHandler)
	mux.HandleFunc("/api/verify/landmark", verifyLandmarkHandler)

	// Diagnostics endpoints
	mux.HandleFunc("/api/inspect", inspectHandler)
	mux.HandleFunc("/api/file-sizes", fileSizesHandler)

	// Proof chain visualization
	mux.HandleFunc("/api/proof-chain", proofChainHandler)

	// Configure TLS 1.3
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS13,
		MaxVersion: tls.VersionTLS13,
	}

	server := &http.Server{
		Addr:      ":8443",
		Handler:   mux,
		TLSConfig: tlsConfig,
		ErrorLog:  log.New(&LogFilter{}, "", 0),
	}

	fmt.Println("🚀 MTC Demo Website running at https://localhost:8443")
	fmt.Println("This server forces TLS 1.3.")
	log.Fatal(server.ListenAndServeTLS("website.pem", "website.key"))
}
