#!/bin/bash

set -a
source .env
set +a

kubectl delete -f yml/${SERVERSETUP_SERVER_NAME}.yml
