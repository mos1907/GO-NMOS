package mqtt

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type Client struct {
	client      mqtt.Client
	enabled     bool
	topicPrefix string
}

type FlowEvent struct {
	Event     string                 `json:"event"` // created, updated, deleted
	FlowID    string                 `json:"flow_id"`
	Flow      map[string]interface{} `json:"flow,omitempty"`
	Diff      map[string]interface{} `json:"diff,omitempty"`
	Timestamp string                 `json:"timestamp"`
}

func NewClient(brokerURL, topicPrefix string, enabled bool) (*Client, error) {
	if !enabled {
		return &Client{enabled: false}, nil
	}

	opts := mqtt.NewClientOptions()
	opts.AddBroker(brokerURL)
	opts.SetClientID(fmt.Sprintf("go-nmos-backend-%d", time.Now().Unix()))
	opts.SetConnectTimeout(5 * time.Second)
	opts.SetAutoReconnect(true)
	opts.SetKeepAlive(30 * time.Second)
	opts.SetPingTimeout(5 * time.Second)

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, fmt.Errorf("mqtt connect failed: %w", token.Error())
	}

	log.Printf("MQTT client connected to %s", brokerURL)
	return &Client{
		client:      client,
		enabled:     true,
		topicPrefix: topicPrefix,
	}, nil
}

func (c *Client) PublishFlowEvent(event string, flowID string, flow map[string]interface{}, diff map[string]interface{}) {
	if !c.enabled {
		return
	}

	payload := FlowEvent{
		Event:     event,
		FlowID:    flowID,
		Flow:      flow,
		Diff:      diff,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("mqtt: failed to marshal event: %v", err)
		return
	}

	topicAll := fmt.Sprintf("%s/all", c.topicPrefix)
	topicFlow := fmt.Sprintf("%s/flow/%s", c.topicPrefix, flowID)

	if token := c.client.Publish(topicAll, 0, false, data); token.Wait() && token.Error() != nil {
		log.Printf("mqtt: failed to publish to %s: %v", topicAll, token.Error())
	}
	if token := c.client.Publish(topicFlow, 0, false, data); token.Wait() && token.Error() != nil {
		log.Printf("mqtt: failed to publish to %s: %v", topicFlow, token.Error())
	}
}

func (c *Client) Close() {
	if c.enabled && c.client != nil && c.client.IsConnected() {
		c.client.Disconnect(250)
	}
}
