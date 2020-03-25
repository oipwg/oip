package events

import (
	"github.com/asaskevich/EventBus"
)

var bus = EventBus.New()

// SubscribeAsync subscribes to a topic with an asynchronous callback
// Subsequent callbacks for a topic are run concurrently
// Does nothing if fn is not a function.
func SubscribeAsync(topic string, fn interface{}) {
	_ = bus.SubscribeAsync(topic, fn, false)
}

// SubscribeOnceAsync subscribes to a topic once with an asynchronous callback
// Handler will be removed after executing.
// Does nothing if fn is not a function.
func SubscribeOnceAsync(topic string, fn interface{}) {
	_ = bus.SubscribeOnceAsync(topic, fn)
}

// Unsubscribe removes callback defined for a topic if it exists.
func Unsubscribe(topic string, handler interface{}) {
	_ = bus.Unsubscribe(topic, handler)
}

// Publish executes callback defined for a topic. Any additional argument will be transferred to the callback.
func Publish(topic string, args ...interface{}) {
	bus.Publish(topic, args...)
}
