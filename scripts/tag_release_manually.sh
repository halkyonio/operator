#!/bin/bash

TAG_ID=$2
GITHUB_API_TOKEN=$1

OWNER="snowdrop"
REPO="component-operator"
AUTH="Authorization: token $GITHUB_API_TOKEN"
GH_API="https://api.github.com"
GH_REPO="$GH_API/repos/$OWNER/$REPO"

echo "Tagging ..."
git tag -a $TAG_ID -m "$TAG_ID"

echo "Releasing $TAG_ID ..."
JSON='{"tag_name": "'"$TAG_ID"'","target_commitish": "master","name": "'"$TAG_ID"'","body": "'"$TAG_ID"'","draft": false,"prerelease": false}'
curl -H "$AUTH" \
    -H "Content-Type: application/json" \
    -d "$JSON" \
    $GH_REPO/releases
