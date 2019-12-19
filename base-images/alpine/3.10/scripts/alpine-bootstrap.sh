#!/bin/sh

apk upgrade --update-cache --available
apk add linux-virt
apk add openrc
apk add util-linux
apk add rsync
apk add python3
ln -s agetty /etc/init.d/agetty.ttyS0
echo ttyS0 > /etc/securetty
rc-update add agetty.ttyS0 default

# Make sure special file systems are mounted on boot:
rc-update add devfs boot
rc-update add procfs boot
rc-update add sysfs boot

# Then, copy the newly configured system to the rootfs image:
rsync -ahPHAXx --delete --exclude=/dev/* --exclude=/proc/* --exclude=/sys/* --exclude=/tmp/* --exclude=/run/* --exclude=/mnt/* --exclude=/media/* --exclude=/lost+found --exclude=/rootfs --exclude=/output / /rootfs
cp /boot/vmlinuz-virt /output/kernel.elf
cp /boot/config-virt /output/config
echo "noapic reboot=k panic=1 pci=off nomodules console=ttyS0" > /output/boot
