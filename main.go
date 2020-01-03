package main

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	cfgFile = kingpin.Flag("config", "Path to mappings file").Default("~/.mqtt2prom.yml").String()
	config  Config
)

func main() {
	log.AddFlags(kingpin.CommandLine)
	kingpin.Parse()

	loadConfig()
	initCollector()
	startMQTT()
	log.Infof("Listening on %s%s", config.Web.Listen, config.Web.MetricPath)
	http.Handle(config.Web.MetricPath, promhttp.Handler())
	http.ListenAndServe(config.Web.Listen, nil)
}
