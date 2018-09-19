FROM alpine:latest

RUN apk --update upgrade && \
    apk add curl ca-certificates && \
    update-ca-certificates && \
    rm -rf /var/cache/apk/*

ADD release/linux/amd64/gopushserver /bin/

RUN mkdir /db
RUN mkdir /keys

EXPOSE 80

HEALTHCHECK --start-period=2s --interval=10s --timeout=5s \
  CMD ["/bin/gopushserver", "--ping"]

ENTRYPOINT ["/bin/gopushserver", "--prod", "-p", "80"]
