package mqtt

import (
	"fmt"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/haimgel/mqtt-buttons/internal/config"
	"github.com/haimgel/mqtt-buttons/internal/controls"
	"go.uber.org/zap"
	"math/rand"
)

const OnPayload = "ON"
const OffPayload = "OFF"

type Client struct {
	handle   MQTT.Client
	switches []controls.Switch
	logger   *zap.SugaredLogger
}

func Connect(config *config.MqttConfig, switches []controls.Switch, logger *zap.SugaredLogger) (*Client, error) {
	opts := MQTT.NewClientOptions()
	opts.AddBroker(config.Broker)
	opts.SetOrderMatters(false)
	opts.SetClientID(generateClientId())
	if (config.User != nil) && (config.Password != nil) {
		opts.SetUsername(*config.User)
		opts.SetPassword(*config.Password)
	}
	client := MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}
	logger.Infow("Connected to MQTT", "broker", config.Broker)
	return &Client{handle: client, switches: switches, logger: logger}, nil
}

func generateClientId() string {
	return fmt.Sprintf("%s-%016x", config.AppName, rand.Uint64())
}

func (client *Client) Subscribe() error {
	for _, sw := range client.switches {
		topic := commandTopic(sw)
		client.logger.Debugw("Subscribing", "topic", topic)
		if token := client.handle.Subscribe(topic, 1, func(mqttClient MQTT.Client, message MQTT.Message) {
			client.processSetPayload(sw, string(message.Payload()))
		}); token.Wait() && token.Error() != nil {
			return token.Error()
		}
	}
	return nil
}

func (client *Client) processSetPayload(sw controls.Switch, payload string) {
	logger := client.logger.With(zap.String("switch", sw.Name))
	logger.Infow("Received switch command", "payload", payload)
	command, err := parsePayload(payload)
	if err != nil {
		logger.Error(err)
		return
	}
	response, err := sw.SwitchOnOff(command)
	if err != nil {
		logger.Errorw("Error running switch command", "error", err, "output", response)
		return
	}
	logger.Debugw("Executed switch command successfully", "output", response)
	client.setState(sw, command)
}

func (client *Client) setState(sw controls.Switch, state bool) {
	logger := client.logger.With(zap.String("switch", sw.Name))
	token := client.handle.Publish(stateTopic(sw), 1, true, generatePayload(state))
	token.Wait()
	if token.Error() != nil {
		logger.Error("Error publishing state to MQTT", "error", token.Error())
		return
	}
	logger.Debugw("Published state to MQTT")
}

func parsePayload(payload string) (bool, error) {
	if payload == OnPayload {
		return true, nil
	} else if payload == OffPayload {
		return false, nil
	} else {
		return false, fmt.Errorf("invalid payload: %s", payload)
	}
}

func generatePayload(state bool) string {
	if state {
		return OnPayload
	} else {
		return OffPayload
	}
}

func commandTopic(sw controls.Switch) string {
	return fmt.Sprintf("%s/switches/%s/set", config.AppName, sw.Name)
}

func stateTopic(sw controls.Switch) string {
	return fmt.Sprintf("%s/switches/%s", config.AppName, sw.Name)
}
