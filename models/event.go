package models

import (
	"fmt"
)

type EventObject interface {
	//emit(data interface{})
	GetId() uint
	//addListener(listener Listener)
}

/*type Listener struct {
	// Channel for receiving notifications from the database.  In some cases a
	// nil value will be sent.  See section "Notifications" above.
	Notify chan *Notification

	name                 string
	minReconnectInterval time.Duration
	maxReconnectInterval time.Duration
	dialer               Dialer
	eventCallback        EventCallbackType

	lock                 sync.Mutex
	isClosed             bool
	reconnectCond        *sync.Cond
	cn                   *ListenerConn
	connNotificationChan <-chan *Notification
	channels             map[string]struct{}
}*/


type ProductCreate struct {
	AccountId uint
	ProductId uint
}
func (ProductCreate) emit(product Product) {}

type ListenerEventType = int

const (
	// ListenerEventConnected is emitted only when the database connection
	// has been initially initialized. The err argument of the callback
	// will always be nil.
	ListenerEventConnected ListenerEventType = iota

	// ListenerEventDisconnected is emitted after a database connection has
	// been lost, either because of an error or because Close has been
	// called. The err argument will be set to the reason the database
	// connection was lost.
	ListenerEventDisconnected

	// ListenerEventReconnected is emitted after a database connection has
	// been re-established after connection loss. The err argument of the
	// callback will always be nil. After this event has been emitted, a
	// nil pq.Notification is sent on the Listener.Notify channel.
	ListenerEventReconnected

	// ListenerEventConnectionAttemptFailed is emitted after a connection
	// to the database was attempted, but failed. The err argument will be
	// set to an error describing why the connection attempt did not
	// succeed.
	ListenerEventConnectionAttemptFailed
)

func emit(event ListenerEventType, data interface{}) {
	fmt.Println("name event: ", event)
}
