#!/bin/bash

. .env
kubectl create secret docker-registry -n "$NAMESPACE" regcred --docker-server=$REGISTRY_HOST --docker-username=$REGISTRY_USER --docker-password=$REGISTRY_PASSWORD
