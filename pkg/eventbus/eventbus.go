package eventbus

import "sync"

type Event struct {
	Payload interface{}
}

type (
	EventChan chan Event
	Topic     string
	ChannelID string
)

type EventBus struct {
	mutex              sync.RWMutex
	topicToSubscribers map[Topic]map[ChannelID][]EventChan
}

var (
	eventBus *EventBus
)

func InitEventBus() {
	if eventBus != nil {
		return
	}

	eventBus = &EventBus{}
}

func NewEventBus() *EventBus {
	return &EventBus{
		topicToSubscribers: make(map[Topic]map[ChannelID][]EventChan),
	}
}

func (eb *EventBus) GetSubscribers(topic Topic) map[ChannelID][]EventChan {
	eb.mutex.RLock()
	defer eb.mutex.RUnlock()
	var subscribers map[ChannelID][]EventChan
	for channelID, subscriber := range eb.topicToSubscribers[topic] {
		subscribers[channelID] = subscriber
	}
	return subscribers
}

func (eb *EventBus) CreateTopic(topic Topic) bool {
	_, ok := eb.topicToSubscribers[topic]
	return ok
}

func (eb *EventBus) IsTopicExist(topic Topic) bool {
	_, ok := eb.topicToSubscribers[topic]
	return ok
}

func (eb *EventBus) Publish(topic string, event Event) {
	eb.mutex.RLock()
	defer eb.mutex.RUnlock()
	subscribers := append([]EventChan{}, eb.subscribers[topic]...)
	go func() {
		for _, subscriber := range subscribers {
			subscriber <- event
		}
	}()
}

func (eb *EventBus) Subscribe(topic string) EventChan {
	eb.mutex.Lock()
	defer eb.mutex.Unlock()
	ch := make(EventChan)
	eb.subscribers[topic] = append(eb.subscribers[topic], ch)
	return ch
}

func (eb *EventBus) Unsubscribe(topic string, ch EventChan) {
	eb.mutex.Lock()
	defer eb.mutex.Unlock()
	if subscribers, ok := eb.subscribers[topic]; ok {
		for i, subscriber := range subscribers {
			if ch == subscriber {
				eb.subscribers[topic] = append(subscribers[:i], subscribers[i+1:]...)
				close(ch)
				for range ch {
				}
				return
			}
		}
	}
}
