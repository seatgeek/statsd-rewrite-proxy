package main

import (
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

	cfg := AppConfig{"0.0.0.0", 8126}

	// go listenTCP(cfg)
	go startHTTPServer()
	go listenUDP(cfg)

	// go emitter()

	// Start workers
	for x := 0; x < workerCount; x++ {
		go work(dataDogClient, x)
	}

	<-quitChannel
}

func listenUDP(cfg AppConfig) {
	logger.Infof("Starting StatsD UDP listener on %s and port %d", cfg.Host, cfg.Port)
	listener, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP(cfg.Host), Port: cfg.Port})
	if err != nil {
		logger.Fatalf("Error setting up UDP listener: %s (exiting...)", err)
	}
	defer listener.Close()

	for {
		buf := make([]byte, 512)
		num, _, err := listener.ReadFromUDP(buf)
		if err != nil {
			logger.Infof("Error reading from UDP buffer: %s (skipping...)", err)
			continue
		}

		workerChannel <- buf[0:num]
	}
}

// func listenTCP(cfg AppConfig) {
// 	logger.Infof("Starting StatsD TCP listener on %s and port %d", cfg.Host, cfg.Port)
// 	listener, err := net.ListenTCP("tcp", &net.TCPAddr{IP: net.ParseIP(cfg.Host), Port: cfg.Port})
// 	if err != nil {
// 		logger.Fatalff("Error setting up TCP listener: %s (exiting...)", err)
// 	}
// 	defer listener.Close()

// 	conn, err := listener.AcceptTCP()
// 	if err != nil {
// 		panic(err)
// 	}

// 	for {
// 		buf := make([]byte, 512)
// 		num, _, err := listener.Re(buf)
// 		if err != nil {
// 			logger.Infof("Error reading from UDP buffer: %s (skipping...)", err)
// 			continue
// 		}

// 		workerChannel <- buf[0:num]
// 	}
// }

func init() {
	rules = append(rules, NewRule("fabio.{service}.{host}.{upstream}", "fabio.service.requests"))
	rules = append(rules, NewRule("fabio.http.status.{code}", "fabio.http.status"))
}

func work(dataDogClient *datadog.Client, workerID int) {
	logger.Infof("[%d] Starting worker", workerID)

	for {
		select {
		case data := <-workerChannel:
			// the buffer can contain multiple lines
			metrics := strings.Split(string(data), "\n")

			// loop the metric lines
			for _, str := range metrics {
				logger.Infof("Handling metric: %s", str)

				// parse into a statsd metrict struct
				metric, err := parsePacketString(str)
				if err != nil {
					logger.Errorf("Invalid package '%s': %s", str, err)
					continue
				}

				// loop our rewrite rules until we find a match
				for _, rule := range rules {
					// try to match the metric to our rules
					result := rule.FindStringSubmatchMap(metric.name)

					// if no captures, move on
					if len(result.Captures) == 0 {
						continue
					}

					ruleHitsSuccess.Add(1)
					logger.Infof("[%d] Found match for '%s', emitting as '%s'", workerID, metric.name, rule.name)

					switch metric.metricType {
					case metricTypeCount:
						dataDogClient.Count(rule.name, int64(metric.value), result.Tags, 1)
					case metricTypeTiming:
						dataDogClient.Timing(rule.name, time.Duration(metric.value), result.Tags, 1)
					case metricTypeGauge:
						dataDogClient.Gauge(rule.name, metric.value, result.Tags, 1)
					}

					break
				}

				ruleHitsMiss.Add(1)

				logger.Infof("[%d] No match found for '%s', relaying unmodified", workerID, metric.name)

				switch metric.metricType {
				case metricTypeCount:
					dataDogClient.Count(metric.name, int64(metric.value), noTags, 1)
				case metricTypeTiming:
					dataDogClient.Timing(metric.name, time.Duration(metric.value), noTags, 1)
				case metricTypeGauge:
					dataDogClient.Gauge(metric.name, metric.value, noTags, 1)
				}
			}
		case <-quitChannel:
			return
		}
	}
}

// dummy emitter to test functionality while testing
func emitter() {
	client, err := statsd.NewClient("127.0.0.1:1234", "")
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
			client.Inc("stat1", 42, 1.0)
			client.Inc("fabio.admin_nginx.admin_bownty_net..62_210_94_249_23928", 1, 1)
			client.Inc("fabio.http.status.200", 1, 1)
			client.Inc("fabio.catch_all...bownty_com_80", 1, 1)
			client.Gauge("demo.derp.gauge", 43, 1.0)
			client.TimingDuration("derp.timing", 5*time.Second, 1.0)
		}
	}
}

// parse a statsd line into a metric struct
func parsePacketString(data string) (*StatsDMetric, error) {
	ret := new(StatsDMetric)

	first := strings.Split(data, ":")
	if len(first) < 2 {
		logger.Errorf("Malformatted metric: %s", data)
		return ret, fmt.Errorf("Malformatted metric: %s", data)
	}

	name := first[0]
	second := strings.Split(first[1], "|")
	value64, _ := strconv.ParseInt(second[0], 10, 0)
	value := float64(value64)

	// check for a samplerate
	third := strings.Split(second[1], "@")
	metricType := third[0]

	switch metricType {
	case "c":
		ret.metricType = metricTypeCount
		fallthrough
	case "ms":
		ret.metricType = metricTypeTiming
		fallthrough
	case "g":
		ret.metricType = metricTypeGauge
		ret.name = name
		ret.value = value
		ret.raw = data
	default:
		logger.Errorf("Unknown metrics type: %s", metricType)
		return ret, fmt.Errorf("Unknown metrics type: %s", metricType)
	}

	return ret, nil
}

func getHTTPListenPort() string {
	port := os.Getenv("NOMAD_PORT_http")
	if port == "" {
		port = "4200"
	}

	return port
}
