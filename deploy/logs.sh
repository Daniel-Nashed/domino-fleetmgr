#!/bin/bash

set -a
source .env
set +a

kubectl logs pod/${SERVERSETUP_SERVER_NAME}-0 -n "$NAMESPACE" 
