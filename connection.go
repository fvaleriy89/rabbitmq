package rabbitmq

import (
	"sync"

	"github.com/rabbitmq/amqp091-go"
)

func NewConnection(cfg ConfigConnection) *Connection {
	return &Connection{
		cfg: cfg,
	}
}

type Connection struct {
	cfg        ConfigConnection

	mx         sync.Mutex
	connection *amqp091.Connection
	channels   []*amqp091.Channel
}

func (this *Connection) Connect() error {
	this.mx.Lock()
	defer this.mx.Unlock()
	if this.connection != nil {
		return ErrorAlreadyConnected
	}
	connection, connectionError := amqp091.Dial(this.cfg.Url())
	if connectionError != nil {
		return connectionError
	}
	this.connection = connection
	return nil
}

func (this *Connection) GetChannel() (*amqp091.Channel, error) {
	if this == nil {
		return nil, ErrorMissedConnection
	}
	if e := this.Connect(); e != nil {
		if e != ErrorAlreadyConnected {
			return nil, e
		}
	}
	channel, channelError := this.connection.Channel()
	if channelError != nil {
		return nil, channelError
	}
	this.mx.Lock()
	defer this.mx.Unlock()
	this.channels = append(this.channels, channel)
	return channel, nil
}

func (this *Connection) CloseChannel(toclose *amqp091.Channel) error {
	this.mx.Lock()
	defer this.mx.Unlock()
	for pos, channel := range this.channels {
		if toclose == channel {
			newchannels := make([]*amqp091.Channel, len(this.channels)-1)
			copy(newchannels[:pos], this.channels[:pos])
			copy(newchannels[pos:], this.channels[pos+1:])

			this.channels = newchannels
			defer toclose.Close()
			break
		}
	}
	return nil
}

func (this *Connection) Disconnect() error {
	this.mx.Lock()
	defer this.mx.Unlock()
	if this.connection != nil {
		toclose := this.connection
		defer toclose.Close()
	}
	for _, channel := range this.channels {
		toclose := channel
		defer toclose.Close()
	}
	this.connection = nil
	this.channels = nil
	return nil
}
