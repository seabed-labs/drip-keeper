package eventbus

import "github.com/asaskevich/EventBus"

type EventBusTopic string

const (
	VaultConfigTopic = EventBusTopic("vault_config_topic")
)

func NewEventBus() EventBus.Bus {
	return EventBus.New()
}
