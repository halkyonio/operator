#!/bin/bash

# this script assumes that runs on linux

BIN_DIR="./dist/bin/"
RELEASE_DIR="./dist/release/"
APP="component-operator"

mkdir -p $RELEASE_DIR

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