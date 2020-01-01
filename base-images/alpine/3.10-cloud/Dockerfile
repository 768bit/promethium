FROM alpine:3.10

ENV KERNEL_SOURCE_VERSION 4.18

WORKDIR /root

VOLUME [ "/output", "/rootfs" ]

ADD scripts /scripts

WORKDIR /root

CMD [ "/bin/sh", "/scripts/alpine-bootstrap-cloud.sh" ]
