package ext

import "winbeebot"

type Handler interface {
	// CheckUpdate checks whether the update should handled by this handler.
	CheckUpdate(b *winbeebot.Bot, ctx *Context) bool
	// HandleUpdate processes the update.
	HandleUpdate(b *winbeebot.Bot, ctx *Context) error
	// Name gets the handler name; used to differentiate handlers programmatically. Names should be unique.
	Name() string
}
