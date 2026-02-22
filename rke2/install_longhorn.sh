#!/bin/bash

helm repo add longhorn https://charts.longhorn.io
helm repo update
helm search repo longhorn/longhorn --versions

helm install longhorn longhorn/longhorn --namespace longhorn-system --create-namespace

kubectl -n longhorn-system get pod


# Inspect defaults for a specific version
#helm show values longhorn/longhorn --version 1.9.1 > values.yaml

# Install (pinned)
#helm install longhorn longhorn/longhorn --namespace longhorn-system --create-namespace --version 1.9.1 -f values.yaml


# Upgrade later (still pinned)
#helm upgrade longhorn longhorn/longhorn -n longhorn-system --version 1.9.2 -f values.yaml

