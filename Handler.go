package ranni

type EventHandler interface {
	Do(ctx *EventContext)
	Filter(ctx *EventContext) bool
	Help() string
}
