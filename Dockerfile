FROM gliderlabs/alpine
RUN apk add --no-cache ca-certificates tzdata
EXPOSE 8093/tcp
COPY ifqld /
COPY LICENSE /
COPY README.md /
ENTRYPOINT ["/ifqld"]
