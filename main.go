package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const version = "v0.1.0"

var (
	statFile = flag.String("f", "/home/reforger/profile/profile/ServerAdminTools_Stats.json",
		"file to read stats from")
	listen = flag.String("listen", ":9840", "ip:port to listen on")
	lvs    = flag.String("l", "",
		"Labels/values to augment metrics with, in the form label1=val1,label2=val2")
	namespace = flag.String("namespace", "reforger_sat_exporter",
		"Namespace (prefix) to use for Prometheus metrics")
	timeout = flag.Duration("timeout", time.Second*3,
		"Timeout for webserver reading client request")
	once = flag.Bool("once", false, "Only output the stats to stdout and exit (for testing)")
	vers = flag.Bool("version", false, "Print version number and exit")
)

func main() {
	flag.Parse()
	if *vers {
		fmt.Fprintf(os.Stderr, "sat_exporter %s", version)
		os.Exit(0)
	}
	var labels, values []string
	var err error
	if *lvs != "" {
		labels, values, err = parseLabelsValues(*lvs)
		if err != nil {
			log.Printf("Could not parse labelvalue commandline flag: %s", err)
			os.Exit(-1)
		}
	}
	if *once {
		stats, err := readStats(*statFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not read stats file '%s': %s\n", *statFile, err)
			os.Exit(-1)
		}
		fmt.Printf("%s\n", stats)
		os.Exit(0)
	}
	log.Printf("Starting webserver on %s", *listen)
	http.Handle("/metrics", promhttp.Handler())
	srv := &http.Server{
		Addr:              *listen,
		ReadHeaderTimeout: *timeout,
	}

	reg := prometheus.NewPedanticRegistry()
	jt := newJSONCollector(*namespace, labels, values)
	prometheus.MustRegister(jt, reg)
	panic(srv.ListenAndServe())
}

func parseLabelsValues(ls string) ([]string, []string, error) {
	lstokens := strings.Split(ls, ",")
	labels := make([]string, 0, len(lstokens))
	values := make([]string, 0, len(lstokens))
	for _, slabelvalue := range lstokens {
		tokens := strings.Split(slabelvalue, "=")
		if len(tokens) != 2 {
			return labels, values,
				fmt.Errorf("label/values arg contains malformed token '%s'", slabelvalue)
		}
		labels = append(labels, tokens[0])
		values = append(values, tokens[1])
	}
	return labels, values, nil
}

func newJSONCollector(namespace string, labels, values []string) prometheus.Collector {
	eventlabels := labels
	eventlabels = append(eventlabels, "type")
	c := jsonCollector{
		namespace: namespace,
		RegSystems: prometheus.NewDesc(
			namespace+"_registered_systems", "Total number of registered systems", labels, nil),
		RegEntities: prometheus.NewDesc(
			namespace+"_registered_entities", "Total number of registered entities", labels, nil),
		RegGroups: prometheus.NewDesc(
			namespace+"_registered_groups", "Total number of registered groups", labels, nil),
		Uptime: prometheus.NewDesc(
			namespace+"_uptime_seconds", "Server uptime in seconds", labels, nil),
		AIChars: prometheus.NewDesc(
			namespace+"_ai_characters", "Total number AI characters", labels, nil),
		RegTasks: prometheus.NewDesc(
			namespace+"_registered_tasks", "Total number registered tasks", labels, nil),
		RegVics: prometheus.NewDesc(
			namespace+"_registered_vehicles", "Total number of registered vehicles", labels, nil),
		FPS: prometheus.NewDesc(
			namespace+"_frames_per_second", "Frames per second server-side", labels, nil),
		Players: prometheus.NewDesc(
			namespace+"_player_count", "Current number of players", labels, nil),
		Updated: prometheus.NewDesc(
			namespace+"_updated", "Last update", labels, nil),
		Events: prometheus.NewDesc(
			namespace+"_events_total", "Last update", eventlabels, nil),
		labels: labels,
		values: values,
	}
	return &c
}

type jsonCollector struct {
	namespace   string
	RegSystems  *prometheus.Desc
	RegEntities *prometheus.Desc
	RegGroups   *prometheus.Desc
	Uptime      *prometheus.Desc
	AIChars     *prometheus.Desc
	RegTasks    *prometheus.Desc
	RegVics     *prometheus.Desc
	FPS         *prometheus.Desc
	Players     *prometheus.Desc
	Updated     *prometheus.Desc
	Events      *prometheus.Desc
	labels      []string
	values      []string
}

func (c *jsonCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(c, ch)
}

