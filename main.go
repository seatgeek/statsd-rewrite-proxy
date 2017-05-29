package main

import (
	"errors"
	"net"
	"os"
	"runtime"
	"strconv"
	"strings"

	"time"

	"fmt"

	datadog "github.com/DataDog/datadog-go/statsd"
	"github.com/sirupsen/logrus"
)

const (
	// UDP packet limit, see
	// https://en.wikipedia.org/wiki/User_Datagram_Protocol#Packet_structure
	UDP_MAX_PACKET_SIZE int = 64 * 1024
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
	floatvalue float64
	intvalue   int64
	strvalue   string
	samplerate float64
}

var (
	logger         = logrus.New()
	listenPortHTTP = getHTTPListenPort()
	workerChannel  = make(chan []byte, 10000)
	quitChannel    = make(chan string)
	rules          = &Rules{}
	noTags         = make([]string, 0) // pre-computed empty tags for fallthrough metrics

	counterProcessed int64
	counterRewritten int64
	counterRelayed   int64
	counterOverflow  int64
	counterDropped   int64
	countersPassed   int64
	countersMissed   int64
)

func main() {
	dataDogClient, err := datadog.NewBuffered("127.0.0.1:8125", 10)
	if err != nil {
		logger.Fatal(err)
	}

	cfg := AppConfig{"0.0.0.0", 8126}

	createRules()

	go startHTTPServer()
	go listenUDP(cfg)
	go printStats()

	workerCount := runtime.NumCPU()
	for x := 0; x < workerCount; x++ {
		go work(dataDogClient, x)
	}

	<-quitChannel
}

func printStats() {
	ticker := time.NewTicker(1 * time.Minute)
	for {
		logger.Infof("Processed %s | Rewritten: %s | Relayed: %s | Dropped: %s | Passed: %s | Skipped: %s | Overflow: %s, Queued: %s",
			formatNumber(counterProcessed),
			formatNumber(counterRewritten),
			formatNumber(counterRelayed),
			formatNumber(counterDropped),
			formatNumber(countersPassed),
			formatNumber(countersMissed),
			formatNumber(counterOverflow),
			formatNumber(int64(len(workerChannel))),
		)
		<-ticker.C
	}
}

func formatNumber(n int64) string {
	in := strconv.FormatInt(n, 10)
	out := make([]byte, len(in)+(len(in)-2+int(in[0]/'0'))/3)
	if in[0] == '-' {
		in, out[0] = in[1:], '-'
	}

	for i, j, k := len(in)-1, len(out)-1, 0; ; i, j = i-1, j-1 {
		out[j] = in[i]
		if i == 0 {
			return string(out)
		}
		if k++; k == 3 {
			j, k = j-1, 0
			out[j] = ','
		}
	}
}

func listenUDP(cfg AppConfig) {
	logger.Infof("Starting StatsD UDP listener on %s and port %d", cfg.Host, cfg.Port)
	listener, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP(cfg.Host), Port: cfg.Port})
	if err != nil {
		logger.Fatalf("Error setting up UDP listener: %s (exiting...)", err)
	}

	defer listener.Close()

	buf := make([]byte, UDP_MAX_PACKET_SIZE)

	for {
		n, _, err := listener.ReadFromUDP(buf)
		if err != nil && !strings.Contains(err.Error(), "closed network") {
			logger.Errorf("Error READ: %s\n", err.Error())
			continue
		}

		bufCopy := make([]byte, n)
		copy(bufCopy, buf[:n])

		select {
		case workerChannel <- bufCopy:
		default:
			counterOverflow = counterOverflow + 1
			logger.Error("StatsD message queue is full, dropping message")
		}
	}
}

