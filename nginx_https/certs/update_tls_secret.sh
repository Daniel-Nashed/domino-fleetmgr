#!/bin/bash

if [ -z "$NAMESPACE" ]; then
  NAMESPACE=domino
fi

if [ -z "$SECRET_NAME" ]; then
  SECRET_NAME="tls-secret"
fi

if [ -z "$CERT_FILE" ]; then
  CERT_FILE="cert.pem"
fi

if [ -z "$KEY_FILE" ]; then
  KEY_FILE="key_unencrypted.pem"
fi

# Create namespace if not present
kubectl create namespace "$NAMESPACE" --dry-run=client -o yaml | kubectl apply -f -

# Create or update the secret in place
kubectl create secret tls "$SECRET_NAME" --cert="$CERT_FILE" --key="$KEY_FILE" --namespace "$NAMESPACE" --dry-run=client -o yaml | kubectl apply -f -

