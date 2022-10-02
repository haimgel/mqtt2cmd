package mqtt

import (
	"fmt"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/haimgel/mqtt-buttons/internal/config"
	"github.com/haimgel/mqtt-buttons/internal/controls"
	"go.uber.org/zap"
	"math/rand"
	"time"
)

const OnPayload = "ON"
const OffPayload = "OFF"

type Switch struct {
	control     controls.Switch
	state       bool
	lastRefresh time.Time
}

type Client struct {
	handle   MQTT.Client
	switches []*Switch
	logger   *zap.SugaredLogger
}

func Init(config *config.MqttConfig, controls []controls.Switch, logger *zap.SugaredLogger) (*Client, error) {
	client, err := Connect(config, controls, logger)
	if err != nil {
		return nil, err
	}
	err = client.Subscribe()
	if err != nil {
		client.handle.Disconnect(0)
		return nil, err
	}
	client.Refresh(true)
	return client, nil
}

func Connect(config *config.MqttConfig, controls []controls.Switch, logger *zap.SugaredLogger) (*Client, error) {
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
	switches := make([]*Switch, len(controls))
	for i, control := range controls {
		switches[i] = &Switch{control: control}
	}
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

func (client *Client) processSetPayload(sw *Switch, payload string) {
	defer client.syncLog()
	logger := client.logger.With(zap.String("switch", sw.control.Name))
	logger.Infow("Received switch command", "payload", payload)
	command, err := parsePayload(payload)
	if err != nil {
		logger.Error(err)
		return
	}
	response, err := sw.control.SwitchOnOff(command)
	if err != nil {
		logger.Errorw("Error running switch command", "error", err, "output", response)
		return
	}
	logger.Debugw("Executed switch command successfully", "output", response)
	client.setState(sw, command)
}

func (client *Client) Refresh(force bool) {
	defer client.syncLog()
	for _, sw := range client.switches {
		if force || (sw.control.RefreshInterval != 0 && time.Now().After(sw.lastRefresh.Add(sw.control.RefreshInterval))) {
			newState := sw.control.GetState()
			sw.lastRefresh = time.Now()
			if force || (newState != sw.state) {
				client.setState(sw, newState)
			}
		}
	}
}

func (client *Client) setState(sw *Switch, state bool) {
	topic := stateTopic(sw)
	logger := client.logger.With(zap.String("switch", sw.control.Name), zap.Bool("state", state), zap.String("topic", topic))
	token := client.handle.Publish(topic, 1, true, generatePayload(state))
	token.Wait()
	if token.Error() != nil {
		logger.Error("Error publishing state to MQTT", "error", token.Error())
		return
	}
	sw.state = state
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

func commandTopic(sw *Switch) string {
	return fmt.Sprintf("%s/switches/%s/set", config.AppName, sw.control.Name)
}

func stateTopic(sw *Switch) string {
	return fmt.Sprintf("%s/switches/%s", config.AppName, sw.control.Name)
}

func (client *Client) syncLog() {
	// noinspection GoUnhandledErrorResult
	defer client.logger.Sync()
}
