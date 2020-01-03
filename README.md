# mqtt2prom
This is a small Prometheus gateway/proxy that exposes values published on MQTT so it can be consumed and tracked by Prometheus.
```
mqtt2prom --help
usage: mqtt2prom [<flags>]

Flags:
  --help                        Show context-sensitive help (also try --help-long and --help-man).
  --config="~/.mqtt2prom.yml"   Path to mappings file
  --log.level="info"            Only log messages with the given severity or above. Valid levels: [debug, info, warn, error, fatal]
  --log.format="logger:stderr"  Set the log target and format. Example: "logger:syslog?appname=bob&local=7" or "logger:stdout?json=true"
```
A part from the sensors/metrics you configure mqtt2prom has two additional internal counters that are exposed
```
# HELP mqtt2prom_messages_recieved_registred The number of messages recieved on message bus and is accociated with atleast one metric
# TYPE mqtt2prom_messages_recieved_registred counter
mqtt2prom_messages_recieved_registred 17
# HELP mqtt2prom_messages_recieved_total The total number of messages recieved on message bus
# TYPE mqtt2prom_messages_recieved_total counter
mqtt2prom_messages_recieved_total 20
```

Table of contents
-----------------

- [Installation](#Installation)
- [Configuration](#Configuration)
  - [mqtt](#mqtt)
  - [web](#web)
  - [metrics](#metrics)
  - [jobs](#jobs)
  - [sensor types](#sensor-types)
    - [value](#value)
    - [json](#json)
  - [templates](#templates)
- [Prometheus](#Prometheus)

Installation
------------
```shell
mkdir ~/go
export GOPATH=~/go
export PATH="$PATH:$GOPATH/bin"
go get github.com/SweBarre/mqtt2prom
```

Configuration
-------------
Configuration is done in [YAML](https://yaml.org/), the `mqtt2prom.yml` is an example on how it could/should look like. It consists of several sections 
### mqtt
The mqtt in the sample file shows the defaults
```yaml
mqtt:
  broker: tcp://127.0.0.1:1883
  qos: 0
  subscribe: "#"
  clientid: mqtt2prom
  username: 
  password:
```
### web
The web sections configures how the metrics is exposed to prometheus, default it listens on all IP-addresses on port 9337
```yaml
 web:
   listen: ":9337"
   metricpath: /metrics
```
### metrics
This is where you define what metrics you want to expose, currently I've only added the gauge type of metric (just because I haven't had the need for any other type yet)
```yaml
metrics:
  sensor_temperature_celsius:
    type: gauge
    help: Temperature in celsius
  sensor_humidity_percent:
    type: gauge
    help: Humidity in percent
  sensor_linkquality_rssi:
    type: gauge
    help: Quality of link to sensor
```
The `sensor_temperature_celsius`in the above example is the name of the metric, and it will show up like this
```
# HELP sensor_temperature_celsius Temperature in celsius
# TYPE sensor_temperature_celsius gauge
sensor_temperature_celsius{instance="computer_room",job="zigbee",sensorid="computer_room"} 19.8
sensor_temperature_celsius{instance="kitchen",job="rfxcom",sensorid="kitchen"} 21.8
sensor_temperature_celsius{instance="laundry",job="zigbee",sensorid="laundry"} 19.84
sensor_temperature_celsius{instance="outside",job="rfxcom",sensorid="backside"} 5.9
```
### jobs
jobs section consists of a list where you define all different jobs.
```yaml
jobs:
  - job: <the job name>
    topic_prefix: <prefix>
    instances:
      - name: <the instance name>
        sensors:
          - id: <sensor id>
            topic: <topic>
            template: <name of type template>
            type:
              type: <json or value>
              map:
                <key>: <value>
```
The job name, instance name and the sensor id will all be added as labels to the metrics
`sensor_temperature_celsius{instance="computer_room",job="zigbee",sensorid="computer_room"} 19.8`
The following example shows one job with two instances, one of the instances has one sensor defined and the other has two sensors defined
```yaml
jobs:
  - job: rfxcom
    topic_prefix: nodered/rfxcom
    instances:
      - name: outside
        sensors:
          - id: backside
            topic: outside/sensor
            template: rfxcom_sensor
      - name: kitchen
        sensors:
          - id: kitchen
            topic: kitchen/sensor
            template: rfxcom_sensor
          - id: ticker
            topic: ticker
            type:
              type: value
              map:
                metric: sensor_battery_level
```
### sensor types
Currently there are two different sensor types that can be configured and it is based on the payload they post on the message bus.
#### value
the *value* sensor type is the simpliest one.
```
type:
  type: value
  map:
    metric: <name of metric to publish on
```
This is used for sensors that just publish their sensor value in a specific topic. As an example if I have a sensor that publish its value to the following topic `nodered/rfxcom/ticker` you would configure it like this to use the sensor_battery_level metric
```
jobs:
  - job: rfxcom
    topic_prefix: nodered/rfxcom
    instances:
      - name: kitchen
        sensors:
          - id: ticker
            topic: ticker
            type:
              type: value
              map:
                metric: sensor_battery_level
```
When mqtt2prom catches this `nodered/rfxcom/ticker 56.7` on the message bus it will expose the following metric `sensor_battery_level{instance="kitchen",job="rfxcom",sensorid="ticker"} 56.7`
#### json
This sensor type is for json formed payloads. It is possible to have several metrics accosiated with the same payload.
The json sensor type is configured like this
```yaml
type:
  type: json
    <key>: <metric name>
    <key>: <metric name>
  ...
```
the `<key>` is what element that should be used for the metric and the levels are seperated with a dot.
As an example one of my sensors pubish the following payload 
`{"temperature":{"value":21.9,"unit":"degC"},"status":{"rssi":3,"battery":9}}`
The configuration for this would be
```yaml
jobs:
  - job: rfxcom
    topic_prefix: nodered/rfxcom
    instances:
      - name: kitchen
        sensors:
          - id: kitchen
            topic: kitchen/sensor
            type:
              type: json
              map:
                temperature.value: sensor_temperature_celsius
                status.rssi: sensor_linkquality_rssi
                status.battery: sensor_battery_level
```
When mqtt2prom cathes `nodered/rfxcom/kitchen/sensor {"temperature":{"value":21.9,"unit":"degC"},"status":{"rssi":4,"battery":9}}`on the message bus it will expose the following metrics
```
sensor_battery_level{instance="kitchen",job="rfxcom",sensorid="kitchen"} 9
sensor_linkquality_rssi{instance="kitchen",job="rfxcom",sensorid="kitchen"} 4
sensor_temperature_celsius{instance="kitchen",job="rfxcom",sensorid="kitchen"} 21.9
```
### templates
Instead of configure the sensor type and mappings for each sensor every time you can use templates of sensor types
```yaml
templates:
  rfxcom_sensor:
    type: json
    map:
      temperature.value: sensor_temperature_celsius
      humidity.value: sensor_humidity_percent
      status.rssi: sensor_linkquality_rssi
      status.battery: sensor_battery_level
```
and then you can refere to those templates when you configure the job/instances/sensors
```
jobs:
  - job: rfxcom
    topic_prefix: nodered/rfxcom
    instances:
      - name: outside
        sensors:
          - id: backside
            topic: outside/sensor
            template: rfxcom_sensor
      - name: kitchen
        sensors:
          - id: kitchen
            topic: kitchen/sensor
            template: rfxcom_sensor
```
These two sensors uses the same template although the "kitchen" sensor doesn't do humidity. 

Prometheus
----------
You need to configure the prometheus scraper for mqtt2prom as a push gateway by adding `honor_labels: true` in the prometheus configuration, read more [here](https://github.com/prometheus/pushgateway#about-the-job-and-instance-labels)
