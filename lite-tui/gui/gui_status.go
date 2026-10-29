package gui

import (
	"fmt"
	"strings"
)

type Status struct {
	host string
	port int

	starting bool
	stopping bool
	started  bool
	stopped  bool

	PressKeyCode  byte
	PressKeyCount uint
}

func (inst *Status) innerIsRunning() bool {
	return (inst.starting && inst.started && (!inst.stopped))
}

func (inst *Status) String() string {

	const nl = "\n"
	b := new(strings.Builder)

	if inst.innerIsRunning() {
		b.WriteString("gui.server: running ")
	} else {
		b.WriteString("gui.server: stopped")
	}
	b.WriteString(nl)

	b.WriteString(fmt.Sprintf("[listen @%s:%d]"+nl, inst.host, inst.port))

	charPK := rune(inst.PressKeyCode)
	b.WriteString(fmt.Sprintf("[key_press code:%d count:%d  char:'%s' ]"+nl, charPK, inst.PressKeyCount, string(charPK)))

	return b.String()
}
