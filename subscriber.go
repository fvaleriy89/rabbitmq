package rabbitmq

import (
	"fmt"
	"time"
	"sync"

	"github.com/streadway/amqp"
)

const WARNING_DURATION time.Duration = 100 * time.Millisecond

func NewSubscriber() *Subscriber {
	return &Subscriber{
		configConnection: DefaultConfigConnection,
		configQos: DefaultConfigQos,
		configConsumer: DefaultConfigConsumer,
		configConflicts: DefaultConfigConflicts,
		resolver: NewConflictResolver(),
		errors: make(chan error, 1024),
	}
}

type ProcessableParser interface {
	Match(key string) bool
	Parse(key string, msg []byte) (ProcessableEntity, error)
}

type ProcessableEntity interface {
	Process() error
	EntityID() string // conflict tracking
	MarkConflict() error
}

type processingCallback func(key string, body []byte, e error, t time.Duration)

type Subscriber struct {
	mx                 sync.Mutex

	configConnection   ConfigConnection
	configQos          ConfigQos
	configConsumer     ConfigConsumer
	configConflicts    ConfigConflicts

	channel            *amqp.Channel
	connection         *Connection
	resolver           ConflictResolver // *resolver
	parsers            []ProcessableParser
	processingCallback processingCallback

	errors             chan error
}

func (this *Subscriber) ConfigConnection(cfg ConfigConnection) *Subscriber {
	this.configConnection = cfg
	return this
}

func (this *Subscriber) SetConnection(connection *Connection) *Subscriber {
	this.connection = connection
	return this
}

func (this *Subscriber) SetChannel(channel *amqp.Channel) *Subscriber {
	this.mx.Lock()
	defer this.mx.Unlock()
	this.channel = channel
	return this
}

func (this *Subscriber) ConfigQos(cfg ConfigQos) *Subscriber {
	this.configQos = cfg
	return this
}

func (this *Subscriber) ConfigConsumer(cfg ConfigConsumer) *Subscriber {
	this.configConsumer = cfg
	return this
}

func (this *Subscriber) ConfigConflicts(cfg ConfigConflicts) *Subscriber {
	this.configConflicts = cfg
	return this
}

func (this *Subscriber) PostProcessingCallback(f processingCallback) *Subscriber {
	this.processingCallback = f
	return this
}

func (this *Subscriber) UnresolvedLocksCallback(fn func(string)) error {
	interval, e1 := time.ParseDuration(this.configConflicts.CheckIdleInterval)
	if e1 != nil {
		return e1
	}

	ttl, e2 := time.ParseDuration(this.configConflicts.CheckIdleTTL)
	if e2 != nil {
		return e2
	}

	this.resolver.CheckLocks(interval, ttl, fn, nil)

	return nil
}

func (this *Subscriber) CheckLocks(errfn func(string), infofn func(string)) error {
	interval, e1 := time.ParseDuration(this.configConflicts.CheckIdleInterval)
	if e1 != nil {
		return e1
	}

	ttl, e2 := time.ParseDuration(this.configConflicts.CheckIdleTTL)
	if e2 != nil {
		return e2
	}

	this.resolver.CheckLocks(interval, ttl, errfn, infofn)

	return nil
}

func (this *Subscriber) Channel() (*amqp.Channel, error) {
	this.mx.Lock()
	defer this.mx.Unlock()
	if this.channel == nil {
		if this.connection == nil {
			this.connection = NewConnection(this.configConnection)
		}
		channel, channelError := this.connection.GetChannel()
		if channelError != nil {
			return nil, channelError
		}
		qosError := channel.Qos(
			this.configQos.PrefetchCount,
			this.configQos.PrefetchSize,
			this.configQos.Global,
		)
		if qosError != nil {
			defer this.connection.CloseChannel(channel)
			return nil, qosError
		}
		this.channel = channel
	}
	return this.channel, nil
}

func (s *Subscriber) Listen(parsers ...ProcessableParser) error {
	s.parsers = parsers

	for i := 0; i < s.configConsumer.Count; i++ {
		if e := s.consume(s.configConsumer.EnumConsumerTag(i)); e != nil {
			return e
		}
	}

	return nil
}

func (this *Subscriber) Wait() (<-chan error) {
	return this.errors
}

func (this *Subscriber) consume(cfg ConfigConsumer) error {
	channel, channelError := this.Channel()
	if nil != channelError {
		return channelError
	}

	if e := amqp.Table(cfg.Args).Validate(); e != nil {
		return e
	}

	// TODO: channel.Cancel(cfg.Consumer, cfg.NoWait)
	msgs, msgsError := channel.Consume(
		cfg.Queue,
		cfg.Consumer,
		cfg.AutoAck,
		cfg.Exclusive,
		cfg.NoLocal,
		cfg.NoWait,
		cfg.Args,
	)
	if nil != msgsError {
		return msgsError
	}

	go this.processing(cfg, msgs)

	return nil
}

func (this *Subscriber) processing(cfg ConfigConsumer, msgs <-chan amqp.Delivery) {
	for msg := range msgs {
		started := time.Now()
		processingError := this.process(msg)
		processingDuration := time.Since(started)

		// TODO: WARNING_DURATION to config
		if processingError == nil && processingDuration > WARNING_DURATION {
			processingError = ErrorProcessingDuration
		}

		if cb := this.processingCallback; cb != nil {
			cb(msg.RoutingKey, msg.Body, processingError, processingDuration)
		}

		if !cfg.AutoAck {
			msg.Ack(false) // non-multiple acknowledgement
		}
	}

	this.errors <- fmt.Errorf(`Consumer "%s" finished processing`, cfg.Consumer)
}

func (this *Subscriber) process(msg amqp.Delivery) error {
	entity, parsingError := this.parse(msg)
	if parsingError != nil {
		return parsingError
	}

	if this.configConflicts.Enabled {
		eid, conflict := this.resolver.Enter(entity.EntityID(), msg.DeliveryTag)
		defer this.resolver.Leave(entity.EntityID(), eid)
		if conflict {
			if e := entity.MarkConflict(); e != nil {
				return e
			}
		}
	}

	if e := entity.Process(); e != nil {
		return e
	}

	return nil
}

func (this *Subscriber) parse(msg amqp.Delivery) (ProcessableEntity, error) {
	for _, p := range this.parsers {
		if p.Match(msg.RoutingKey) {
			return p.Parse(msg.RoutingKey, msg.Body)
		}
	}
	return nil, ErrorUnprocessable
}
