package rabbitmq

import (
	"sync"

	"github.com/streadway/amqp"
)

const DEFAULT_MESSAGE_CONTENT_TYPE = "text/plain"

type PublishMessage interface {
	BuildRoutingKey() string
	EncodePushMessage() []byte
	WithActor(name string) PublishMessage
}

func NewPublisher() *Publisher {
	return &Publisher{
		configConnection: DefaultConfigConnection,
		configPublisher: DefaultConfigPublisher,
	}
}

type Publisher struct {
	mx               sync.Mutex

	configConnection ConfigConnection
	configPublisher  ConfigPublisher

	channel          *amqp.Channel
	connection       *Connection
}

func (this *Publisher) ConfigConnection(cfg ConfigConnection) *Publisher {
	this.configConnection = cfg
	return this
}

func (this *Publisher) ConfigPublisher(cfg ConfigPublisher) *Publisher {
	this.configPublisher = cfg
	return this
}

func (this *Publisher) SetConnection(connection *Connection) *Publisher {
	this.connection = connection
	return this
}

func (this *Publisher) SetChannel(channel *amqp.Channel) *Publisher {
	this.mx.Lock()
	defer this.mx.Unlock()
	this.channel = channel
	return this
}

func (this *Publisher) Channel() (*amqp.Channel, error) {
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

func (this *Publisher) Disconnected(e error) bool {
	if amqpError, ok := e.(*amqp.Error); ok && (amqpError.Code == amqp.ChannelError) {
		return true
	}
	return false
}

func (this *Publisher) Publish(body []byte, opts ...PublishOption) error {
	channel, channelError := this.Channel()
	if channelError != nil {
		return channelError
	}

	publish := Publish{
		Exchange: this.configPublisher.Exchange,
		RoutingKey: this.configPublisher.RoutingKey,
		Mandatory: this.configPublisher.Mandatory,
		Immediate: this.configPublisher.Immediate,
		Message: amqp.Publishing{
			ContentType: DEFAULT_MESSAGE_CONTENT_TYPE,
			Body: body,
		},
	}
	for _, o := range opts {
		o(&publish)
	}

	return channel.Publish(
		publish.Exchange,
		publish.RoutingKey,
		publish.Mandatory,
		publish.Immediate,
		publish.Message,
	)
}

type Publish struct {
	Exchange   string
	RoutingKey string
	Mandatory  bool
	Immediate  bool

	// https://pkg.go.dev/github.com/streadway/amqp#Publishing
	Message    amqp.Publishing
}

type PublishOption func(*Publish)

/*func PubExchange(exchange string) PublishOption {
	return func(p *Publish) {
		p.Exchange = exchange
	}
}*/
func PubRoutingKey(routingKey string) PublishOption {
	return func(p *Publish) {
		p.RoutingKey = routingKey
	}
}
func PubMandatory(mandatory bool) PublishOption {
	return func(p *Publish) {
		p.Mandatory = mandatory
	}
}
func PubImmediate(immediate bool) PublishOption {
	return func(p *Publish) {
		p.Immediate = immediate
	}
}
func PubMessageContentType(contentType string) PublishOption {
	return func(p *Publish) {
		p.Message.ContentType = contentType
	}
}
func PubMessageBody(body []byte) PublishOption {
	return func(p *Publish) {
		p.Message.Body = body
	}
}
