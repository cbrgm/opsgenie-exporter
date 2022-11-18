package main

import (
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/alecthomas/kong"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	levelDebug = "debug"
	levelInfo  = "info"
	levelWarn  = "warn"
	levelError = "error"
)

var (
	// Version of opsgenie-exporter.
	Version string
	// Revision or Commit this binary was built from.
	Revision string
	// GoVersion running this binary.
	GoVersion = runtime.Version()
	// StartTime has the time this was started.
	StartTime = time.Now()
)

const (
	namespace = "opsgenie"

	// subsystems
	alertSubsystem = "alerts"
	teamSubsystem  = "teams"
	userSubsystem  = "users"

	// labels
	labelPriority = "priority"
	labelStatus   = "status"
	labelTeam     = "team"
	labelUserRole = "role"
)

var config struct {
	WebAddr        string `name:"http.addr" default:"0.0.0.0:9212" help:"The address the exporter is running on"`
	WebPath        string `name:"http.path" default:"/metrics" help:"The path metrics will be exposed at"`
	LogJSON        bool   `name:"log.json" default:"false" help:"Tell the exporter to log json and not key value pairs"`
	LogLevel       string `name:"log.level" default:"info" enum:"error,warn,info,debug" help:"The log level to use for filtering logs"`
	OpsgenieApiKey string `required:"true" name:"opsgenie.apikey" help:"The opsgenie api token"`
}

func main() {

	_ = kong.Parse(&config,
		kong.Name("opsgenie-exporter"),
	)

	levelFilter := map[string]level.Option{
		levelError: level.AllowError(),
		levelWarn:  level.AllowWarn(),
		levelInfo:  level.AllowInfo(),
		levelDebug: level.AllowDebug(),
	}

	logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	if config.LogJSON {
		logger = log.NewJSONLogger(log.NewSyncWriter(os.Stderr))
	}

	logger = level.NewFilter(logger, levelFilter[config.LogLevel])
	logger = log.With(logger,
		"ts", log.DefaultTimestampUTC,
		"caller", log.DefaultCaller,
	)

	opsgenieClient, err := NewOpsgenieClient(config.OpsgenieApiKey)
	if err != nil {
		level.Error(logger).Log("msg", "failed to initialize opsgenie client", "err", err)
		os.Exit(1)
	}

	reg := prometheus.NewRegistry()
	reg.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		&opsgenieCollector{
			client: opsgenieClient,
			logger: logger,
			OpsgenieAlertMetricsCreatedTotal: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, alertSubsystem, "created_total"),
				"opsgenie alert metrics",
				[]string{labelTeam, labelPriority},
				nil,
			),
			OpsgenieAlertMetricsCount: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, alertSubsystem, "status_count"),
				"opsgenie alert metrics",
				[]string{labelStatus, labelTeam, labelPriority},
				nil,
			),
			OpsgenieTeamMetricsCount: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, teamSubsystem, "count"),
				"opsgenie team metrics",
				nil,
				nil,
			),
			OpsgenieUserMetricsCount: prometheus.NewDesc(
				prometheus.BuildFQName(namespace, userSubsystem, "count"),
				"opsgenie user metrics",
				[]string{labelUserRole},
				nil,
			),
		},
	)

	http.Handle(config.WebPath,
		promhttp.HandlerFor(reg, promhttp.HandlerOpts{}),
	)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`<html>
			<head><title>Opsgenie Exporter</title></head>
			<body>
			<h1>Opsgenie Exporter</h1>
			<p><a href="` + config.WebPath + `">see metrics</a></p>
			</body>
			</html>`))
	})

	level.Info(logger).Log("msg", "listening", "addr", config.WebAddr)
	if err := http.ListenAndServe(config.WebAddr, nil); err != nil {
		level.Error(logger).Log("msg", "http listenandserve error", "err", err)
		os.Exit(1)
	}
}

type opsgenieCollector struct {
	client *OpsgenieClient
	logger log.Logger

	// metrics
	OpsgenieAlertMetricsCreatedTotal *prometheus.Desc
	OpsgenieAlertMetricsCount        *prometheus.Desc
	OpsgenieTeamMetricsCount         *prometheus.Desc
	OpsgenieUserMetricsCount         *prometheus.Desc
}

func (c *opsgenieCollector) Describe(descs chan<- *prometheus.Desc) {
	descs <- c.OpsgenieAlertMetricsCreatedTotal
	descs <- c.OpsgenieAlertMetricsCount
	descs <- c.OpsgenieTeamMetricsCount
	descs <- c.OpsgenieUserMetricsCount
}

