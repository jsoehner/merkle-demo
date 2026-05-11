package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
)

type VerifyResponse struct {
	Success bool   `json:"success"`
	Output  string `json:"output"`
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

func inspectHandler(w http.ResponseWriter, r *http.Request) {
	target := r.URL.Query().Get("target")
	var cmd *exec.Cmd
	
	switch target {
	case "cert":
		cmd = exec.Command("../mtc-cli", "inspect", "-ca-params", "ca/www/mtc/v04b/ca-params", "cert", "website.mtc")
	case "vw":
		cmd = exec.Command("../mtc-cli", "inspect", "-ca-params", "ca/www/mtc/v04b/ca-params", "validity-window", "website.vw")
	case "ca-params":
		cmd = exec.Command("../mtc-cli", "inspect", "ca-params", "ca/www/mtc/v04b/ca-params")
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

func main() {
	mux := http.NewServeMux()
	
	// Serve static UI
	mux.Handle("/", http.FileServer(http.Dir("static")))
	
	// Expose MTC files statically so the frontend can display their contents or sizes
	mux.HandleFunc("/api/files/cert", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "website.mtc")
	})
	mux.HandleFunc("/api/files/vw", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "website.vw")
	})
	mux.HandleFunc("/api/files/params", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "ca/www/mtc/v04b/ca-params")
	})
	
	// Verification endpoint
	mux.HandleFunc("/api/verify", verifyHandler)
	
	// Diagnostics endpoint
	mux.HandleFunc("/api/inspect", inspectHandler)
	
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
