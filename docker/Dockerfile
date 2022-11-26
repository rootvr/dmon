# FROM golang:alpine3.13

FROM golang:alpine

LABEL maintainers="Valentino Di Giosaffatte, Riccardo Armando Di Prinzio"

RUN apk update && \
    apk add redis tshark make

RUN addgroup root wireshark && \
    chgrp root /usr/bin/dumpcap && \
    chmod 750 /usr/bin/dumpcap

COPY docker/entrypoint.sh /entrypoint.sh
RUN chmod 755 /entrypoint.sh

COPY . /go/src/dmon
WORKDIR /go/src/dmon

RUN make clean && \
    make deps && \
    make build && \
    go install

WORKDIR /

ENTRYPOINT ["/entrypoint.sh"]
