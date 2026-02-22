#!/bin/bash

if [ -z "$NAMESPACE" ]; then
  NAMESPACE=domino
fi

if [ -z "$SECRET_NAME" ]; then
  SECRET_NAME=tls-nginx-secret
fi

# Create namespace if not present
kubectl create namespace "$NAMESPACE" --dry-run=client -o yaml | kubectl apply -f -

# Create or update NGINX TLS Secret
kubectl create secret generic "$SECRET_NAME" -n "$NAMESPACE" --from-file=cert.pem --from-file=key.pem --from-file=password.txt --save-config --dry-run=client -o yaml | kubectl apply -f -

