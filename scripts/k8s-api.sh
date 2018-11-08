#!/usr/bin/env bash

ENDPOINT=192.168.99.50:8443
TOKEN=$1
API=v1beta1/customresourcedefinitions/components.component.k8s.io


http --verify=no \
    https://$ENDPOINT/$API \
    "Authorization: Bearer $TOKEN" \
    'Accept:application/json' \

# curl -k \
#     -H "Authorization: Bearer $TOKEN" \
#     -H 'Accept: application/json' \
#     https://$ENDPOINT/api/v1/endpoints