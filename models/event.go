package models

// Priority
const (
	Min         = -300
	Low         = -200
	BelowNormal = -100
	Normal      = 0
	AboveNormal = 100
	High        = 200
	Max         = 300
)

// NewEvent new an basic event instance
func NewEvent(name string, data M) Event {
	if data == nil {
		data = make(map[string]interface{})
	}

	return Event{
		Name: name,
		data: data,
	}
}

// Abort abort event loop exec
func (event *Event) Abort(abort bool) {
	event.aborted = abort
}

// Fill event data
func (event *Event) Fill(target interface{}, data M) *Event {
	if data != nil {
		event.data = data
	}

	event.target = target
	return event
}

// AttachTo add current event to the event manager.
func (event *Event) AttachTo(em ManagerFace) {
	em.AddEvent(*event)
}

// Get get data by index
func (event *Event) Get(key string) interface{} {
	if v, ok := event.data[key]; ok {
		return v
	}

	return nil
}

// Add value by key
func (event *Event) Add(key string, val interface{}) {
	if _, ok := event.data[key]; !ok {
		event.Set(key, val)
	}
}

// Set value by key
func (event *Event) Set(key string, val interface{}) {
	if event.data == nil {
		event.data = make(map[string]interface{})
	}

	event.data[key] = val
}

// Name get event name
func (event *Event) GetName() string {
	return event.Name
}

// Data get all data
func (event *Event) Data() map[string]interface{} {
	return event.data
}

// IsAborted check.
func (event *Event) IsAborted() bool {
	return event.aborted
}

// Target get target
func (event *Event) Target() interface{} {
	return event.target
}

// SetName set event name
func (event *Event) SetName(name string) *Event {
	event.Name = name
	return event
}

func (event *Event) SetData(data M) *Event {
	if data != nil {
		event.data = data
	}
	return event
}

func (event *Event) SetPayload(payload M) *Event {
	event.Set("payload", payload)
	return event
}

func (event *Event) SetRecipientList(recipientList 	[]uint) *Event {
	if recipientList != nil {
		event.recipientList = recipientList
	}
	return event
}

// SetTarget set event target
func (event *Event) SetTarget(target interface{}) *Event {
	event.target = target
	return event
}
