package main

import (
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/prometheus/common/log"
)

func onMessageReceived(client MQTT.Client, message MQTT.Message) {
	mqttTotalMessages.Inc()
	updateMetric(message.Topic(), message.Payload())
}

func startMQTT() {
	opts := MQTT.NewClientOptions()
	opts.AddBroker(config.MQTT.Broker)
	opts.SetClientID(config.MQTT.ClientID)
	if config.MQTT.Username != "" {
		opts.SetUsername(config.MQTT.Username)
		if config.MQTT.Password != "" {
			opts.SetPassword(config.MQTT.Password)
		}
	}
	opts.OnConnect = func(c MQTT.Client) {
		if token := c.Subscribe(config.MQTT.Subscribe, byte(config.MQTT.Qos), onMessageReceived); token.Wait() && token.Error() != nil {
			panic(token.Error())
		}
	}
	client := MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	} else {
		log.Infof("Connected to messge broker @ %s\n", config.MQTT.Broker)
	}

}
