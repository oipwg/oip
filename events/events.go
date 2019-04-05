package events

import (
	"github.com/asaskevich/EventBus"
)

var bus = EventBus.New()

// Subscribe subscribes to a topic.
// Does nothing if fn is not a function.
func Subscribe(topic string, fn interface{}) {
	_ = bus.Subscribe(topic, fn)
}

// SubscribeAsync subscribes to a topic with an asynchronous callback
// Transactional determines whether subsequent callbacks for a topic are
// run serially (true) or concurrently (false)
// Does nothing if fn is not a function.
func SubscribeAsync(topic string, fn interface{}, transactional bool) {
	_ = bus.SubscribeAsync(topic, fn, transactional)
}

// SubscribeOnce subscribes to a topic once. Handler will be removed after executing.
// Returns error if `fn` is not a function.
// Does nothing if fn is not a function.
func SubscribeOnce(topic string, fn interface{}) {
	_ = bus.SubscribeOnce(topic, fn)
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
