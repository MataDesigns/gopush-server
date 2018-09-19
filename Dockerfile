FROM alpine:latest

ADD release/linux/amd64/gopushserver /bin/

RUN mkdir /db
RUN mkdir /keys

EXPOSE 80

HEALTHCHECK --start-period=2s --interval=10s --timeout=5s \
  CMD ["/bin/gopushserver", "--ping"]

ENTRYPOINT ["/bin/gopushserver", "--prod", "-p", "80"]
