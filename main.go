package main

import (
	"log"
	"net"
	"strconv"
	"strings"

	"time"

	datadog "github.com/DataDog/datadog-go/statsd"
	statsd "github.com/cactus/go-statsd-client/statsd"
)

const (
	workerCount  = 8
	metricCount  = "count"
	metricGauge  = "gauge"
	metricTiming = "timing"
)

type config struct {
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

var workerChannel = make(chan []byte)
var quitChannel = make(chan string)
var rules = make([]*CaputeRule, 0)
var noTags = make([]string, 0)

func main() {
	client, err := datadog.New("127.0.0.1:8125")
	if err != nil {
		log.Fatal(err)
	}

	cfg := config{"127.0.0.1", 1234}

	log.Printf("Starting StatsD listener on %s and port %d", cfg.Host, cfg.Port)
	listener, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP(cfg.Host), Port: cfg.Port})
	if err != nil {
		log.Printf("Error setting up listener: %s (exiting...)", err)
	}

	// go emitter()

	for x := 0; x < workerCount; x++ {
		go work(client)
	}

	for {
		buf := make([]byte, 512)
		num, _, err := listener.ReadFromUDP(buf)
		if err != nil {
			log.Printf("Error reading from UDP buffer: %s (skipping...)", err)
		} else {
			workerChannel <- buf[0:num]
		}
	}
}

func init() {
	rules = append(rules, NewRule("fabio.{service}.{host}.{upstream}", "fabio.service.requests"))
	rules = append(rules, NewRule("fabio.http.status.{code}", "fabio.http.status"))
}

func work(client *datadog.Client) {
	for {
		select {
		case data := <-workerChannel:
			// the buffer can contain multiple lines
			metrics := strings.Split(string(data), "\n")

			// loop the metric lines
			for _, str := range metrics {
				found := false

				// parse into a statsd metrict struct
				metric := parsePacketString(str)

				// loop our rewrite rules until we find a match
				for _, rule := range rules {
					// try to match the metric to our rules
					result := rule.FindStringSubmatchMap(metric.name)

					// if no captures, move on
					if len(result.Captures) == 0 {
						continue
					}

					found = true
					log.Printf("Found match for %s!", metric.name)
					log.Printf("%+v", result)
					break
				}

				if found {
					continue
				}

				log.Printf("No match found for '%s', relaying unmodified", metric.name)

				switch metric.metricType {
				case metricCount:
					client.Count(metric.name, int64(metric.value), noTags, 1)
				case metricTiming:
					client.Timing(metric.name, time.Duration(metric.value), noTags, 1)
				case metricGauge:
					client.Gauge(metric.name, metric.value, noTags, 1)
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
		log.Fatal(err)
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
func parsePacketString(data string) *StatsDMetric {
	ret := new(StatsDMetric)
	first := strings.Split(data, ":")
	if len(first) < 2 {
		log.Printf("Malformatted metric: %s", data)
		return ret
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
		ret.metricType = metricCount
		fallthrough
	case "ms":
		ret.metricType = metricTiming
		fallthrough
	case "g":
		ret.metricType = metricGauge
		ret.name = name
		ret.value = value
		ret.raw = data
	default:
		log.Printf("Unknown metrics type: %s", metricType)
	}

	return ret
}
