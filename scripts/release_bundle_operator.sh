#!/usr/bin/env bash

# Command: ./release_bundle_operator.sh QUAY_USER QUAY_PASSWORD

export QUAY_USER=$1
export QUAY_PWD=$2
export AUTH_TOKEN=$(curl -X POST -H "Content-Type: application/json" -d '{"user":{"username":"'"$QUAY_USER"'","password":"'"$QUAY_PWD"'"}}' https://quay.io/cnr/api/v1/users/login | jq -r '.token')
export QUAY_ORG="halkyonio"
export REPOSITORY="operator"
export RELEASE="0.0.1"
export BUNDLE_DIR="deploy/olm-catalog/bundle"

operator-courier verify $BUNDLE_DIR
operator-courier push $BUNDLE_DIR $QUAY_ORG $REPOSITORY $RELEASE "$AUTH_TOKEN"
