package rabbitmq

import (
	"sync"

	"github.com/streadway/amqp"
)

func NewBinding() *Binding {
	return &Binding{
		configConnection: DefaultConfigConnection,
		configBinding: DefaultConfigBinding,
	}
}

type Binding struct {
	mx               sync.Mutex

	configConnection ConfigConnection
	configBinding    ConfigBinding

	connection       *Connection
	channel          *amqp.Channel
}


func (this *Binding) ConfigConnection(cfg ConfigConnection) *Binding {
	this.configConnection = cfg
	return this
}

func (this *Binding) ConfigBinding(cfg ConfigBinding) *Binding {
	this.configBinding = cfg
	return this
}

func (this *Binding) SetConnection(connection *Connection) *Binding {
	this.connection = connection
	return this
}

func (this *Binding) SetChannel(channel *amqp.Channel) *Binding {
	this.mx.Lock()
	defer this.mx.Unlock()
	this.channel = channel
	return this
}

func (this *Binding) Channel() (*amqp.Channel, error) {
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

func (this *Binding) Declare() error {
	channel, channelError := this.Channel()
	if channelError != nil {
		return channelError
	}
	return channel.QueueBind(
		this.configBinding.Queue,
		this.configBinding.RoutingKey,
		this.configBinding.Exchange,
		this.configBinding.NoWait,
		this.configBinding.Args,
	)
}
