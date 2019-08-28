#!/bin/bash

set -e

CONFIG=$@

for line in $CONFIG; do
  eval "$line"
done

owner="halkyonio"
repo="operator"

AUTH="Authorization: token $GITHUB_API_TOKEN"

GH_API="https://api.github.com"
GH_REPO="$GH_API/repos/$owner/$repo"

WGET_ARGS="--content-disposition --auth-no-challenge --no-cookie"
CURL_ARGS="-LJO#"

TAG="v$VERSION"
GH_TAGS="$GH_REPO/releases/tags/$TAG"

BIN_DIR="./build/_output/bin/"
RELEASE_DIR="./build/_output/bin/release/"
APP="component-operator"

TEMP_BRANCH="update-operator-yaml-${TAG}"

git config user.email "snow-bot@snowdrop.me"
git config user.name "Snow-bot"
# revert CircleCI global setting
git config --remove-section url.ssh://git@github.com 2> /dev/null || true
# ensure that the CI user can make changes
git remote set-url origin https://${GITHUB_USER}:${GITHUB_API_TOKEN}@github.com/halkyonio/operator.git

# "stash" the build output
mkdir -p /tmp/operator
cp -R build/_output /tmp/operator/

git checkout -b ${TEMP_BRANCH}

# "unstash" the build output
cp -R /tmp/operator/_output/ ./build/

echo "update all necessary operator related files"
sed -i "s/:latest/:${TAG}/g" ./deploy/operator.yaml
sed -i "s/:latest/:${TAG}/g" ./deploy/olm-catalog/bundle/halkyon.v0.0.0.clusterserviceversion.yaml
sed -i "s/v0.0.0/${TAG}/g" ./deploy/olm-catalog/bundle/halkyon.v0.0.0.clusterserviceversion.yaml
mv ./deploy/olm-catalog/bundle/halkyon.v0.0.0.clusterserviceversion.yaml ./deploy/olm-catalog/bundle/halkyon.${TAG}.clusterserviceversion.yaml
sed -i "s/v0.0.0/${TAG}/g" ./deploy/olm-catalog/bundle/halkyon.package.yaml

echo "commit, push and tag"
git add .
git commit -am "[ci skip] Update operator deployment file to ${TAG}"
git push --set-upstream origin ${TEMP_BRANCH}
git tag -a $TAG -m "$TAG release"

echo "Create Release for tag $TAG"
JSON='{"tag_name": "'"$TAG"'","target_commitish": "'"$TEMP_BRANCH"'","name": "'"$TAG"'","body": "'"$TAG"'-release","draft": false,"prerelease": false}'

curl -H "$AUTH" \
  -H "Content-Type: application/json" \
  -d "$JSON" \
  $GH_REPO/releases

# Read asset tags.
response=$(curl -sH "$AUTH" $GH_TAGS)

# Get ID of the asset based on given filename.
eval $(echo "$response" | grep -m 1 "id.:" | grep -w id | tr : = | tr -cd '[[:alnum:]]=')
[ "$id" ] || {
  echo "Error: Failed to get release id for tag: $tag"
  echo "$response" | awk 'length($0)<100' >&2
  exit 1
}
echo "ID : $id"

# Upload assets compressed and non compressed file
for arch in $(ls -1 $BIN_DIR/); do

  target_file=${RELEASE_DIR}/${APP}-${arch}
  target_tgz_file=${target_file}.tgz
  array=(${target_file} ${target_tgz_file})
  for i in "${array[@]}"; do
    target_file=$i
    content_type=$(file -b --mime-type ${target_file})
    echo "Asset : ${target_file}"
    echo "Content type : $content_type"

    GH_ASSET_URL="https://uploads.github.com/repos/$owner/$repo/releases/$id/assets?name=$(basename ${target_file})"
    echo "GH_ASSET_URL : $GH_ASSET_URL"

    curl -i -H "Authorization: token $GITHUB_API_TOKEN" \
      -H "Accept: application/vnd.github.v3+json" \
      -H "Content-Type: $content_type" \
      -X POST \
      --data-binary @${target_file} \
      $GH_ASSET_URL
  done
done

# delete the temporary branch
git push --delete origin ${TEMP_BRANCH}
