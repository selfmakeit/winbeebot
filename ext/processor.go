package ext

import (
	"winbeebot"
)

// Processor is used to provide an entry point for wrapping internal Dispatcher logic, such that the basic logic can be
// customised and extended.

// type Processor interface {
// 	ProcessUpdate(d *Dispatcher, b *winbeebot.Bot, ctx *Context) error
// }

// var _ Processor = BaseProcessor{}
//上面的这种方式更方便
type Processor interface {
	ProcessUpdate(d *Dispatcher, b *winbeebot.Bot, update  *winbeebot.Update) error
}
// BaseProcessor is the simplest version of the Processor; it simply calls the dispatcher straight away.
type BaseProcessor struct{}

// ProcessUpdate iterates over the list of groups to execute the matching handlers.
func (bp BaseProcessor) ProcessUpdate(d *Dispatcher, b *winbeebot.Bot, ctx *Context) error {
	return d.iterateOverHandlerGroups(b, ctx)
}
