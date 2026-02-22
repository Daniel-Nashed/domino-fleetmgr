#!/bin/bash

export DOMSETUP_HOST=127.0.0.1
export DOMSETUP_PORT=1352
DOMSETUP_BEARER_FILE=/tmp/domsetup_bearer.txt


set -a
source .env
set +a


if [ -z "$POD_NAME" ]; then
  export POD_NAME=${SERVERSETUP_SERVER_NAME}-0
fi

if [ -z "$DOMSETUP_MAX_POD_WAIT_SEC" ]; then
  DOMSETUP_MAX_OTS_WAIT_SEC=300
fi

if [ -z "$DOMSETUP_MAX_POD_WAIT_SEC" ]; then
  DOMSETUP_MAX_POD_WAIT_SEC=300
fi


delim()
{
  echo "--------------------------------------------------------------------------------"
}


header()
{
  echo
  delim
  echo "$@"
  delim
  echo
}


log()
{
  echo
  echo "$@"
  echo
}


remove_file()
{
  if [ -z "$1" ]; then
    return 1
  fi

  if [ ! -e "$1" ]; then
    return 2
  fi

  rm -f "$1"
  return 0
}


wait_for_ots_ready()
{
  local MAX_SECONDS=60

  if [ -n "$1" ]; then
    MAX_SECONDS=$1
  fi

  echo

  SECONDS=0
  while [ "$SECONDS" -lt "$MAX_SECONDS" ]; do

    DOMSETUP_STATUS=$(curl -sk --connect-timeout 5 --max-time 10 https://$DOMSETUP_HOST:$DOMSETUP_PORT/status)

    if [ -n "$DOMSETUP_STATUS" ]; then
      echo "Domino Setup is ready after $SECONDS seconds"
      return 0
    fi

    sleep 2

  done
}


wait_for_pod_running()
{
  local MAX_SECONDS=60

  if [ -n "$1" ]; then
    MAX_SECONDS=$1
  fi

  echo

  SECONDS=0
  while [ "$SECONDS" -lt "$MAX_SECONDS" ]; do

    POD_STATUS=$(kubectl get pod "$POD_NAME" -n "$NAMESPACE" 2> /dev/null -o jsonpath='{.status.phase}')

    if [ "$POD_STATUS" = "Running" ]; then
      echo "Domino Pod running after $SECONDS seconds"
      return 0
    fi

    sleep 2

  done
}

# --- Main ---

# Create a stable secret file for setup retries
if [ -z "$DOMSETUP_BEARER" ]; then
  if [ -e "$DOMSETUP_BEARER_FILE" ]; then
    DOMSETUP_BEARER=$(cat $DOMSETUP_BEARER_FILE)
  else
    export DOMSETUP_BEARER="$(openssl rand -hex 42)"
    echo -n "$DOMSETUP_BEARER" > "$DOMSETUP_BEARER_FILE"
  fi
fi


header "Creating Domino StatefulSet"

envsubst < domino.yml > "yml/$SERVERSETUP_SERVER_NAME.yml"
kubectl apply -f "yml/$SERVERSETUP_SERVER_NAME.yml" 

echo "Domino StatefulSet created"


header "Waiting until Domino pod $POD_NAME is running (Max sec: $DOMSETUP_MAX_POD_WAIT_SEC)"

wait_for_pod_running "$DOMSETUP_MAX_POD_WAIT_SEC"

if [ "$POD_STATUS" != "Running" ]; then
  log "Pod $POD_NAME is not running"
  exit 1
fi

sleep 5


header "Starting kubectl port forwarding"

LOG_FILE="/tmp/${SERVERSETUP_SERVER_NAME}_setup.log"

kubectl -n "$NAMESPACE" port-forward pod/$POD_NAME 1352:1352 > "$LOG_FILE" 2>&1 &
KUBECTL_PID=$!

sleep 5

log "Port forwarding via kubectl started with PID: $KUBECTL_PID"
cat "$LOG_FILE"
echo

header "Waiting until OTS listener is running (Max sec: $DOMSETUP_MAX_OTS_WAIT_SEC)"

wait_for_ots_ready "$DOMSETUP_MAX_OTS_WAIT_SEC"

echo

if [ -z "$DOMSETUP_STATUS" ]; then
  log "DomSetup is not ready"
  exit 1
fi

header "Running OTS config"

# Don't prompt for environment variables already present it environment
export domCfgJSON_mode=force

domsetup file:/opt/nashcom/startscript/OneTouchSetup/first_server.json

sleep 1
kill "$KUBECTL_PID"
rm "$LOG_FILE"


SETUP_COMPLETED=$(kubectl -n "$NAMESPACE" logs pod/$POD_NAME | grep "Starting Domino for xLinux")

if [ -n "$SETUP_COMPLETED" ]; then
  header "Domino Setup completed"
  remove_file "$DOMSETUP_BEARER_FILE"
else
  header "Domino Setup failed"
  kubectl -n "$NAMESPACE" logs pod/$POD_NAME
fi

echo
