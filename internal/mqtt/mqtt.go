package mqtt

import (
	"fmt"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/haimgel/mqtt2cmd/internal/config"
	"github.com/haimgel/mqtt2cmd/internal/controls"
	"go.uber.org/zap"
	"math/rand"
	"time"
)

const onPayload = "ON"
const offPayload = "OFF"
const availablePayload = "online"
const unavailablePayload = "offline"

type Switch struct {
	control      controls.Switch
	state        bool
	stateSet     bool
	available    bool
	availableSet bool
	lastRefresh  time.Time
}

type Client struct {
	appName  string
	handle   MQTT.Client
	switches []*Switch
	logger   *zap.SugaredLogger
}

func Init(appName string, config *config.MqttConfig, controls []controls.Switch, logger *zap.SugaredLogger) (*Client, error) {
	client, err := Connect(appName, config, controls, logger)
	if err != nil {
		return nil, err
	}
	err = client.Subscribe()
	if err != nil {
		client.handle.Disconnect(0)
		return nil, err
	}
	client.setAppAvailable()
	return client, nil
}

func Connect(appName string, config *config.MqttConfig, controls []controls.Switch, logger *zap.SugaredLogger) (*Client, error) {
	opts := MQTT.NewClientOptions()
	opts.AddBroker(config.Broker)
	opts.SetOrderMatters(false)
	opts.SetClientID(generateClientId(appName))
	opts.SetWill(appAvailabilityTopic(appName), unavailablePayload, 1, true)
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
	return &Client{appName: appName, handle: client, switches: switches, logger: logger}, nil
}

func generateClientId(appName string) string {
	return fmt.Sprintf("%s-%016x", appName, rand.Uint64())
}

func (client *Client) Subscribe() error {
	for _, sw := range client.switches {
		topic := client.commandTopic(sw)
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
		client.setAvailable(sw, false)
		return
	}
	logger.Debugw("Executed switch command successfully", "output", response)
	client.setState(sw, command)
}

func (client *Client) Refresh() {
	defer client.syncLog()
	for _, sw := range client.switches {
		client.refreshOne(sw)
	}
}

func (client *Client) refreshOne(sw *Switch) {
	logger := client.logger.With(zap.String("switch", sw.control.Name))
	if !sw.availableSet || !sw.stateSet || (sw.control.RefreshInterval != 0 && time.Now().After(sw.lastRefresh.Add(sw.control.RefreshInterval))) {
		newState, response, err := sw.control.GetState()
		if err != nil {
			logger.Errorw("Error running switch query command", "error", err, "output", response)
		}
		sw.lastRefresh = time.Now()
		client.setState(sw, newState)
		client.setAvailable(sw, err == nil)
	}
}

func (client *Client) setState(sw *Switch, state bool) {
	if sw.stateSet && sw.state == state {
		return
	}
	topic := client.stateTopic(sw)
	logger := client.logger.With(
		zap.String("switch", sw.control.Name),
		zap.String("topic", topic),
		zap.Bool("state", state),
	)
	token := client.handle.Publish(topic, 1, true, generateStatePayload(state))
	token.Wait()
	if token.Error() != nil {
		logger.Error("Error publishing state to MQTT", "error", token.Error())
		return
	}
	sw.state = state
	sw.stateSet = true
	logger.Debugw("Published state to MQTT")
}

func (client *Client) setAvailable(sw *Switch, available bool) {
	if sw.availableSet && sw.available == available {
		return
	}
	topic := client.availabilityTopic(sw)
	logger := client.logger.With(
		zap.String("switch", sw.control.Name),
		zap.String("topic", topic),
		zap.Bool("available", available),
	)
	token := client.handle.Publish(topic, 1, true, generateAvailablePayload(available))
	token.Wait()
	if token.Error() != nil {
		logger.Error("Error publishing availability to MQTT", "error", token.Error())
		return
	}
	sw.available = available
	sw.availableSet = true
	logger.Debugw("Published availability to MQTT")
}

func (client *Client) setAppAvailable() {
	topic := appAvailabilityTopic(client.appName)
	logger := client.logger.With(zap.String("topic", topic))
	token := client.handle.Publish(topic, 1, false, generateAvailablePayload(true))
	token.Wait()
	if token.Error() != nil {
		logger.Error("Error publishing application availability to MQTT", "error", token.Error())
	}
	logger.Debugw("Published application availability to MQTT")
}

func parsePayload(payload string) (bool, error) {
	if payload == onPayload {
		return true, nil
	} else if payload == offPayload {
		return false, nil
	} else {
		return false, fmt.Errorf("invalid payload: %s", payload)
	}
}

func generateStatePayload(state bool) string {
	if state {
		return onPayload
	} else {
		return offPayload
	}
}

func generateAvailablePayload(available bool) string {
	if available {
		return availablePayload
	} else {
		return unavailablePayload
	}
}

func (client *Client) commandTopic(sw *Switch) string {
	return fmt.Sprintf("%s/switches/%s/set", client.appName, sw.control.Name)
}

func (client *Client) stateTopic(sw *Switch) string {
	return fmt.Sprintf("%s/switches/%s", client.appName, sw.control.Name)
}

func (client *Client) availabilityTopic(sw *Switch) string {
	return fmt.Sprintf("%s/switches/%s/available", client.appName, sw.control.Name)
}

func appAvailabilityTopic(appName string) string {
	return fmt.Sprintf("%s/available", appName)
}

func (client *Client) syncLog() {
	// noinspection GoUnhandledErrorResult
	defer client.logger.Sync()
}
