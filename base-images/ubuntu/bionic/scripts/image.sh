#! /bin/bash
set -ex

cp /root/linux-source-$KERNEL_SOURCE_VERSION/vmlinux /output/kernel.elf
cp /root/linux-source-$KERNEL_SOURCE_VERSION/.config /output/config

##dd if=/dev/zero of=/output/rootfs.ext4 bs=1M count=500
##truncate -s 1G /output/rootfs.ext4
##mkfs.ext4 /output/rootfs.ext4

##mount -t ext4 /output/rootfs.ext4 /rootfs
if [[ $INCLUDE_CLOUD_INIT == "true" ]]; then
  debootstrap --include rsync,python3,python3-cffi-backend,python3-cryptography,python3-oauthlib,openssh-server,netplan.io,nano,cloud-init bionic /rootfs http://archive.ubuntu.com/ubuntu/
else
  debootstrap --include rsync,openssh-server,netplan.io,nano bionic /rootfs http://archive.ubuntu.com/ubuntu/
fi

mount --bind / /rootfs/mnt

chroot /rootfs /bin/bash /mnt/scripts/provision.sh

umount /rootfs/mnt

##umount /rootfs

echo "init=/bin/systemd noapic reboot=k panic=1 pci=off nomodules console=ttyS0" > /output/boot


##cd /output
##tar czvf ubuntu-bionic.tar.gz rootfs.ext4 kernel.elf config boot
##cd /
