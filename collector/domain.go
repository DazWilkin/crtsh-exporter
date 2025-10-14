package collector

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
)

// const (
// 	baseURL          string = "https://crt.sh"
// 	expiryDateLayout string = "2006-01-02T15:04:05.999" // Reference time
// )

// type Record

// type Time

// DomainCollector is a type that represents a collector of certificates by domain
type DomainCollector struct {
	client *http.Client
	domain string
	log    *slog.Logger

	// Metrics
	CertificateRecords *prometheus.Desc
	CertificateExpiry  *prometheus.Desc
}

// NewDomainCollector is a function that returns a new DomainCollector
func NewDomainCollector(domain string, log *slog.Logger) *DomainCollector {
	log = log.With("function", "DomainCollector")
	client := &http.Client{}
	return &DomainCollector{
		client: client,
		domain: domain,
		log:    log,

		CertificateRecords: prometheus.NewDesc(
			BuildFQName("certificate_records", log),
			"Number of Certificate records, labeled by most recent record's metadata",
			[]string{
				"domain",
			},
			nil,
		),
		CertificateExpiry: prometheus.NewDesc(
			BuildFQName("certificate_expiry", log),
			"Expiration (\"not after\") timestamp of most recent record",
			[]string{
				"domain",
				"host",
				"not_before",
				"not_after",
				"serial_number",
			},
			nil,
		),
	}
}

// Collect implements Prometheus' Collector interface and is used to collect metrics
func (c *DomainCollector) Collect(ch chan<- prometheus.Metric) {
	log := c.log.With("method", "Collect")

	log = log.With("domain", c.domain)
	log.Info("Creating HTTP request")
	rqst, err := http.NewRequest(http.MethodGet, baseURL, nil)
	if err != nil {
		msg := "unable to create request"
		log.Info(msg)
		return
	}

	log.Info("Building URL query string")
	q := rqst.URL.Query()
	q.Add("q", c.domain)
	q.Add("exclude", "expired")
	q.Add("deduplicate", "Y")
	q.Add("output", "json")
	rqst.URL.RawQuery = q.Encode()

	log.Info("Executing HTTP request")
	resp, err := c.client.Do(rqst)
	if err != nil {
		msg := "unable to execute request"
		log.Info(msg)
		return
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Error("unable to close response body", "err", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		msg := "non-200 response"
		log.Info(msg,
			"code", resp.StatusCode,
		)
		return
	}

	log.Info("Reading HTTP response")
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		msg := "unable to read response body"
		log.Info(msg)
		return
	}

	// Debugging
	// log.Info("Response",
	// 	"body", string(b),
	// )

	records := []Record{}
	if err := json.Unmarshal(b, &records); err != nil {
		msg := "unable to unmarshal response"
		log.Info(msg)
		return
	}

	log.Info("Records",
		"number", len(records),
	)
	if len(records) == 0 {
		msg := "expected at least one record in response"
		log.Info(msg)
		return
	}

	ch <- prometheus.MustNewConstMetric(
		c.CertificateRecords,
		prometheus.GaugeValue,
		float64(len(records)),
		[]string{
			c.domain,
		}...,
	)

	d, err := NewDomain(c.domain)
	if err != nil {
		msg := "unable to create domain"
		log.Info(msg)
		return
	}

	for _, record := range records {
		log.Info("Most recent record",
			"record", record,
		)

		host, err := d.Hostname(record.CommonName)
		if err != nil {
			msg := "unable to get host name"
			log.Info(msg)
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			c.CertificateExpiry,
			prometheus.GaugeValue,
			float64(record.NotAfter.Unix()),
			[]string{
				c.domain,
				host,
				strconv.FormatInt(record.NotBefore.Unix(), 10),
				strconv.FormatInt(record.NotAfter.Unix(), 10),
				record.SerialNumber,
			}...,
		)
	}
}

// Describe implements Prometheus' Collector interface and is used to describe metrics
func (c *DomainCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.CertificateRecords
	ch <- c.CertificateExpiry
}
