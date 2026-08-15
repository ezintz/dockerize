FROM alpine:3.20

LABEL org.opencontainers.image.source="https://github.com/ezintz/dockerize"

ARG TARGETPLATFORM
COPY $TARGETPLATFORM/dockerize /usr/local/bin

ENTRYPOINT ["dockerize"]
CMD ["--help"]
