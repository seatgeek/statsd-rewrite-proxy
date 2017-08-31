FROM alpine

RUN apk --update upgrade && \
    apk add curl ca-certificates && \
    update-ca-certificates && \
    rm -rf /var/cache/apk/*

ADD ./build/statsd-rewrite-proxy-linux-amd64 /statsd-rewrite-proxy

ENTRYPOINT ["/statsd-rewrite-proxy"]
