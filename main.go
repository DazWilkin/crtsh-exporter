package main

import (
	"flag"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/DazWilkin/crtsh-exporter/collector"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	sRoot string = `
<h2>A Prometheus Exporter for <a href="https://crt.sh">Certificate Search</a></h2>
<ul>
	<li><a href="{{ .Metrics }}">metrics</a></li>
	<li><a href="/healthz">healthz</a></li>
</ul>`
)

var (
	// GitCommit is the git commit value and is expected to be set during build
	GitCommit string
	// GoVersion is the Golang runtime version
	GoVersion = runtime.Version()
	// OSVersion is the OS version (uname --kernel-release) and is expected to be set during build
	OSVersion string
	// StartTime is the start time of the exporter represented as a UNIX epoch
	StartTime = time.Now().Unix()
)
var (
	endpoint    = flag.String("endpoint", ":8080", "The endpoint of the Exporter HTTP server")
	domain      = flag.String("domain", "", "The domain name to be queried")
	hostList    = flag.String("hosts", "", "Comma-separated list of hosts to be queried")
	metricsPath = flag.String("path", "/metrics", "The path on which Prometheus metrics will be served")
)
var (
	tRoot = template.Must(template.New("root").Parse(sRoot))
)

func handleHealthz(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	// nolint:errcheck
	w.Write([]byte("ok"))
}
func handleRoot(w http.ResponseWriter, _ *http.Request) {
	// nolint:errcheck
	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	// nolint:errcheck
	fmt.Fprint(w)
	// nolint:errcheck
	tRoot.Execute(w, struct {
		Metrics string
	}{
		Metrics: *metricsPath,
	})
}
func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	if GitCommit == "" {
		logger.Info("GitCommit value unchanged: expected to be set during build")
	}
	if OSVersion == "" {
		logger.Info("OSVersion value unchanged: expected to be set during build")
	}

	registry := prometheus.NewRegistry()
	registry.MustRegister(collector.NewExporterCollector(OSVersion, GoVersion, GitCommit, StartTime))

	flag.Parse()

	if *hostList != "" {
		if hosts := strings.Split(*hostList, ","); len(hosts) > 0 {
			logger.Info("Flag --hosts set and contains hosts, registering Hosts collector",
				"hosts", hosts,
			)
			registry.MustRegister(collector.NewHostsCollector(hosts, logger))
		}
	}

	if *domain != "" {
		logger.Info("Flag --domain set and contains domain, registering Domain collector")
		registry.MustRegister(collector.NewDomainCollector(*domain, logger))
	}

	mux := http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(handleRoot))
	mux.Handle("/healthz", http.HandlerFunc(handleHealthz))
	mux.Handle(*metricsPath, promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))

	logger.Info("Server starting",
		"endpoint", *endpoint,
	)
	logger.Info("metrics path",
		"path", *metricsPath,
	)
	logger.Error("Server failed", "err", http.ListenAndServe(*endpoint, mux))
}
