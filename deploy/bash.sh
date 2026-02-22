#!/bin/bash

set -a
source .env
set +a

kubectl exec -it pod/${SERVERSETUP_SERVER_NAME}-0 -n "$NAMESPACE" -- bash 
