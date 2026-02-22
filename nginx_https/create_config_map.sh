#!/bin/bash


if [ -z "$NAMESPACE" ]; then
  NAMESPACE=domino
fi

kubectl create -n "$NAMESPACE" configmap domino-https-nginx-conf --from-file=nginx.conf=./nginx.conf
