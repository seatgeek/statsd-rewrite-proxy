package main

import (
	"bytes"
	"net"
	"os"
	"strconv"
	"strings"

	"time"

	"fmt"

	datadog "github.com/DataDog/datadog-go/statsd"
	statsd "github.com/cactus/go-statsd-client/statsd"
	"github.com/sirupsen/logrus"
)

const (
	workerCount      = 8
	metricTypeCount  = "count"
	metricTypeGauge  = "gauge"
	metricTypeTiming = "timing"

	// ip packet size is stored in two bytes and that is how big in theory the packet can be.
	// In practice it is highly unlikely but still possible to get packets bigger than usual MTU of 1500.
	packetSizeUDP = 0xffff
)

// AppConfig ...
type AppConfig struct {
	Host string
	Port int
}

// StatsDMetric ...
type StatsDMetric struct {
	name       string
	metricType string
	value      float64
	raw        string
}

var (
	logger         = logrus.New()
	listenPortHTTP = getHTTPListenPort()
	workerChannel  = make(chan []byte)
	quitChannel    = make(chan string)
	rules          = make([]*CaputeRule, 0)
	noTags         = make([]string, 0) // pre-computed empty tags for fallthrough metrics
)

func main() {
	dataDogClient, err := datadog.New("127.0.0.1:8125")
	if err != nil {
		logger.Fatal(err)
	}

	// parsePacketString("fabio.--frontend-node-renderer.frontend-node-renderer_service_bownty./.62_210_91_111_57029.std-dev:0")
	// return

	cfg := AppConfig{"0.0.0.0", 8126}

	createRules()

	go startHTTPServer()
	go listenUDP(cfg)

	// go emitter()

	// Start workers
	for x := 0; x < workerCount; x++ {
		go work(dataDogClient, x)
	}

	<-quitChannel
}

func createRules() {
	// fabio rewrites
	rules = append(rules, NewRule("fabio.http.status.{code}", "b.fabio.http.response_code"))
	rules = append(rules, NewRule("fabio.{fabio_service}.{fabio_host}.{fabio_path}.{fabio_upstream}.{fabio_dimension}", "b.fabio.service.requests.{fabio_dimension}"))
	rules = append(rules, NewRule("fabio.{fabio_service}.{fabio_host}.{fabio_path}", "b.fabio.service.requests"))

	// nomad rewrites
	rules = append(rules, NewRule("nomad.client.uptime.*", "nomad.client.uptime"))

	rules = append(rules, NewRule("nomad.client.host.memory.{client_id}.{nomad_metric}", "nomad.client.memmory.{nomad_metric}"))
	rules = append(rules, NewRule("nomad.client.host.cpu.{client_id}.{nomad_cpu_core}.{nomad_metric}", "nomad.client.cpu.{nomad_metric}"))
	rules = append(rules, NewRule("nomad.client.host.disk.{client_id}.{nomad_device}.{nomad_metric}", "nomad.client.disk.{nomad_metric}"))

	rules = append(rules, NewRule("nomad.client.allocs.{nomad_job}.{nomad_task_group}.{nomad_allocation_id}.{nomad_task}.memory.{nomad_metric}", "nomad.allocation.memory.{nomad_metric}"))
	rules = append(rules, NewRule("nomad.client.allocs.{nomad_job}.{nomad_task_group}.{nomad_allocation_id}.{nomad_task}.cpu.{nomad_metric}", "nomad.allocation.cpu.{nomad_metric}"))
}

func listenUDP(cfg AppConfig) {
	logger.Infof("Starting StatsD UDP listener on %s and port %d", cfg.Host, cfg.Port)
	listener, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP(cfg.Host), Port: cfg.Port})
	if err != nil {
		logger.Fatalf("Error setting up UDP listener: %s (exiting...)", err)
	}
	defer listener.Close()

	for {
		buf := make([]byte, packetSizeUDP)
		num, _, err := listener.ReadFromUDP(buf)

		if err != nil {
			logger.Infof("Error reading from UDP buffer: %s (skipping...)", err)
			continue
		}

		workerChannel <- buf[:num]
	}
}

