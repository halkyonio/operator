#!/usr/bin/env bash

export QUAY_USER="ch007m+dabou"
export QUAY_PWD="KTT2EFOYCTXPQ7AYNBBFL7LXGI691UWS9TZ0MIR8B8ORWPSJCRA84SGJN7WVHGGP"
export AUTH_TOKEN=$(curl -X POST -H "Content-Type: application/json" -d '{"user":{"username":"'"$QUAY_USER"'","password":"'"$QUAY_PWD"'"}}' https://quay.io/cnr/api/v1/users/login | jq -r '.token')
export QUAY_ORG="ch007m"
export REPOSITORY="component"
export RELEASE="0.10.0"
export BUNDLE_DIR="deploy/olm-catalog/bundle"

operator-courier verify $BUNDLE_DIR
operator-courier push $BUNDLE_DIR $QUAY_ORG $REPOSITORY $RELEASE "$AUTH_TOKEN"
