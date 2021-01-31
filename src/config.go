package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"

	"github.com/prometheus/common/log"
	"gopkg.in/yaml.v2"
)

type PayloadType struct {
	Type string            `yaml:"type"`
	Map  map[string]string `yaml:"map"`
}

type MQTTConfig struct {
	Broker    string `yaml:"broker"`
	Qos       int    `yaml:"qos"`
	Subscribe string `yaml:"subscribe"`
	ClientID  string `yaml:"clientid"`
	Username  string `yaml:"username"`
	Password  string `yaml:"password"`
}

type WebConfig struct {
	Listen     string `yaml:"listen"`
	MetricPath string `yaml:"metricpath"`
}

type Config struct {
	MQTT    MQTTConfig `yaml:"mqtt"`
	Web     WebConfig  `yaml:"web"`
	Metrics map[string]struct {
		Type string `yaml:"type"`
		Help string `yaml:"help"`
	} `yaml:"metrics"`
	Templates map[string]PayloadType `yaml:"templates"`
	Jobs      []struct {
		Job          string `yaml:"job"`
		Topic_prefix string `yaml:"topic_prefix"`
		Instances    []struct {
			Name    string `yaml:"name"`
			Sensors []struct {
				Id       string      `yaml:"id"`
				Topic    string      `yaml:"topic"`
				Type     PayloadType `yaml:"type"`
				Template string      `yaml:"template"`
			} `yaml:"sensors"`
		} `yaml:"instances"`
	} `yaml:"jobs"`
}

func homeDir() string {
	usr, err := user.Current()
	if err != nil {
		log.Fatal("Couldn't get users home directory")
		os.Exit(1)
	}
	return usr.HomeDir
}

func loadConfig() {
	var filename = ""
	if *cfgFile == "~/.mqtt2prom.yml" {
		filename = homeDir() + "/.mqtt2prom.yml"
	} else {
		filename = *cfgFile
	}
	log.Debugln("Loading mappings file: " + filename)
	yamlfile, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal(yamlfile, &config)
	if err != nil {
		panic(err)
	}
	if config.MQTT.Broker == "" {
		config.MQTT.Broker = "tcp://127.0.0.1:1883"
	}
	if config.MQTT.ClientID == "" {
		config.MQTT.ClientID = "mqtt2prom"
	}
	if config.MQTT.Subscribe == "" {
		config.MQTT.Subscribe = "#/"
	}

	if config.Web.Listen == "" {
		config.Web.Listen = ":9337"
	}
	if config.Web.MetricPath == "" {
		config.Web.MetricPath = "/metrics"
	}

	log.Debugln("Loaded mappings: " + fmt.Sprintf("%+v\n)", config))

}
