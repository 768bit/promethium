#!/usr/bin/env bash

ALPINE_VERSION="3.10"
if [ -z "$2" ]; then
  ALPINE_VERSION="$2"
fi

CURR_DIR=`pwd`

SCRIPT_DIR="`dirname "$0"`"

ENV_DIRECTORY="`readlink -f "$SCRIPT_DIR/../"`"

WORK_DIR="$1"

MOUNT_POINT="$3"

#rm -rf $WORK_DIR
#mkdir -p $WORK_DIR

##MNT_POINT=`mktemp -d`

##dd if=/dev/zero of=rootfs.ext4 bs=1M count=50
##mkfs.ext4 rootfs.ext4

##sudo mount rootfs.ext4 $MNT_POINT


docker run --rm -v $MOUNT_POINT:/rootfs -v $WORK_DIR:/output -v $ENV_DIRECTORY/scripts/support/alpine:/support alpine:$ALPINE_VERSION /bin/sh "/support/alpine-bootstrap.sh"

###sudo umount $MNT_POINT
cd $CURR_DIR
