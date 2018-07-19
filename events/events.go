package events

import (
	"github.com/asaskevich/EventBus"
)

var Bus EventBus.Bus

func init() {
	Bus = EventBus.New()
}

//func PushFloData(floData string, tx datastore.TransactionData) {
//	Bus.Publish("flo:floData", floData, tx)
//}

//import "sync"
//
//var rwm sync.RWMutex
//var subs = make(map[string][]chan interface{})
//
//func Subscribe(e string, c chan interface{}) {
//	rwm.Lock()
//	subs[e] = append(subs[e], c)
//	rwm.Unlock()
//}
//
//func Unsubscribe(e string, c chan interface{}) {
//	rwm.Lock()
//	chans, ok := subs[e]
//	if !ok {
//		rwm.Unlock()
//		return
//	}
//	var newChans []chan interface{}
//	for _, v := range chans {
//		if v == c {
//			close(c)
//		} else {
//			newChans = append(newChans, v)
//		}
//	}
//	subs[e] = newChans
//	rwm.Unlock()
//}
//
//func Notify(e string, data interface{}) {
//	rwm.RLock()
//	chans, ok := subs[e]
//	if !ok {
//		rwm.RUnlock()
//		return
//	}
//	for _, c := range chans {
//		c <- data
//	}
//	rwm.RUnlock()
//}
