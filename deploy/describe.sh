#!/bin/bash

set -a
source .env
set +a

kubectl describe pod/${SERVERSETUP_SERVER_NAME}-0 -n "$NAMESPACE" 
