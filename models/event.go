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
func NewEvent(name string, payload M) Event {
	if payload == nil {
		payload = make(map[string]interface{})
	}

	return Event{
		name: name,
		payload: payload,
	}
}

// Abort abort event loop exec
func (event *Event) Abort(abort bool) {
	event.aborted = abort
}

// Fill event data
func (event *Event) Fill(target interface{}, payload M) *Event {
	if payload != nil {
		event.payload = payload
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
	if v, ok := event.payload[key]; ok {
		return v
	}

	return nil
}

// Add value by key
func (event *Event) Add(key string, val interface{}) {
	if _, ok := event.payload[key]; !ok {
		event.Set(key, val)
	}
}

// Set value by key
func (event *Event) Set(key string, val interface{}) {
	if event.payload == nil {
		event.payload = make(map[string]interface{})
	}

	event.payload[key] = val
}

// Name get event name
func (event *Event) Name() string {
	return event.name
}

// Data get all data
func (event *Event) Data() map[string]interface{} {
	return event.payload
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
	event.name = name
	return event
}

func (event *Event) SetPayload(payload M) Event {
	if payload != nil {
		event.payload = payload
	}
	return *event
}

// SetTarget set event target
func (event *Event) SetTarget(target interface{}) *Event {
	event.target = target
	return event
}
