FROM alpine:3.10

WORKDIR /root

VOLUME [ "/output", "/rootfs" ]

ADD scripts /scripts

CMD [ "/bin/sh", "/scripts/alpine-bootstrap.sh" ]
