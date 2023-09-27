package rabbitmq

import (
	"sync"

	"github.com/streadway/amqp"
)

func NewQueue(cfg ConfigQueue) *Queue {
	return (&Queue{
		configConnection: DefaultConfigConnection,
	}).ConfigQueue(cfg)
}

type Queue struct {
	mx               sync.Mutex

	configConnection ConfigConnection
	configQueue      ConfigQueue

	connection       *Connection
	channel          *amqp.Channel
}

func (this *Queue) ConfigQueue(cfg ConfigQueue) *Queue {
	this.configQueue = cfg
	return this
}

func (this *Queue) ConfigConnection(cfg ConfigConnection) *Queue {
	this.configConnection = cfg
	return this
}

func (this *Queue) SetConnection(connection *Connection) *Queue {
	this.connection = connection
	return this
}

func (this *Queue) SetChannel(channel *amqp.Channel) *Queue {
	this.mx.Lock()
	defer this.mx.Unlock()
	this.channel = channel
	return this
}

func (this *Queue) Channel() (*amqp.Channel, error) {
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

func (this *Queue) Declare() (*amqp.Queue, error) {
	channel, channelError := this.Channel()
	if channelError != nil {
		return nil, channelError
	}

        if e := amqp.Table(this.configQueue.Args).Validate(); e != nil {
                return nil, e
        }

	queue, queueError := channel.QueueDeclare(
		this.configQueue.Name,
		this.configQueue.Durable,
		this.configQueue.AutoDelete,
		this.configQueue.Exclusive,
		this.configQueue.NoWait,
		this.configQueue.Args,
	)
	if queueError != nil {
		return nil, queueError
	}
	return &queue, nil
}

func (this *Queue) GetInfo() (*amqp.Queue, error) {
	channel, channelError := this.Channel()
	if channelError != nil {
		return nil, channelError
	}
	queue, queueError := channel.QueueInspect(this.configQueue.Name)
	if queueError != nil {
		return nil, queueError
	}
	return &queue, nil
}

func (this *Queue) Purge() (int, error) {
	channel, channelError := this.Channel()
	if channelError != nil {
		return 0, channelError
	}
	return channel.QueuePurge(this.configQueue.Name, false)
}
