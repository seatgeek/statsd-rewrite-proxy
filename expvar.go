package main

import (
	"expvar"
	"fmt"
	"net/http"

	yaml "gopkg.in/yaml.v2"
)

var (
	ruleHitsSuccess = expvar.NewInt("rule_hits_success")
	ruleHitsMiss    = expvar.NewInt("rule_hits_miss")
)

func startHTTPServer() {
	logger.Infof("Starting HTTP server @ :%s", listenPortHTTP)
	http.HandleFunc("/datadog/expvar", showExprVar)
	http.ListenAndServe(":"+listenPortHTTP, nil)
}

func showExprVar(w http.ResponseWriter, r *http.Request) {
	metrics := make([]map[string]string, 0)
	metrics = append(metrics, map[string]string{"path": "rule_hits_success"})
	metrics = append(metrics, map[string]string{"path": "rule_hits_miss"})

	config := struct {
		ExpvarURL string              `yaml:"expvar_url"`
		Tags      []string            `yaml:"tags"`
		Metrics   []map[string]string `yaml:"metrics"`
	}{
		"http://127.0.0.1:" + listenPortHTTP + "/debug/vars",
		[]string{"project:statsd-rewrite-proxy"},
		metrics,
	}

	resp, err := yaml.Marshal(&config)
	if err != nil {
		message := fmt.Sprintf("[showExprVar] Could not marshal YAML: %s", err)
		logger.Errorf(message)
		http.Error(w, message, 500)
		return
	}

	text := string(resp)
	text = "---\n" + text

	resp = []byte(text)

	w.Header().Add("Content-Type", "text/yaml")
	w.Write(resp)
}
