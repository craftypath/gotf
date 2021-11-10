FROM alpine:3.14

COPY gotf /usr/local/bin/gotf

RUN gotf --version
