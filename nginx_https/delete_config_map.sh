#!/bin/bash

if [ -z "$NAMESPACE" ]; then
  NAMESPACE=domino
fi


kubectl delete -n "$NAMESPACE" configmap domino-https-nginx-conf
