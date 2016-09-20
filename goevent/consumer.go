package goevent

import (
	"amqp"
	"encoding/json"
	"fmt"
	"log"
)

type Consumer struct {
	queue    string
	conn     *amqp.Connection
	channel  *amqp.Channel
	delivery <-chan amqp.Delivery
}

func NewConsumer(amqpURI string, queueName string) (*Consumer, error) {
	c := &Consumer{
		queue:    queueName,
		conn:     nil,
		channel:  nil,
		delivery: nil,
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

	log.Printf("starting Consume (consumer tag %q)", c.queue)
	c.delivery, err = c.channel.Consume(
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

	return nil
}

func Subscribe(amqpURI string, queueName string, h interface{}) error {
	handler, ok := h.(Handler)
	if !ok {
		return fmt.Errorf("eee")
	}
	c, err := NewConsumer(amqpURI, queueName)
	if err != nil {
		fmt.Errorf("%s", err)
	}
	defer func() {
		if err := c.Shutdown(); err != nil {
			log.Fatalf("error during shutdown: %s", err)
		}
	}()

	for {
		d := <-c.delivery
		if json.Unmarshal(d.Body, h); err != nil {
			d.Reject(true)
			return nil
		}
		if err := handler.Handle(); err != nil {
			d.Reject(true)
		} else {
			d.Ack(false)
		}
	}
}