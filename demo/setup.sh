#!/bin/bash
set -e

# Path to the mtc CLI tool we built
MTC_CLI="../mtc-cli"

echo "=== Merkle Tree Certificates (MTC) Demo Setup ==="

# 1. Clean up old state
rm -rf ca website.key website.pub website-asr website.mtc website.vw website.pem

echo "[1/4] Setting up Merkle Tree CA..."
$MTC_CLI ca -p ca new --batch-duration 2s --lifetime 8760h 62253.12.15 localhost:8080

echo "[2/4] Generating P-256 key for our website..."
openssl ecparam -name prime256v1 -genkey -out website.key
chmod 644 website.key  # Allow the server process to read the key regardless of how it's launched
openssl ec -in website.key -pubout -out website.pub
# Generate a self-signed X.509 cert to use standard TLS alongside MTC
openssl req -new -x509 -key website.key -out website.pem -days 365 -subj "/CN=localhost"

echo "[3/4] Queueing assertion request for website..."
$MTC_CLI new-assertion-request --tls-pem website.pub --dns localhost --ip4 127.0.0.1 -o website-asr
$MTC_CLI ca -p ca queue -i website-asr

echo "[4/4] Issuing batches..."
# We wait a bit to ensure the batch duration has passed
sleep 5
$MTC_CLI ca -p ca issue
echo "First batch issued."

# Wait again and issue a second batch so we have a valid validity-window with a previous tree head
sleep 5
$MTC_CLI ca -p ca queue --tls-pem website.pub --dns other.localhost
$MTC_CLI ca -p ca issue
echo "Second batch issued."

echo "Extracting the MTC certificate for the website..."
$MTC_CLI ca -p ca cert -i website-asr -o website.mtc

echo "Extracting the signed validity window..."
cp ca/www/mtc/v04b/batches/latest/validity-window website.vw

echo "=== Setup Complete! ==="
echo ""
echo "=== DIAGNOSTICS: CA Parameters ==="
$MTC_CLI inspect ca-params ca/www/mtc/v04b/ca-params
echo ""
echo "=== DIAGNOSTICS: Validity Window ==="
$MTC_CLI inspect -ca-params ca/www/mtc/v04b/ca-params validity-window website.vw
echo ""
echo "=== DIAGNOSTICS: Website MTC Certificate ==="
$MTC_CLI inspect -ca-params ca/www/mtc/v04b/ca-params cert website.mtc
echo ""
echo "=== DIAGNOSTICS: Live Verification ==="
$MTC_CLI verify -ca-params ca/www/mtc/v04b/ca-params -validity-window website.vw website.mtc && echo "Verification PASSED ✅" || echo "Verification FAILED ❌"
echo ""
echo "All payloads generated and ready."
