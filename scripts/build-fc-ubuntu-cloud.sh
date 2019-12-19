#!/usr/bin/env bash

CURR_DIR=`pwd`

SCRIPT_DIR="`dirname "$0"`"

ENV_DIRECTORY="`readlink -f "$SCRIPT_DIR/../"`"

WORK_DIR="$1"

MOUNT_POINT="$3"

cd $ENV_DIRECTORY/scripts/support/ubuntu/bionic
docker build -t ubuntu-firecracker .
#rm -rf $WORK_DIR
#mkdir -p $WORK_DIR

docker run --privileged --rm -e "INCLUDE_CLOUD_INIT=true" -v $WORK_DIR:/output -v $MOUNT_POINT:/rootfs ubuntu-firecracker

cd $CURR_DIR
