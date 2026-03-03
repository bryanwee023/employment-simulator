package mail

import (
	"encoding/json"
	"errors"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

var ErrQueueNotFound = errors.New("recipient queue does not exist")

type Mailer struct {
	conn       *amqp.Connection
	pubCh      *amqp.Channel
	senderID   string
	senderName string
	senderRole string
}

func NewMailer(amqpURL, senderID, senderName, senderRole string) (*Mailer, error) {
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	pubCh, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open publishing channel: %w", err)
	}

	return &Mailer{
		conn:       conn,
		pubCh:      pubCh,
		senderID:   senderID,
		senderName: senderName,
		senderRole: senderRole,
	}, nil
}

func (m *Mailer) Send(to, subject, body string) error {
	queueName := "emails." + to

	// Passive declare to check queue exists (uses a temporary channel
	// because a failed passive declare closes the channel)
	ch, err := m.conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}
	_, err = ch.QueueDeclarePassive(queueName, true, false, false, false, nil)
	ch.Close()
	if err != nil {
		return fmt.Errorf("%w: %s", ErrQueueNotFound, queueName)
	}

	email := map[string]interface{}{
		"from": map[string]string{
			"id":   m.senderID,
			"name": m.senderName,
			"role": m.senderRole,
		},
		"subject": subject,
		"body":    body,
	}

	data, err := json.Marshal(email)
	if err != nil {
		return fmt.Errorf("failed to marshal email: %w", err)
	}

	return m.pubCh.Publish("", queueName, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        data,
	})
}

func (m *Mailer) Receive(employeeID string) (<-chan amqp.Delivery, error) {
	ch, err := m.conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open consumer channel: %w", err)
	}

	queueName := "emails." + employeeID
	_, err = ch.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		ch.Close()
		return nil, fmt.Errorf("failed to declare queue %s: %w", queueName, err)
	}

	msgs, err := ch.Consume(queueName, "", false, false, false, false, nil)
	if err != nil {
		ch.Close()
		return nil, fmt.Errorf("failed to start consuming: %w", err)
	}

	return msgs, nil
}
