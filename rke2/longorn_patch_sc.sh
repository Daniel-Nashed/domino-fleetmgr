#!/bin/bash

kubectl patch sc longhorn -p '{"metadata": {"annotations":{"storageclass.kubernetes.io/is-default-class":"false"}}}'
kubectl patch sc longhorn-static -p '{"metadata": {"annotations":{"storageclass.kubernetes.io/is-default-class":"true"}}}'