func (c *opsgenieCollector) Collect(metrics chan<- prometheus.Metric) {
	level.Debug(c.logger).Log("msg", "scraping opsgenie api")
	c.collectOpsgenieAlertMetrics(metrics)
	c.collectOpsgenieTeamMetrics(metrics)
	c.collectOpsgenieUserMetrics(metrics)
}

func (c *opsgenieCollector) collectOpsgenieAlertMetrics(metrics chan<- prometheus.Metric) {
	c.processOpsgenieAlertsCreatedTotal(metrics)
	c.processOpsgenieAlertMetrics(metrics, alertOpenStatus)
	c.processOpsgenieAlertMetrics(metrics, alertClosedStatus)
}

func (c *opsgenieCollector) processOpsgenieAlertsCreatedTotal(metrics chan<- prometheus.Metric) {
	value, err := c.client.CountAlerts()
	if err != nil {
		level.Error(c.logger).Log("msg", "failed to query alerts from opsgenie", "err", err)
		return
	}

	metrics <- prometheus.MustNewConstMetric(
		c.OpsgenieAlertMetricsCreatedTotal,
		prometheus.CounterValue,
		value,
		[]string{"", ""}...,
	)

	teams, err := c.client.ListTeams()
	if err != nil {
		level.Error(c.logger).Log("msg", "failed to query teams from opsgenie", "err", err)
		return
	}

	for _, team := range teams {
		for _, priority := range priorities {
			value, err := c.client.CountAlertsBy(countAlertsParams{
				Team:     team.Name,
				Priority: priority,
			})
			if err != nil {
				level.Error(c.logger).Log("msg", "failed to query alerts from opsgenie", "err", err)
				return
			}

			metrics <- prometheus.MustNewConstMetric(
				c.OpsgenieAlertMetricsCreatedTotal,
				prometheus.CounterValue,
				value,
				[]string{team.Name, priority}...,
			)
		}
	}
}

func (c *opsgenieCollector) processOpsgenieAlertMetrics(metrics chan<- prometheus.Metric, status string) {
	value, err := c.client.CountAlertsBy(countAlertsParams{Status: status})
	if err != nil {
		level.Error(c.logger).Log("msg", "failed to query alerts from opsgenie", "err", err)
		return
	}

	metrics <- prometheus.MustNewConstMetric(
		c.OpsgenieAlertMetricsCount,
		prometheus.GaugeValue,
		value,
		[]string{status, "", ""}...,
	)

	teams, err := c.client.ListTeams()
	if err != nil {
		level.Error(c.logger).Log("msg", "failed to query teams from opsgenie", "err", err)
		return
	}

	for _, team := range teams {
		for _, priority := range priorities {
			value, err := c.client.CountAlertsBy(countAlertsParams{
				Status:   status,
				Team:     team.Name,
				Priority: priority,
			})
			if err != nil {
				level.Error(c.logger).Log("msg", "failed to query alerts from opsgenie", "err", err)
				return
			}

			metrics <- prometheus.MustNewConstMetric(
				c.OpsgenieAlertMetricsCount,
				prometheus.GaugeValue,
				value,
				[]string{status, team.Name, priority}...,
			)
		}
	}
}

func (c *opsgenieCollector) collectOpsgenieTeamMetrics(metrics chan<- prometheus.Metric) {
	c.processOpsgenieTeamsCount(metrics)
}

func (c *opsgenieCollector) processOpsgenieTeamsCount(metrics chan<- prometheus.Metric) {
	teams, err := c.client.ListTeams()
	if err != nil {
		level.Error(c.logger).Log("msg", "failed to query teams from opsgenie", "err", err)
		return
	}
	metrics <- prometheus.MustNewConstMetric(
		c.OpsgenieTeamMetricsCount,
		prometheus.GaugeValue,
		float64(len(teams)),
	)
}

func (c *opsgenieCollector) collectOpsgenieUserMetrics(metrics chan<- prometheus.Metric) {
	c.processOpsgenieUsersCount(metrics)
}

func (c *opsgenieCollector) processOpsgenieUsersCount(metrics chan<- prometheus.Metric) {
	valuesMap, err := c.client.CountUsersByRole()
	if err != nil {
		level.Error(c.logger).Log("msg", "failed to query users from opsgenie", "err", err)
		return
	}

	for k, v := range valuesMap {
		labels := []string{
			k,
		}
		metrics <- prometheus.MustNewConstMetric(
			c.OpsgenieUserMetricsCount,
			prometheus.GaugeValue,
			v,
			labels...,
		)
	}
}
