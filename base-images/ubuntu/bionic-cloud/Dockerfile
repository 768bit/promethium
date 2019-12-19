FROM ubuntu:bionic

ENV DEBIAN_FRONTEND noninteractive
ENV KERNEL_SOURCE_VERSION 4.18.0
ENV INCLUDE_CLOUD_INIT true

WORKDIR /root

RUN apt-get update && apt-get upgrade -y && apt-get install -y debootstrap build-essential kernel-package \
  fakeroot linux-source-$KERNEL_SOURCE_VERSION bc kmod cpio flex cpio bison libncurses5-dev libelf-dev libssl-dev && \
  tar xvf /usr/src/linux-source-$KERNEL_SOURCE_VERSION.tar.*

ADD config/kernel-config /root/linux-source-$KERNEL_SOURCE_VERSION/.config

WORKDIR /root/linux-source-$KERNEL_SOURCE_VERSION
RUN yes '' | make oldconfig  && \
  make -j $(nproc) deb-pkg && make clean && make -j $(nproc) vmlinux
WORKDIR /root

VOLUME [ "/output", "/rootfs" ]

ADD scripts /scripts
ADD config /config

CMD [ "/bin/bash", "/scripts/image.sh" ]
