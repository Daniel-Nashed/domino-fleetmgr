#!/bin/bash

curl -sfL https://get.rke2.io | sh -

systemctl enable --now rke2-server.service

ln -s /var/lib/rancher/rke2/bin/kubectl /usr/local/bin/kubectl

mkdir -p ~/.kube
ln -s /etc/rancher/rke2/rke2.yaml ~/.kube/config

