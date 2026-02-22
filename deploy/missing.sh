#!/bin/bash

. .env envsubst --variables "$(cat domino.yml)"| sort | uniq
