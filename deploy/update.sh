#!/bin/bash

set -a
source .env
set +a

kubectl rollout restart statefulset/${SERVERSETUP_SERVER_NAME} -n "$NAMESPACE" 

