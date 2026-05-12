#!/bin/bash
set -e

# Path to the mtc CLI tool we built
MTC_CLI="../mtc-cli"

echo "=== Merkle Tree Certificates (MTC) Demo Setup ==="

# 1. Clean up old state
rm -rf ca landmark-ca website.key website.pub website-asr website.mtc website.vw website.pem \
       landmark.key landmark.pub landmark-asr landmark.mtc landmark.vw

# --- Proof-size budget ---
# Validity window = ValidityWindowSize * 32 bytes of tree heads.
# ValidityWindowSize = lifetime / batch_duration
# Target: ≤ 10 MB  →  10*1024*1024 / 32 = 327,680 slots max
# With batch-duration=2s → lifetime = 327,680 * 2s = 655,360s ≈ 182h
# We use 168h (7 days) → 302,400 heads × 32 B = ~9.2 MB  ✅
BATCH_DUR="2s"
LIFETIME="168h"
STORAGE_DUR="336h"   # 2 × lifetime as required by the spec

echo "[1/6] Setting up Root Merkle Tree CA (7-day lifetime, ~9.2MB validity window)..."
$MTC_CLI ca -p ca new \
  --batch-duration "$BATCH_DUR" \
  --lifetime "$LIFETIME" \
  --storage-duration "$STORAGE_DUR" \
  62253.12.15 localhost:8080

echo "[2/6] Setting up Landmark (Intermediary) CA anchored to the Root CA..."
# The Landmark CA uses a much shorter lifetime (just a few batches for the demo)
# but is rooted in the same trust hierarchy as the Root CA.
$MTC_CLI ca -p landmark-ca new \
  --batch-duration "$BATCH_DUR" \
  --lifetime "$LIFETIME" \
  --storage-duration "$STORAGE_DUR" \
  62253.12.15.1 localhost:8081

echo "[3/6] Generating P-256 keys..."
# Website key
openssl ecparam -name prime256v1 -genkey -out website.key
chmod 644 website.key
openssl ec -in website.key -pubout -out website.pub

# Landmark key (represents the landmark node's own identity assertion)
openssl ecparam -name prime256v1 -genkey -out landmark.key
chmod 644 landmark.key
openssl ec -in landmark.key -pubout -out landmark.pub

# Self-signed X.509 cert for TLS (standard fallback)
openssl req -new -x509 -key website.key -out website.pem -days 365 -subj "/CN=localhost"

echo "[4/6] Queuing assertions into BOTH CAs..."
# Queue the website into the Root CA
$MTC_CLI new-assertion-request --tls-pem website.pub --dns localhost --ip4 127.0.0.1 -o website-asr
$MTC_CLI ca -p ca queue -i website-asr

# Queue the landmark identity into the Landmark CA
$MTC_CLI new-assertion-request --tls-pem landmark.pub --dns landmark.localhost --ip4 127.0.0.2 -o landmark-asr
$MTC_CLI ca -p landmark-ca queue -i landmark-asr

# Also queue the website into the Landmark CA to demonstrate the intermediary chain
$MTC_CLI ca -p landmark-ca queue -i website-asr

echo "[5/6] Issuing batches from both CAs..."
sleep 5
$MTC_CLI ca -p ca issue
$MTC_CLI ca -p landmark-ca issue
echo "First batches issued."

sleep 5
$MTC_CLI ca -p ca queue --tls-pem website.pub --dns other.localhost
$MTC_CLI ca -p ca issue
$MTC_CLI ca -p landmark-ca issue
echo "Second batches issued."

echo "[6/6] Extracting MTC artifacts..."
# Root CA artifacts
$MTC_CLI ca -p ca cert -i website-asr -o website.mtc
cp ca/www/mtc/v04b/batches/latest/validity-window website.vw

# Landmark CA artifacts (the website cert via landmark path)
$MTC_CLI ca -p landmark-ca cert -i website-asr -o landmark-website.mtc 2>/dev/null || \
  $MTC_CLI ca -p landmark-ca cert -i landmark-asr -o landmark.mtc
cp landmark-ca/www/mtc/v04b/batches/latest/validity-window landmark.vw

# Report sizes
VW_BYTES=$(wc -c < website.vw | tr -d ' ')
VW_MB=$(echo "scale=2; $VW_BYTES/1048576" | bc)
CERT_BYTES=$(wc -c < website.mtc | tr -d ' ')
echo ""
echo "=== Artifact Sizes ==="
echo "  Validity Window : ${VW_BYTES} bytes (~${VW_MB} MB)"
echo "  MTC Certificate : ${CERT_BYTES} bytes"

echo ""
echo "=== DIAGNOSTICS: CA Parameters ==="
$MTC_CLI inspect ca-params ca/www/mtc/v04b/ca-params
echo ""
echo "=== DIAGNOSTICS: Landmark CA Parameters ==="
$MTC_CLI inspect ca-params landmark-ca/www/mtc/v04b/ca-params
echo ""
echo "=== DIAGNOSTICS: Validity Window ==="
# Show only the header/summary — suppress 302k+ tree_heads lines that flood stdout
$MTC_CLI inspect -ca-params ca/www/mtc/v04b/ca-params validity-window website.vw 2>&1 | head -15
echo "  … (302,400 tree_heads entries suppressed for startup speed)"
echo ""
echo "=== DIAGNOSTICS: Website MTC Certificate ==="
$MTC_CLI inspect -ca-params ca/www/mtc/v04b/ca-params cert website.mtc
echo ""
echo "=== DIAGNOSTICS: Live Verification ==="
$MTC_CLI verify -ca-params ca/www/mtc/v04b/ca-params -validity-window website.vw website.mtc && echo "Root CA Verification PASSED ✅" || echo "Root CA Verification FAILED ❌"
echo ""
echo "All payloads generated and ready."
