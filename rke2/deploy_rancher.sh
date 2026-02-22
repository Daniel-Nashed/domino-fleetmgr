#!/bin/bash


header()
{
  echo
  echo "--------------------------------------------------"
  echo "$@"
  echo "--------------------------------------------------"
  echo
}


log_error()
{
    echo
    echo "$@"
    echo
}


check_exists()
{

   if [ -e "$1" ];  then
     return 0
   fi

  if  [ -z "$2" ]; then
    log_error "Please make sure [$1] does exist before starting"
    exit 1
  fi

  log_error "$2: '$1'"
  exit 1
}



set -a # export all variables
. ./.env
set +a


header "Installing Rancher via Helm $SERVERSETUP_EXTERNAL_INETDOMAIN"


kubectl create namespace cattle-system
helm repo add rancher-stable https://releases.rancher.com/server-charts/stable
helm repo update

helm upgrade --install rancher rancher-stable/rancher \
  -n cattle-system --create-namespace \
  --set hostname=rancher.$SERVERSETUP_EXTERNAL_INETDOMAIN \
  --set global.imageRegistry=$IMAGE_REGISTRY \
  --set bootstrapPassword=$RANCHER_PASSWORD \
  --set ingress.enabled=false \
  --set tls=external