func work(dataDogClient *datadog.Client, workerID int) {
	logger.Infof("[%d] Starting worker", workerID)

	for {
		select {
		case <-quitChannel:
			return
		case data := <-workerChannel:
			// loop the metric lines
			for {
				idx := bytes.IndexByte(data, '\n')
				var line []byte

				// protocol does not require line to end in \n
				if idx == -1 { // \n not found
					if len(data) == 0 {
						break
					}

					line = data
					data = nil
				} else { // usual case
					line = data[:idx]
					data = data[idx+1:]
				}

				str := string(line)

				// parse into a statsd metrict struct
				metric, err := parsePacketString(str)
				if err != nil {
					logger.Errorf("Invalid package '%s': %s", str, err)
					continue
				}

				found := false

				// loop our rewrite rules until we find a match
				for _, rule := range rules {
					// try to match the metric to our rules
					result := rule.FindStringSubmatchMap(metric.name)

					// if no captures, move on
					if len(result.Captures) == 0 {
						continue
					}

					ruleHitsSuccess.Add(1)
					logger.Debugf("[%d] Found match for '%s', emitting as '%s' (value: %+v | >> %+v <<)", workerID, metric.name, result.name, metric.value, metric.raw)

					switch metric.metricType {
					case metricTypeCount:
						dataDogClient.Count(result.name, int64(metric.value), result.Tags, 1)
					case metricTypeTiming:
						dataDogClient.Timing(result.name, time.Duration(metric.value), result.Tags, 1)
					case metricTypeGauge:
						dataDogClient.Gauge(result.name, metric.value, result.Tags, 1)
					}

					found = true
					break
				}

				if found {
					continue
				}

				ruleHitsMiss.Add(1)

				logger.Warnf("[%d] No match found for '%s', relaying unmodified (value: %+v | >> %+v <<)", workerID, metric.name, metric.value, metric.raw)

				switch metric.metricType {
				case metricTypeCount:
					dataDogClient.Count(metric.name, int64(metric.value), noTags, 1)
				case metricTypeTiming:
					dataDogClient.Timing(metric.name, time.Duration(metric.value), noTags, 1)
				case metricTypeGauge:
					dataDogClient.Gauge(metric.name, metric.value, noTags, 1)
				}
			}
		}
	}
}

// dummy emitter to test functionality while testing
func emitter() {
	client, err := statsd.NewClient("127.0.0.1:8126", "")
	if err != nil {
		logger.Fatalf(err.Error())
	}
	defer client.Close()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-quitChannel:
			return
		case <-ticker.C:
			// Send a stat
			client.Inc("fabio.admin_nginx.admin_bownty_net..62_210_94_249_23928", 1, 1)
			client.Inc("fabio.--audits-nginx.audits_bownty_net./.62_210_248_90_50825.50-percentile", 43, 1.0)
			client.Inc("fabio.http.status.200", 1, 1)
			client.Inc("nomad.client.host.memory.c668a4ca-a835-de6a-4aa2-564d802c28e3.available", 1, 1)
		}
	}
}

// parse a statsd line into a metric struct
func parsePacketString(data string) (*StatsDMetric, error) {
	logger.Debugf("Parse metric: %s", data)

	ret := new(StatsDMetric)

	first := strings.Split(data, ":")
	if len(first) < 2 {
		logger.Errorf("Malformatted metric: %s", data)
		return ret, fmt.Errorf("Malformatted metric: %s", data)
	}

	name := first[0]

	second := strings.Split(first[1], "|")
	if len(second) < 2 {
		return ret, fmt.Errorf("Malformatted metric: %s", data)
	}

	value64, err := strconv.ParseInt(second[0], 10, 0)
	if err != nil {
		return ret, fmt.Errorf("Could not parse value to int64: %+v", second[0])
	}
	value := float64(value64)

	// check for a samplerate
	third := strings.Split(second[1], "@")
	metricType := third[0]

	ret.name = name
	ret.value = value
	ret.raw = data

	switch metricType {
	case "c":
		ret.metricType = metricTypeCount
	case "ms":
		ret.metricType = metricTypeTiming
	case "gf":
		ret.metricType = metricTypeGauge
	case "g":
		ret.metricType = metricTypeGauge
	default:
		logger.Errorf("Unknown metrics type: %s", metricType)
		return ret, fmt.Errorf("Unknown metrics type: %s", metricType)
	}

	// remove stupid double dashes from metrics
	ret.name = strings.Replace(ret.name, "--", "", -1)

	return ret, nil
}

func getHTTPListenPort() string {
	port := os.Getenv("NOMAD_PORT_http")
	if port == "" {
		port = "4200"
	}

	return port
}
