package main

import (
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

func main() {
	c, err := NewConsumer("amqp://guest:guest@localhost:5672/", "queue")
	defer func() {
		if err := c.Shutdown(); err != nil {
			log.Fatalf("error during shutdown: %s", err)
		}
	}()
	if err != nil {
		log.Fatalf("%s", err)
	}
	select {}
}

type Consumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	queue   string
	done    chan error
}

func NewConsumer(amqpURI string, queueName string) (*Consumer, error) {
	c := &Consumer{
		conn:    nil,
		channel: nil,
		queue:   queueName,
		done:    make(chan error),
	}

	var err error

	log.Printf("dialing %q", amqpURI)
	c.conn, err = amqp.Dial(amqpURI)
	if err != nil {
		return nil, fmt.Errorf("Dial: %s", err)
	}

	go func() {
		fmt.Printf("closing: %s", <-c.conn.NotifyClose(make(chan *amqp.Error)))
	}()

	log.Printf("got Connection, getting Channel")
	c.channel, err = c.conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("Channel: %s", err)
	}

	log.Printf("got Channel, declaring Exchange (%q)", "exchange")
	if err = c.channel.ExchangeDeclare(
		"exchange", // name of the exchange
		"direct",   // type
		true,       // durable
		false,      // delete when complete
		false,      // internal
		false,      // noWait
		nil,        // arguments
	); err != nil {
		return nil, fmt.Errorf("Exchange Declare: %s", err)
	}

	log.Printf("declared Exchange, declaring Queue %q", queueName)
	queue, err := c.channel.QueueDeclare(
		c.queue, // name of the queue
		true,    // durable
		false,   // delete when usused
		false,   // exclusive
		false,   // noWait
		nil,     // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("Queue Declare: %s", err)
	}

	log.Printf("declared Queue (%q %d messages, %d consumers), binding to Exchange (key %q)",
		queue.Name, queue.Messages, queue.Consumers, "key")

	if err = c.channel.QueueBind(
		queue.Name, // name of the queue
		"key",      // bindingKey
		"exchange", // sourceExchange
		false,      // noWait
		nil,        // arguments
	); err != nil {
		return nil, fmt.Errorf("Queue Bind: %s", err)
	}
	log.Printf("starting Consume (consumer tag %q)", c.queue)
	deliveries, err := c.channel.Consume(
		c.queue, // name
		c.queue, // consumerTag,
		false,   // noAck
		false,   // exclusive
		false,   // noLocal
		false,   // noWait
		nil,     // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("Queue Consume: %s", err)
	}

	go handle(deliveries, c.done)

	return c, nil
}

func (c *Consumer) Shutdown() error {
	// will close() the deliveries channel
	if err := c.channel.Cancel(c.queue, true); err != nil {
		return fmt.Errorf("Consumer cancel failed: %s", err)
	}

	if err := c.conn.Close(); err != nil {
		return fmt.Errorf("AMQP connection close error: %s", err)
	}

	defer log.Printf("AMQP shutdown OK")

	// wait for handle() to exit
	return <-c.done
}

func handle(deliveries <-chan amqp.Delivery, done chan error) {
	for d := range deliveries {
		log.Printf(
			"got %dB delivery: [%v] %q",
			len(d.Body),
			d.DeliveryTag,
			d.Body,
		)
		d.Ack(false)
	}
	log.Printf("handle: deliveries channel closed")
	done <- nil
}
