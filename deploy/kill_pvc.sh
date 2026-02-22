#!/bin/bash

set -a
source .env
set +a

kubectl delete pvc/domino-data-${SERVERSETUP_SERVER_NAME}-0 -n "$NAMESPACE"
