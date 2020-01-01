#!/bin/sh

apk upgrade --update-cache --available
apk add git
apk add alpine-sdk
apk add bison
apk add flex
apk add g++
apk add linux-virt-dev
apk add linux-vanilla-dev
apk add linux-headers
apk add openssl-dev

git clone https://github.com/torvalds/linux.git linux-source-$KERNEL_SOURCE_VERSION

cp /root/kernel-config /root/linux-source-$KERNEL_SOURCE_VERSION/.config

cd /root/linux-source-$KERNEL_SOURCE_VERSION
git checkout v$KERNEL_SOURCE_VERSION
yes '' | make oldconfig  && make -j $(nproc) vmlinux

echo "{\"machine\": \"$(uname -m)\", \"platform\": \"$(uname -i)\", \"version\": \"${KERNEL_SOURCE_VERSION}\", \"from\": \"github.com/torvalds/linux.git\"}" > /output/_kernel_meta.json
cp /root/linux-source-$KERNEL_SOURCE_VERSION/vmlinux /output/kernel.elf
cp /root/linux-source-$KERNEL_SOURCE_VERSION/.config /output/config
