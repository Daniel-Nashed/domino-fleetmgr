#!/bin/bash

if [ -z "$NAMESPACE" ]; then
  NAMESPACE=domino
fi

if [ -z "$SECRET_NAME" ]; then
  SECRET_NAME="tls-secret"
fi

kubectl delete "secret/$SECRET_NAME" -n "$NAMESPACE"

