ARG GOLANG_VERSION=1.25.3

# TARGETOS and TARGETARCH will be populated at 'docker build'
ARG TARGETOS  
ARG TARGETARCH

ARG COMMIT
ARG VERSION

FROM --platform=$TARGETARCH docker.io/golang:${GOLANG_VERSION} AS build

WORKDIR /crtsh-exporter

COPY go.* ./
COPY main.go .
COPY collector ./collector

# TARGETOS and TARGETARCH will be populated at 'docker build'
ARG TARGETOS  
ARG TARGETARCH
 
ARG VERSION
ARG COMMIT

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build \
    -ldflags "-X main.OSVersion=${VERSION} -X main.GitCommit=${COMMIT}" \
    -a -installsuffix cgo \
    -o /go/bin/exporter \
    ./main.go

FROM --platform=$TARGETARCH gcr.io/distroless/static-debian12:latest

LABEL org.opencontainers.image.description="Prometheus Exporter for crt.sh"
LABEL org.opencontainers.image.source=https://github.com/DazWilkin/crtsh-exporter

COPY --from=build /go/bin/exporter /

EXPOSE 8080

ENTRYPOINT ["/exporter"]
CMD ["--endpoint=0.0.0.0:8080","--hosts=","--path=/metrics"]
