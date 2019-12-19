#!/usr/bin/env bash

LINUX_KERNEL_VERSION="v4.18"

CURR_DIR=`pwd`

SCRIPT_DIR="`dirname "$0"`"

ENV_DIRECTORY="`readlink -f "$SCRIPT_DIR/../"`"

WORK_DIR="$ENV_DIRECTORY/workspace"

cd $WORK_DIR

git clone https://github.com/torvalds/linux.git linux-kernel-$LINUX_KERNEL_VERSION
cd linux-kernel-$LINUX_KERNEL_VERSION

git checkout $LINUX_KERNEL_VERSION

make -j$(nproc) vmlinux

cd $CURR_DIR
