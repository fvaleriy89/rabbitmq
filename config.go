package rabbitmq

import (
	"fmt"
)

type ConfigConnection struct {
	Host     string `json:"host"`
	Port     uint   `json:"port"`
	Login    string `json:"login"`
	Password string `json:"password"`
	Vhost    string `json:"vhost"`
}

type ConfigQueue struct {
	Name       string `json:"name"`
	Durable    bool   `json:"durable"`
	AutoDelete bool   `json:"auto-delete"`
	Exclusive  bool   `json:"exclusive"`
	NoWait     bool   `json:"no-wait"`

	Args       map[string]interface{} `json:"args"`

	Passive    bool   `json:"passive"`
}

type ConfigExchange struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	Durable    bool   `json:"durable"`
	AutoDelete bool   `json:"auto-delete"`
	Internal   bool   `json:"internal"`
	NoWait     bool   `json:"no-wait"`

	Args       map[string]interface{} `json:"args"`
}

type ConfigBinding struct {
	Queue      string `json:"queue"`
	Exchange   string `json:"exchange"`
	RoutingKey string `json:"routing-key"`
	NoWait     bool   `json:"no-wait"`

	Args       map[string]interface{} `json:"args"`
}

type ConfigConsumer struct {
	Count     int    `json:"count"`
	Queue     string `json:"queue"`
	Consumer  string `json:"consumer"`

	AutoAck   bool   `json:"auto-ack"`
	Exclusive bool   `json:"exclusive"`
	NoLocal   bool   `json:"no-local"`
	NoWait    bool   `json:"no-wait"`

	Args      map[string]interface{} `json:"args"`
}

type ConfigQos struct {
	PrefetchCount int  `json:"prefetch_count"`
	PrefetchSize  int  `json:"prefetch_size"`
	Global        bool `json:"global"`
}

type ConfigConflicts struct {
	Enabled           bool   `json:"enabled"`
	CheckIdleTTL      string `json:"check-idle-ttl"`
	CheckIdleInterval string `json:"check-idle-interval"`
}

type ConfigPublisher struct {
	Exchange   string `json:"exchange"`
	RoutingKey string `json:"routing-key"`
	Mandatory  bool   `json:"mandatory"`
	Immediate  bool   `json:"immediate"`
}

func (c ConfigConnection) Url() string {
	return fmt.Sprintf(
		"amqp091.//%s:%s@%s:%d%s",
		c.Login,
		c.Password,
		c.Host,
		c.Port,
		c.Vhost,
	)
}

func (cc ConfigConsumer) EnumConsumerTag(tag int) ConfigConsumer{
	cc.Consumer = fmt.Sprintf("%s_%04d", cc.Consumer, tag)
	return cc
}

var DefaultConfigConnection ConfigConnection = ConfigConnection{
	Host:     "127.0.0.1",
	Port:     5672,
	Login:    "guest",
	Password: "guest",
	Vhost:    "/",
}
var DefaultConfigQueue ConfigQueue = ConfigQueue{
	Name:       "",
	Durable:    true,
	AutoDelete: false,
	Exclusive:  false,
	NoWait:     false,
}
var DefaultConfigExchange ConfigExchange = ConfigExchange{
	Name:       "",
	Type:       "topic",
	Durable:    true,
	AutoDelete: false,
	Internal:   false,
	NoWait:     false,
}
var DefaultConfigBinding ConfigBinding = ConfigBinding{
	Queue:      "",
	Exchange:   "",
	RoutingKey: "",
	NoWait:     false,
}
var DefaultConfigConsumer ConfigConsumer = ConfigConsumer{
	Count: 1,
	Queue: "",
	Consumer: "",
	AutoAck: false,
	Exclusive: false,
	NoLocal: false,
	NoWait: false,
	Args: nil,
}
var DefaultConfigQos ConfigQos =  ConfigQos{
	PrefetchCount: 1,
	PrefetchSize: 0,
	Global: false,
}
var DefaultConfigConflicts ConfigConflicts = ConfigConflicts{
	Enabled: true,
	CheckIdleTTL: "15s",
	CheckIdleInterval: "3s",
}
var DefaultConfigPublisher ConfigPublisher = ConfigPublisher{
	Exchange: "",
	RoutingKey: "",
	Mandatory: false,
	Immediate: false,
}
