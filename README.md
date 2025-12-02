# Prometheus Exporter for [`crt.sh`](https://crt.sh)

[![build](https://github.com/DazWilkin/crtsh-exporter/actions/workflows/multi-platform-docker-build.yml/badge.svg)](https://github.com/DazWilkin/crtsh-exporter/actions/workflows/multi-platform-docker-build.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/DazWilkin/crtsh-exporter.svg)](https://pkg.go.dev/github.com/DazWilkin/crtsh-exporter)
[![Go Report Card](https://goreportcard.com/badge/github.com/DazWilkin/crtsh-exporter)](https://goreportcard.com/report/github.com/DazWilkin/crtsh-exporter)

+ `ghcr.io/dazwilkin/crtsh-exporter:a4a2214`

With thanks to [Oğuzhan Yilmaz](https://github.com/oguzhan-yilmaz), GitHub Actions now supports mutli-platform (`linux/amd64`,`linux/arm64`,`linux/arm64/v7`) builds.

## Example

```bash
HOST="..."

curl \
--silent \
--get \
--data-urlencode "q=${HOST}" \
--data-urlencode "output=json" \
https://crt.sh
```
Returns
```JSON
[
    {
        "issuer_ca_id": 123456,
        "issuer_name": "C=US, O=Let's Encrypt, CN=R3",
        "common_name": "{HOST}",
        "name_value": "{HOST}",
        "id": 10123456789,
        "entry_timestamp": "2023-01-01T23:59:59.000",
        "not_before": "2023-01-01T23:59:59",
        "not_after": "2023-01-01T23:59:59",
        "serial_number": "123456789abcdef0123456789abcdef0"
    }
]
```

## Run

```bash
HOSTS="{host1}.{domain1},{host2}.{domain2},..."

HOST_PORT="8080"
CONT_PORT="8080"

podman run \
--interactive --tty --rm \
--name=crtsh-exporter \
--publish=${HOST_PORT}:${CONT_PORT}/tcp \
ghcr.io/dazwilkin/crtsh-exporter:a4a2214 \
--hosts=${HOSTS} \
--endpoint=:${CONT_PORT} \
--path=/metrics
```

## Prometheus

```bash
VERS="v2.46.0"

# Binds to host network to scrape crt.sh Exporter
podman run \
--interactive --tty --rm \
--net=host \
--volume=${PWD}/prometheus.yml:/etc/prometheus/prometheus.yml \
--volume=${PWD}/rules.yml:/etc/alertmanager/rules.yml \
quay.io/prometheus/prometheus:${VERS} \
  --config.file=/etc/prometheus/prometheus.yml \
  --web.enable-lifecycle
```

## Metrics

|Name|Type|Description|
|----|----|-----------|
|`crtsh_exporter_build_info`|Counter|A metric with a constant '1' value|
|`crtsh_exporter_certificate_expiry`|Gauge|Expiration ("not after") timestamp of most recent record|
|`crtsh_exporter_certificate_records`|Gauge|Number of Certificate records, labeled by most recent record's metadata|
|`crtsh_exporter_start_time`|Gauge|Exporter start time in UNIX epoch|

## [Sigstore](https://www.sigstore.dev/)

`crtsh-exporter` container images are signed by [Sigstore](https://www.sigstore.dev/) and may be verified:

```bash
cosign verify \
--key=./cosign.pub \
ghcr.io/dazwilkin/crtsh-exporter:a4a2214
```

> **NOTE** `cosign.pub` may be downloaded [here](./cosign.pub)

To install `cosign` e.g.:

```bash
go install github.com/sigstore/cosign/cmd/cosign@latest
```

## Similar Exporters

+ [Prometheus Exporter for Azure](https://github.com/DazWilkin/azure-exporter)
+ [Prometheus Exporter for crt.sh](https://github.com/DazWilkin/crtsh-exporter)
+ [Prometheus Exporter for Fly.io](https://github.com/DazWilkin/fly-exporter)
+ [Prometheus Exporter for GoatCounter](https://github.com/DazWilkin/goatcounter-exporter)
+ [Prometheus Exporter for Google Cloud](https://github.com/DazWilkin/gcp-exporter)
+ [Prometheus Exporter for Koyeb](https://github.com/DazWilkin/koyeb-exporter)
+ [Prometheus Exporter for Linode](https://github.com/DazWilkin/linode-exporter)
+ [Prometheus Exporter for PorkBun](https://github.com/DazWilkin/porkbun-exporter)
+ [Prometheus Exporter for updown.io](https://github.com/DazWilkin/updown-exporter)
+ [Prometheus Exporter for Vultr](https://github.com/DazWilkin/vultr-exporter)

<hr/>
<br/>
<a href="https://www.buymeacoffee.com/dazwilkin" target="_blank"><img src="https://cdn.buymeacoffee.com/buttons/default-orange.png" alt="Buy Me A Coffee" height="41" width="174"></a>
