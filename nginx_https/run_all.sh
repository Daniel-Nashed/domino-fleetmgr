#!/bin/bash

get_logs()
{
  echo
  echo "--- Container Logs ---"
  echo
  kubectl logs daemonset/domino-https -n "$NAMESPACE"
  echo
}


if [ -z "$NAMESPACE" ]; then
  NAMESPACE=domino
fi

echo

case "$1" in

  "")
    ;;

  log)
    get_logs
    exit 0
    ;;

  bash|sh)
    kubectl exec -it daemonset/domino-https -n "$NAMESPACE" -- $1
    exit 0
    ;;

  *)
    echo "Invalid option: $1"
    exit 1
    ;;

esac

./kill_all.sh
./create_config_map.sh
kubectl apply -f https_daemon_set.yml

sleep 5

get_logs