// Collect returns the current state of all metrics of the collector.
func (c *jsonCollector) Collect(ch chan<- prometheus.Metric) {
	stats, err := readStats(*statFile)
	if err != nil {
		log.Printf("Could not read stats file %s: %s", *statFile, err)
		close(ch)
		return
	}
	ch <- prometheus.MustNewConstMetric(
		c.RegSystems, prometheus.GaugeValue, float64(stats.RegisteredSystems), c.values...)
	ch <- prometheus.MustNewConstMetric(
		c.RegEntities, prometheus.GaugeValue, float64(stats.RegisteredEntities), c.values...)
	ch <- prometheus.MustNewConstMetric(
		c.RegGroups, prometheus.GaugeValue, float64(stats.RegisteredGroups), c.values...)
	ch <- prometheus.MustNewConstMetric(
		c.Uptime, prometheus.CounterValue, float64(stats.UptimeSeconds), c.values...)
	ch <- prometheus.MustNewConstMetric(
		c.AIChars, prometheus.GaugeValue, float64(stats.AiCharacters), c.values...)
	ch <- prometheus.MustNewConstMetric(
		c.RegTasks, prometheus.GaugeValue, float64(stats.RegisteredTasks), c.values...)
	ch <- prometheus.MustNewConstMetric(
		c.FPS, prometheus.GaugeValue, float64(stats.Fps), c.values...)
	ch <- prometheus.MustNewConstMetric(
		c.Players, prometheus.GaugeValue, float64(stats.Players), c.values...)
	ch <- prometheus.MustNewConstMetric(
		c.Updated, prometheus.CounterValue, float64(stats.Updated), c.values...)
	for en, val := range stats.Events {
		lv := c.values
		lv = append(lv, en)
		ch <- prometheus.MustNewConstMetric(
			c.Events, prometheus.CounterValue, val, lv...)
	}
}

type satStatsRaw struct {
	RegisteredSystems  int                    `json:"registered_systems"`
	RegisteredEntities int                    `json:"registered_entities"`
	RegisteredGroups   int                    `json:"registered_groups"`
	UptimeSeconds      int                    `json:"uptime_seconds"`
	Fps                int                    `json:"fps"`
	RegisteredTasks    int                    `json:"registered_tasks"`
	RegisteredVehicles int                    `json:"registered_vehicles"`
	AiCharacters       int                    `json:"ai_characters"`
	Players            int                    `json:"players"`
	Updated            int                    `json:"updated"`
	Events             map[string]interface{} `mapstructure:",remain"`
}

func (r satStatsRaw) cook() satStats {
	return satStats{
		RegisteredSystems:  r.RegisteredSystems,
		RegisteredEntities: r.RegisteredEntities,
		RegisteredGroups:   r.RegisteredGroups,
		UptimeSeconds:      r.UptimeSeconds,
		Fps:                r.Fps,
		RegisteredTasks:    r.RegisteredTasks,
		RegisteredVehicles: r.RegisteredVehicles,
		AiCharacters:       r.AiCharacters,
		Players:            r.Players,
		Updated:            r.Updated,
		Events:             make(map[string]float64),
	}
}

type satStats struct {
	RegisteredSystems  int
	RegisteredEntities int
	RegisteredGroups   int
	UptimeSeconds      int
	Fps                int
	RegisteredTasks    int
	RegisteredVehicles int
	AiCharacters       int
	Players            int
	Updated            int
	Events             map[string]float64
}

type Events struct {
	ServeradmintoolsGameStarted int `json:"serveradmintools_game_started"`
}

func (s satStats) String() string {
	return fmt.Sprintf("RegisteredSystems: %d\nRegisteredEntities: %d\nRegisteredGroups: %d\nUptimeSeconds: %d\nAiCharacters: %d\nRegisteredTasks: %d\nRegisteredVehicles: %d\nFps: %d\nPlayers: %d\nUpdated: %d\nEvents: %#v\n", s.RegisteredSystems, s.RegisteredEntities, s.RegisteredGroups, s.UptimeSeconds, s.AiCharacters, s.RegisteredTasks, s.RegisteredVehicles, s.Fps, s.Players, s.Updated, s.Events)
}

func (e Events) String() string {
	return fmt.Sprintf("Events:\n    ServeradmintoolsGameStarted: %d\n", e.ServeradmintoolsGameStarted)
}

func readStats(fn string) (*satStats, error) {
	sr := &satStatsRaw{}
	data, err := os.ReadFile(fn)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, sr)
	if err != nil {
		return nil, err
	}
	s := sr.cook()
	for name, stat := range sr.Events {
		st, ok := stat.(float64)
		if !ok {
			return nil, fmt.Errorf("%s stat is not a float64: %T %#v", name, stat, stat)
		}
		s.Events[strings.TrimPrefix(name, "serveradmintools_")] = st
	}
	return &s, nil
}
