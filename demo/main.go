package main

import (
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
)

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
	var cmd *exec.Cmd

	caParamsPath := "ca/www/mtc/v04b/ca-params"
	if ca == "landmark" {
		caParamsPath = "landmark-ca/www/mtc/v04b/ca-params"
	}

	switch target {
	case "cert":
		cmd = exec.Command("../mtc-cli", "inspect", "-ca-params", caParamsPath, "cert", "website.mtc")
	case "landmark-cert":
		certFile := "landmark-website.mtc"
		if _, err := os.Stat(certFile); os.IsNotExist(err) {
			certFile = "landmark.mtc"
		}
		cmd = exec.Command("../mtc-cli", "inspect", "-ca-params", caParamsPath, "cert", certFile)
	case "vw":
		cmd = exec.Command("../mtc-cli", "inspect", "-ca-params", caParamsPath, "validity-window", "website.vw")
	case "landmark-vw":
		cmd = exec.Command("../mtc-cli", "inspect", "-ca-params", caParamsPath, "validity-window", "landmark.vw")
	case "ca-params":
		cmd = exec.Command("../mtc-cli", "inspect", "ca-params", caParamsPath)
	case "landmark-params":
		cmd = exec.Command("../mtc-cli", "inspect", "ca-params", "landmark-ca/www/mtc/v04b/ca-params")
	default:
		http.Error(w, "Invalid target", http.StatusBadRequest)
		return
	}

	output, err := cmd.CombinedOutput()
	resp := VerifyResponse{
		Success: err == nil,
		Output:  string(output),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func fileSizesHandler(w http.ResponseWriter, r *http.Request) {
	files := map[string]string{
		"cert":         "website.mtc",
		"vw":           "website.vw",
		"landmark-vw":  "landmark.vw",
		"landmark-cert": "landmark.mtc",
		"ca-params":    "ca/www/mtc/v04b/ca-params",
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
	caParamsOut, err := exec.Command("../mtc-cli", "inspect", "ca-params", "ca/www/mtc/v04b/ca-params").CombinedOutput()
	step1 := ProofStep{Step: 1, Actor: "Root CA", Action: "Load CA Parameters", Result: string(caParamsOut), Verified: err == nil}
	resp.Steps = append(resp.Steps, step1)

	// Step 2: Parse Landmark CA params
	lmParamsOut, err2 := exec.Command("../mtc-cli", "inspect", "ca-params", "landmark-ca/www/mtc/v04b/ca-params").CombinedOutput()
	step2 := ProofStep{Step: 2, Actor: "Landmark CA", Action: "Load Landmark Parameters", Result: string(lmParamsOut), Verified: err2 == nil}
	resp.Steps = append(resp.Steps, step2)

	// Step 3: Inspect validity window
	vwOut, err3 := exec.Command("../mtc-cli", "inspect", "-ca-params", "ca/www/mtc/v04b/ca-params", "validity-window", "website.vw").CombinedOutput()
	step3 := ProofStep{Step: 3, Actor: "Root CA", Action: "Inspect Signed Validity Window", Result: string(vwOut), Verified: err3 == nil}
	resp.Steps = append(resp.Steps, step3)

	// Step 4: Inspect certificate (Merkle proof)
	certOut, err4 := exec.Command("../mtc-cli", "inspect", "-ca-params", "ca/www/mtc/v04b/ca-params", "cert", "website.mtc").CombinedOutput()
	step4 := ProofStep{Step: 4, Actor: "Leaf Certificate", Action: "Parse Merkle Inclusion Proof", Result: string(certOut), Verified: err4 == nil}
	resp.Steps = append(resp.Steps, step4)

	// Step 5: Landmark VW
	lmVwOut, err5 := exec.Command("../mtc-cli", "inspect", "-ca-params", "landmark-ca/www/mtc/v04b/ca-params", "validity-window", "landmark.vw").CombinedOutput()
	step5 := ProofStep{Step: 5, Actor: "Landmark CA", Action: "Landmark Validity Window", Result: string(lmVwOut), Verified: err5 == nil}
	resp.Steps = append(resp.Steps, step5)

	// Step 6: Full verification
	verifyOut, verifyErr := exec.Command("../mtc-cli", "verify", "-ca-params", "ca/www/mtc/v04b/ca-params", "-validity-window", "website.vw", "website.mtc").CombinedOutput()
	step6 := ProofStep{Step: 6, Actor: "Verifier", Action: "Verify Merkle Inclusion Proof", Result: string(verifyOut), Verified: verifyErr == nil}
	resp.Steps = append(resp.Steps, step6)

	resp.Success = verifyErr == nil
	resp.ChainValid = verifyErr == nil

	// Build node summaries from inspect output
	resp.RootCA = buildProofNode("Root CA", string(caParamsOut), "website.vw")
	resp.Landmark = buildProofNode("Landmark CA", string(lmParamsOut), "landmark.vw")
	resp.LeafCert = buildLeafNode(string(certOut), "website.mtc")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func buildProofNode(name, inspectOut, vwPath string) ProofNode {
	node := ProofNode{Name: name}

	// Get tree head from validity window inspect
	for _, line := range strings.Split(inspectOut, "\n") {
		if strings.Contains(line, "batch_number") {
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
		if strings.Contains(line, "batch_number") || strings.Contains(line, "batch") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				node.BatchNumber = parts[len(parts)-1]
			}
		}
		if strings.Contains(line, "path") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				val := parts[len(parts)-1]
				n, err := strconv.Atoi(val)
				if err == nil {
					node.PathLength = n
				} else if len(val) > 0 {
					// path is hex-encoded; length = hex_len/2 / 32 = nodes
					node.PathLength = len(val) / 64
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
	}

	fmt.Println("🚀 MTC Demo Website running at https://localhost:8443")
	fmt.Println("This server forces TLS 1.3.")
	log.Fatal(server.ListenAndServeTLS("website.pem", "website.key"))
}
