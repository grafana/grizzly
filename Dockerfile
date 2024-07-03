FROM golang:1.22.5
ARG JSONNET_VERSION=v0.19.1
ARG JB_VERSION=v0.5.1
RUN CGO_ENABLED=0 go install github.com/google/go-jsonnet/cmd/jsonnet@${JSONNET_VERSION} && \
  CGO_ENABLED=0 go install github.com/google/go-jsonnet/cmd/jsonnetfmt@${JSONNET_VERSION} && \
  CGO_ENABLED=0 go install github.com/jsonnet-bundler/jsonnet-bundler/cmd/jb@${JB_VERSION}

FROM alpine
COPY  ./grr /usr/local/bin/grr
COPY --from=0 /go/bin/jsonnet /usr/local/bin/jsonnet
COPY --from=0 /go/bin/jsonnetfmt /usr/local/bin/jsonnetfmt
COPY --from=0 /go/bin/jb /usr/local/bin/jb
ENTRYPOINT ["/usr/local/bin/grr"]
