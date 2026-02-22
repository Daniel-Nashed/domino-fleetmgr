#!/bin/bash

./delete_config_map.sh
kubectl delete -f https_daemon_set.yml

