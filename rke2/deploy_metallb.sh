#!/bin/bash


helm repo add metallb https://metallb.github.io/metallb
helm repo update

helm upgrade --install metallb metallb/metallb -n metallb-system --create-namespace

