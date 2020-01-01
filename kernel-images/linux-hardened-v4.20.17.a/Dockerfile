FROM alpine:3.10

ENV KERNEL_SOURCE_VERSION 4.20.17.a

WORKDIR /root

VOLUME [ "/output", "/rootfs" ]

ADD config/kernel-config /root/kernel-config

ADD scripts /scripts

WORKDIR /root

CMD [ "/bin/sh", "/scripts/kernel-build.sh" ]
