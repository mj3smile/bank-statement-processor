package eventbus

type EventConsumer struct {
	topic     Topic
	channelID ChannelID
}

func NewEventConsumer(topic Topic, channelID ChannelID) *EventConsumer {
	InitEventBus()
	return &EventConsumer{}
}

func (e *EventConsumer) Subscribe(topic Topic, channelID ChannelID) {
	
}
