#!/bin/bash
# Generate self-signed certificate for HTTPS support
# Usage: ./scripts/generate-self-signed-cert.sh [CN] [SANS]

set -e

CERT_DIR="${CERT_DIR:-./certs}"
CN="${1:-localhost}"
SANS="${2:-DNS:localhost,DNS:*.local,IP:127.0.0.1,IP:::1}"
DAYS="${CERT_DAYS:-3650}"

mkdir -p "$CERT_DIR"

echo "Generating self-signed certificate..."
echo "CN: $CN"
echo "SANs: $SANS"
echo "Valid for: $DAYS days"
echo "Output directory: $CERT_DIR"

openssl req -x509 -newkey rsa:4096 \
  -keyout "$CERT_DIR/server.key" \
  -out "$CERT_DIR/server.crt" \
  -days "$DAYS" \
  -nodes \
  -subj "/CN=$CN" \
  -addext "subjectAltName=$SANS"

chmod 600 "$CERT_DIR/server.key"
chmod 644 "$CERT_DIR/server.crt"

echo ""
echo "✅ Certificate generated successfully!"
echo "   Certificate: $CERT_DIR/server.crt"
echo "   Private Key: $CERT_DIR/server.key"
echo ""
echo "⚠️  This is a self-signed certificate. Browsers will show a security warning."
echo "   For production, use certificates from a trusted CA."
