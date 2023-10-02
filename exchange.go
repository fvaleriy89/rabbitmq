package rabbitmq

import (
	"sync"

	"github.com/rabbitmq/amqp091-go"
)

func NewExchange() *Exchange {
	return &Exchange{
		configConnection: DefaultConfigConnection,
		configExchange: DefaultConfigExchange,
	}
}

type Exchange struct {
	mx               sync.Mutex

	configConnection ConfigConnection
	configExchange   ConfigExchange

	connection       *Connection
	channel          *amqp091.Channel
}

func (this *Exchange) ConfigConnection(cfg ConfigConnection) *Exchange {
	this.configConnection = cfg
	return this
}

func (this *Exchange) ConfigExchange(cfg ConfigExchange) *Exchange {
	this.configExchange = cfg
	return this
}

func (this *Exchange) SetConnection(connection *Connection) *Exchange {
	this.connection = connection
	return this
}

func (this *Exchange) SetChannel(channel *amqp091.Channel) *Exchange {
	this.mx.Lock()
	defer this.mx.Unlock()
	this.channel = channel
	return this
}

func (this *Exchange) Channel() (*amqp091.Channel, error) {
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
		this.channel = channel
	}
	return this.channel, nil
}

func (this *Exchange) Declare() error {
	channel, channelError := this.Channel()
	if channelError != nil {
		return channelError
	}

	if e := amqp091.Table(this.configExchange.Args).Validate(); e != nil {
		return e
	}

	return channel.ExchangeDeclare(
		this.configExchange.Name,
		this.configExchange.Type,
		this.configExchange.Durable,
		this.configExchange.AutoDelete,
		this.configExchange.Internal,
		this.configExchange.NoWait,
		this.configExchange.Args,
	)
}

func (this *Exchange) GetName() string {
	return this.configExchange.Name
}
