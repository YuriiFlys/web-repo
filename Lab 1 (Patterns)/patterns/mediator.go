package patterns

import (
	"fmt"
	"time"
)

type EventType string

const (
	EventChorus   EventType = "CHORUS"
	EventDrop     EventType = "DROP"
	EventTalk     EventType = "TALK"
	EventSmokeNow EventType = "SMOKE_NOW"
)

type Event struct {
	Type    EventType
	From    string
	Payload string
	At      time.Time
}

type StageHub interface {
	Register(Node)
	Emit(Event)
	Send(to string, e Event)
}

type Node interface {
	Name() string
	Receive(e Event)
	Bind(h StageHub)
}

type ShowHub struct {
	nodes map[string]Node
}

func NewShowHub() *ShowHub {
	return &ShowHub{nodes: make(map[string]Node)}
}

func (h *ShowHub) Register(n Node) {
	h.nodes[n.Name()] = n
	n.Bind(h)
}

func (h *ShowHub) Emit(e Event) {
	e.At = time.Now()
	for name, n := range h.nodes {
		if name == e.From {
			continue
		}
		n.Receive(e)
	}
}

func (h *ShowHub) Send(to string, e Event) {
	e.At = time.Now()
	n, ok := h.nodes[to]
	if !ok {
		fmt.Printf("[HUB] node %q not found\n", to)
		return
	}
	n.Receive(e)
}

type Artist struct {
	name string
	hub  StageHub
}

func NewArtist(name string) *Artist { return &Artist{name: name} }
func (a *Artist) Name() string      { return a.name }
func (a *Artist) Bind(h StageHub)   { a.hub = h }
func (a *Artist) Receive(e Event) {
	fmt.Printf("[Artist:%s] received %s from %s (%s)\n", a.name, e.Type, e.From, e.Payload)
}

func (a *Artist) Trigger(t EventType, payload string) {
	a.hub.Emit(Event{Type: t, From: a.name, Payload: payload})
}

type LightingRig struct {
	name string
	hub  StageHub
}

func NewLightingRig(name string) *LightingRig { return &LightingRig{name: name} }
func (l *LightingRig) Name() string           { return l.name }
func (l *LightingRig) Bind(h StageHub)        { l.hub = h }
func (l *LightingRig) Receive(e Event) {
	switch e.Type {
	case EventChorus:
		fmt.Printf("[Lights:%s] switching to preset: %s\n", l.name, "WIDE-BRIGHT")
	case EventDrop:
		fmt.Printf("[Lights:%s] strobe ON!\n", l.name)
	case EventTalk:
		fmt.Printf("[Lights:%s] warm spotlight\n", l.name)
	}
}

type SmokeMachine struct {
	name string
	hub  StageHub
}

func NewSmokeMachine(name string) *SmokeMachine { return &SmokeMachine{name: name} }
func (s *SmokeMachine) Name() string            { return s.name }
func (s *SmokeMachine) Bind(h StageHub)         { s.hub = h }
func (s *SmokeMachine) Receive(e Event) {
	if e.Type == EventSmokeNow {
		fmt.Printf("[Smoke:%s] pumping smoke: %s\n", s.name, e.Payload)
	}
}

type SoundRack struct {
	name string
	hub  StageHub
}

func NewSoundRack(name string) *SoundRack { return &SoundRack{name: name} }
func (s *SoundRack) Name() string         { return s.name }
func (s *SoundRack) Bind(h StageHub)      { s.hub = h }
func (s *SoundRack) Receive(e Event) {
	if e.Type == EventDrop {
		fmt.Printf("[Sound:%s] sub-bass boost!\n", s.name)
	}
}
