package collector

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	baseURL          string = "https://crt.sh"
	expiryDateLayout string = "2006-01-02T15:04:05.999" // Reference time
)

// Record is a type that represents a crt.sh result record
type Record struct {
	IssueCAID      int    `json:"issuer_ca_id"`
	IssuerName     string `json:"issuer_name"`
	CommonName     string `json:"common_name"`
	NameValue      string `json:"name_value"`
	ID             int    `json:"id"`
	EntryTimestamp Time   `json:"entry_timestamp"`
	NotBefore      Time   `json:"not_before"`
	NotAfter       Time   `json:"not_after"`
	SerialNumber   string `json:"serial_number"`
}

// Time provides an alternative JSON unmarshaler for string-encoded dates
// Dates are returned either as 2006-01-02T15:04:05.999 or without milliseconds
type Time struct {
	time.Time
}

// UnmarshalJSON is a method that converts a string-encoded date into a Time
func (t *Time) UnmarshalJSON(b []byte) (err error) {
	// []byte is escaped: \"2006-01-02T15:04:05.999\"
	s := strings.Trim(string(b), "\"")

	if s == "null" {
		t.Time = time.Time{}
		return err
	}

	t.Time, err = time.Parse(expiryDateLayout, s)

	return err
}

// HostsCollector is a type that represents a collector of host records
type HostsCollector struct {
	client *http.Client
	hosts  []string
	log    *slog.Logger

	// Metrics
	CertificateRecords *prometheus.Desc
	CertificateExpiry  *prometheus.Desc
}

// NewHostsCollector is a function that returns a new HostCollector
func NewHostsCollector(hosts []string, log *slog.Logger) *HostsCollector {
	log = log.With("function", "HostsCollector")
	client := &http.Client{}
	return &HostsCollector{
		client: client,
		hosts:  hosts,
		log:    log,

		CertificateRecords: prometheus.NewDesc(
			BuildFQName("certificate_records", log),
			"Number of Certificate records, labeled by most recent record's metadata",
			[]string{
				"name",
				"not_before",
				"not_after",
				"serial_number",
			},
			nil,
		),
		CertificateExpiry: prometheus.NewDesc(
			BuildFQName("certificate_expiry", log),
			"Expiration (\"not after\") timestamp of most recent record",
			[]string{
				"name",
				"serial_number",
			},
			nil,
		),
	}
}

// Collect implements Prometheus' Collector interface and is used to collect metrics
func (c *HostsCollector) Collect(ch chan<- prometheus.Metric) {
	log := c.log.With("method", "Collect")

	for _, host := range c.hosts {
		log := log.With("host", host)
		log.Info("Creating HTTP request")
		rqst, err := http.NewRequest(http.MethodGet, baseURL, nil)
		if err != nil {
			msg := "unable to create request"
			log.Info(msg)
			continue
		}

		log.Info("Building URL query string")
		q := rqst.URL.Query()
		q.Add("q", host)
		q.Add("exclude", "expired")
		q.Add("deduplicate", "Y")
		q.Add("output", "json")
		rqst.URL.RawQuery = q.Encode()

		log.Info("Executing HTTP request")
		resp, err := c.client.Do(rqst)
		if err != nil {
			msg := "unable to execute request"
			log.Info(msg)
			continue
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				log.Error("unable to close response body", "error", err)
			}
		}()

		if resp.StatusCode != http.StatusOK {
			msg := "non-200 response"
			log.Info(msg,
				"code", resp.StatusCode,
			)
			continue
		}

		log.Info("Reading HTTP response")
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			msg := "unable to read response body"
			log.Info(msg)
			continue
		}

		// Debugging
		// log.Info("Response",
		// 	"body", string(b),
		// )

		records := []Record{}
		if err := json.Unmarshal(b, &records); err != nil {
			msg := "unable to unmarshal response"
			log.Info(msg)
			continue
		}

		log.Info("Records",
			"number", len(records),
		)
		if len(records) == 0 {
			msg := "no records found in response"
			log.Info(msg)
			continue
		}

		// Process all records
		for _, record := range records {
			log.Info("Processing record",
				"record", record,
			)

			// Create metric for each record
			ch <- prometheus.MustNewConstMetric(
				c.CertificateRecords,
				prometheus.GaugeValue,
				1.0, // Changed to 1.0 since we're counting individual records
				[]string{
					record.NameValue,
					strconv.FormatInt(record.NotBefore.Unix(), 10),
					strconv.FormatInt(record.NotAfter.Unix(), 10),
					record.SerialNumber,
				}...,
			)

			ch <- prometheus.MustNewConstMetric(
				c.CertificateExpiry,
				prometheus.GaugeValue,
				float64(record.NotAfter.Unix()),
				[]string{
					record.NameValue,
					record.SerialNumber,
				}...,
			)
		}
	}
}

// Describe implements Prometheus' Collector interface and is used to describe metrics
func (c *HostsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.CertificateRecords
	ch <- c.CertificateExpiry
}
