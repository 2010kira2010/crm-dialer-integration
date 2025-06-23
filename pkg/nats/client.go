package nats

import (
	"encoding/json"

	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

type Client struct {
	conn   *nats.Conn
	logger *zap.Logger
}

func NewClient(url string, logger *zap.Logger) (*Client, error) {
	conn, err := nats.Connect(url)
	if err != nil {
		return nil, err
	}

	return &Client{
		conn:   conn,
		logger: logger,
	}, nil
}

func (c *Client) Publish(subject string, data interface{}) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return c.conn.Publish(subject, payload)
}

func (c *Client) Subscribe(subject string, handler func(msg *nats.Msg)) (*nats.Subscription, error) {
	return c.conn.Subscribe(subject, handler)
}

func (c *Client) Close() {
	c.conn.Close()
}
