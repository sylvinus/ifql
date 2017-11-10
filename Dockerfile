FROM gliderlabs/alpine
RUN apk add --update ca-certificates tzdata && \
    rm /var/cache/apk/*
EXPOSE 8093/tcp
COPY ifqld /
COPY LICENSE /
COPY README.md /
ENTRYPOINT ["/ifqld"]