func work(dataDogClient *datadog.Client, workerID int) {
	logger.Infof("[%d] Starting worker", workerID)

	for {
		select {
		case <-quitChannel:
			return
		case data := <-workerChannel:
			lines := strings.Split(string(data), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)

				if line == "" {
					continue
				}

				metrics, err := parsePacketString(line)
				if err != nil {
					logger.Error(err)
					continue
				}

				for _, metric := range metrics {

					counterProcessed = counterProcessed + 1

					relay := false
					match := false

					// loop our rewrite rules until we find a match
					for _, rule := range rules.list {
						// try to match the metric to our rules
						result := rule.FindStringSubmatchMap(metric.name)

						// If the rule didn't match the metric, keep searching
						if result.action == ruleActionMiss {
							continue
						}

						// If the result action is anything but "miss", we actually matched something!
						match = true

						// If the rule did match the metric, and it should be ignore, break out
						if result.action == ruleActionDrop {
							relay = false
							counterDropped = counterDropped + 1
							break
						}

						// Relay the metric as-is
						if result.action == ruleActionRelay {
							relay = true
							counterRelayed = counterRelayed + 1
							break
						}

						// If we get to here, we assume it's a match
						if result.action != ruleActionMatch {
							panic(fmt.Sprintf("Unknown result action: %s", result.action))
						}

						// if no captures, keep searching
						if len(result.Captures) == 0 {
							logger.Warningf("Did match '%s' to '%s', but there was 0 capture groups", rule.name, rule.Regexp.String())
							continue
						}

						counterRewritten = counterRewritten + 1
						ruleHitsSuccess.Add(1)

						logger.Debugf("[%d] Found match for '%s', emitting as '%s'", workerID, metric.name, result.name)

						switch metric.metricType {
						case "c":
							dataDogClient.Count(result.name, metric.intvalue, result.Tags, metric.samplerate)
						case "ms":
							dataDogClient.Timing(result.name, time.Duration(metric.floatvalue), result.Tags, metric.samplerate)
						case "g":
							dataDogClient.Gauge(result.name, metric.floatvalue, result.Tags, metric.samplerate)
						case "s":
							dataDogClient.Set(result.name, metric.strvalue, result.Tags, metric.samplerate)
						case "h":
							dataDogClient.Histogram(result.name, metric.floatvalue, result.Tags, metric.samplerate)
						default:
							logger.Fatalf("Unknown metric type: %s", metric.metricType)
						}

						relay = false
						break
					}

					if !relay {
						continue
					}

					ruleHitsMiss.Add(1)

					if !match {
						logger.Warnf("[%d] No match found for '%s', relaying unmodified", workerID, metric.name)
						countersMissed = countersMissed + 1
					} else {
						logger.Debugf("[%d] relaying '%s' unmodified", workerID, metric.name)
					}

					switch metric.metricType {
					case "c":
						dataDogClient.Count(metric.name, metric.intvalue, noTags, metric.samplerate)
					case "ms":
						dataDogClient.Timing(metric.name, time.Duration(metric.floatvalue), noTags, metric.samplerate)
					case "g":
						dataDogClient.Gauge(metric.name, metric.floatvalue, noTags, metric.samplerate)
					case "s":
						dataDogClient.Set(metric.name, metric.strvalue, noTags, metric.samplerate)
					case "h":
						dataDogClient.Histogram(metric.name, metric.floatvalue, noTags, metric.samplerate)
					default:
						logger.Fatalf("Unknown metric type: %s", metric.metricType)
					}
				}
			}
		}
	}
}

func parsePacketString(line string) ([]*StatsDMetric, error) {
	res := make([]*StatsDMetric, 0)

	// Validate splitting the line on ":"
	bits := strings.Split(line, ":")
	if len(bits) < 2 {
		logger.Errorf("Error: splitting ':', Unable to parse metric: %s\n", line)
		return nil, errors.New("Error Parsing statsd line")
	}

	// Extract bucket name from individual metric bits
	bucketName, bits := bits[0], bits[1:]

	// Add a metric for each bit available
	for _, bit := range bits {
		m := StatsDMetric{}

		m.name = bucketName

		// Validate splitting the bit on "|"
		pipesplit := strings.Split(bit, "|")
		if len(pipesplit) < 2 {
			logger.Errorf("Splitting '|', Unable to parse metric: %s\n", line)
			return nil, errors.New("Error Parsing statsd line")
		} else if len(pipesplit) > 2 {
			sr := pipesplit[2]
			errmsg := "Parsing sample rate, %s, it must be in format like: @0.1, @0.5, etc. Ignoring sample rate for line: %s\n"

			if strings.Contains(sr, "@") && len(sr) > 1 {
				samplerate, err := strconv.ParseFloat(sr[1:], 64)

				if err != nil {
					logger.Errorf(errmsg, err.Error(), line)
				} else {
					// sample rate successfully parsed
					m.samplerate = samplerate
				}
			} else {
				logger.Errorf(errmsg, "", line)
			}
		}

		if m.samplerate == 0 {
			m.samplerate = 1
		}

		// Validate metric type
		switch pipesplit[1] {
		case "gf":
			m.metricType = "g"
		case "g", "c", "s", "ms", "h":
			m.metricType = pipesplit[1]
		default:
			logger.Printf("E! Error: Statsd Metric type %s unsupported", pipesplit[1])
			return nil, errors.New("Error Parsing statsd line")
		}

		// Parse the value
		if strings.HasPrefix(pipesplit[0], "-") || strings.HasPrefix(pipesplit[0], "+") {
			if m.metricType != "g" && m.metricType != "c" {
				logger.Printf("E! Error: +- values are only supported for gauges & counters: %s\n", line)
				return nil, errors.New("Error Parsing statsd line")
			}
		}

		switch m.metricType {
		case "g", "ms", "h":
			v, err := strconv.ParseFloat(pipesplit[0], 64)
			if err != nil {
				logger.Errorf("Error: parsing value to float64: %s\n", line)
				return nil, errors.New("Error Parsing statsd line")
			}
			m.floatvalue = v
		case "c":
			var v int64
			v, err := strconv.ParseInt(pipesplit[0], 10, 64)
			if err != nil {
				v2, err2 := strconv.ParseFloat(pipesplit[0], 64)
				if err2 != nil {
					logger.Errorf("Error: parsing value to int64: %s\n", line)
					return nil, errors.New("Error Parsing statsd line")
				}
				v = int64(v2)
			}

			// If a sample rate is given with a counter, divide value by the rate
			if m.samplerate != 0 && m.metricType == "c" {
				v = int64(float64(v) / m.samplerate)
			}

			m.intvalue = v
		case "s":
			m.strvalue = pipesplit[0]
		}

		res = append(res, &m)
	}

	return res, nil
}

func getHTTPListenPort() string {
	port := os.Getenv("NOMAD_PORT_http")
	if port == "" {
		port = "4200"
	}

	return port
}
