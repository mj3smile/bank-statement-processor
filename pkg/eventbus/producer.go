package eventbus

type EventProducer struct {
	topic     Topic
	channelID ChannelID
}

func NewEventProducer() *EventProducer {
	InitEventBus()
	return &EventProducer{}
}

func (ep *EventProducer) Publish(topic Topic, payload interface{}) {
	channels := eventBus.GetSubscribers(topic)
	if channels == nil {
		return
	}

	go func() {
		for _, subscriber := range channels[ep.channelID] {
			subscriber <- Event{Payload: payload}
		}
	}()
}
