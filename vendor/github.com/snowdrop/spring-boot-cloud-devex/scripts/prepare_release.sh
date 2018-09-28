#!/bin/bash

# this script assumes that runs on linux

BIN_DIR="./dist/bin/"
RELEASE_DIR="./dist/release/"
APP="sb"

mkdir -p $RELEASE_DIR

# if this is run on travis make sure that binary was build with corrent version
if [[ -n $TRAVIS_TAG ]]; then
    echo "Checking if app's version was set to the same version as current tag"
    # use sed to get only semver part
    bin_version=$(${BIN_DIR}/linux-amd64/sb version | head -1 | sed "s/^sb \(.*\) (.*)$/\1/")
    if [ "$TRAVIS_TAG" == "${bin_version}" ]; then
        echo "OK: app's  version output is matching current tag"
    else
        echo "ERR: TRAVIS_TAG ($TRAVIS_TAG) is not matching 'sb version' (v${bin_version})"
        exit 1
    fi
fi

# gziped binaries
for arch in `ls -1 $BIN_DIR/`;do
    suffix=""
    if [[ $arch == windows-* ]]; then
        suffix=".exe"
    fi
    source_file=$BIN_DIR/$arch/$APP$suffix
    target_file=$APP-$arch$suffix

    # Move binaries to the release directory
    echo "copying binary $source_file to release directory"
    cp $source_file $RELEASE_DIR/$target_file

    echo "Make bin generated executable"
    chmod +x $RELEASE_DIR/$target_file

    # Create a gzip of the binary
    echo "tar compress the $target_file as $target_file.tgz"
    cd $RELEASE_DIR/
    tar -czf $target_file.tgz $target_file
    cd ../..
done