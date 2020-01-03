package main

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

var (
	// mqtt2prom related counters
	mqttTotalMessages = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "mqtt2prom_messages_recieved_total",
		Help: "The total number of messages recieved on message bus",
	})
	mqttRegistredMessages = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "mqtt2prom_messages_recieved_registred",
		Help: "The number of messages recieved on message bus and is accociated with atleast one metric",
	})
	// map to hold the type of metric, metricdata[topic]
	sensordata   map[string]SensorMetadata
	gaugeMetrics map[string]*prometheus.GaugeVec
)

// Holds information on each metric collected
type SensorMetadata struct {
	Type   PayloadType
	Labels prometheus.Labels
}

func getJSONValue(payload []byte, pattern string) (float64, bool) {
	var message interface{}
	var ok bool

	err := json.Unmarshal(payload, &message)
	if err != nil {
		log.Errorf("Coulnd't parse the payload as json: %v", payload)
		return 0, false
	}
	log.Debugf("Parsing json. pattern: %s  payload %v", pattern, message)

	pat := strings.Split(pattern, ".")
	data := message.(map[string]interface{})
	for i := 0; i < len(pat)-1; i++ {
		data, ok = data[pat[i]].(map[string]interface{})
		if !ok {
			log.Debugf("Couldn't find  %s' on jason: %s", pattern, string(payload))
			return 0, false
		}
	}
	value, ok := data[pat[len(pat)-1]].(float64)
	if !ok {
		log.Errorf("json value is not float64 type. %s %v", pattern, payload)
		return 0, false
	}
	return value, true

}

func updateMetric(topic string, payload []byte) {
	_, exists := sensordata[topic]
	log.Debugf("Using topic: %s", topic)
	if !exists {
		log.Infof("No metric accociated with topic: %s", topic)
		return
	}

	if sensordata[topic].Type.Type == "json" {
		log.Debugln("Parsing json metrics")
		for pattern, metricname := range sensordata[topic].Type.Map {
			value, ok := getJSONValue(payload, pattern)
			if !ok {
				break
			}
			if config.Metrics[metricname].Type == "gauge" {
				gaugeMetrics[metricname].With(sensordata[topic].Labels).Set(value)
			} else {
				log.Errorf("Unkonwn metric type '%s' defined for %s", config.Metrics[metricname].Type, metricname)
			}
		}
	} else if sensordata[topic].Type.Type == "value" {
		log.Debugln("Parsing value metric")
		value, error := strconv.ParseFloat(string(payload), 64)
		if error == nil {
			gaugeMetrics[sensordata[topic].Type.Map["metric"]].With(sensordata[topic].Labels).Set(value)
		}
	}
	mqttRegistredMessages.Inc()
}

func initCollector() {
	var (
		//		topic      string
		//		labels     prometheus.Labels
		labelNames = []string{"job", "instance", "sensorid"}
	)
	prometheus.MustRegister(mqttTotalMessages)
	prometheus.MustRegister(mqttRegistredMessages)
	sensordata = make(map[string]SensorMetadata)
	gaugeMetrics = make(map[string]*prometheus.GaugeVec)

	for metricname, metric := range config.Metrics {
		if metric.Type == "gauge" {
			gaugeMetrics[metricname] = prometheus.NewGaugeVec(
				prometheus.GaugeOpts{
					Name: metricname,
					Help: metric.Help,
				},
				labelNames,
			)
			prometheus.MustRegister(gaugeMetrics[metricname])
			log.Debugf("Loaded gauge metric: %s", metricname)
		} else {
			log.Warnf("Unsupported metric type %s configured on %s", metric.Type, metricname)
		}
	}

	for _, job := range config.Jobs {
		for _, instance := range job.Instances {
			for _, sensor := range instance.Sensors {
				topic := ""
				if job.Topic_prefix != "" {
					topic = job.Topic_prefix
				}
				if sensor.Topic != "" {
					if topic != "" {
						topic = topic + "/"
					}
					topic = topic + sensor.Topic
				}
				if sensor.Template != "" {
					if sensor.Type.Type != "" {
						log.Warnf("%s has both Template and Type define, template used", sensor.Id)
					}
					st, exists := config.Templates[sensor.Template]
					if exists {
						log.Debugf("Added sensor for topic: %s", topic)
						sensordata[topic] = SensorMetadata{Type: st, Labels: prometheus.Labels{"job": job.Job, "instance": instance.Name, "sensorid": sensor.Id}}
					} else {
						log.Errorf("Sensor id %s is using a non existing template: %s", sensor.Id, sensor.Template)
					}
				} else if sensor.Type.Type != "" {
					log.Debugf("Added sensor for topic: %s", topic)
					sensordata[topic] = SensorMetadata{Type: sensor.Type, Labels: prometheus.Labels{"job": job.Job, "instance": instance.Name, "sensorid": sensor.Id}}
				} else {
					log.Errorf("%s sensor doesn't have either template or Type defined, not loaded", sensor.Id)
				}
			}
		}
	}
	log.Debugf("Loaded sensor metadata: %v", sensordata)
}
