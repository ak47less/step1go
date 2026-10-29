package lite

import (
	"fmt"
	"net"
	"strings"
	"time"
)

type TcpController struct {
	context *EngineContext
	monitor *Monitor

	actionCtl ActionController

	keymap map[KeyCode]*ActionHolder
}

func (inst *TcpController) init() {

	ec := inst.context
	ctrl := NewActionController(ec)
	km := make(map[KeyCode]*ActionHolder)

	inst.initKeyMap(km)

	inst.keymap = km
	inst.actionCtl = ctrl
}

func (inst *TcpController) initKeyMap(km map[KeyCode]*ActionHolder) {

	km['p'] = ActionPlay
	km['s'] = ActionStop
	km['r'] = ActionRecord
	km['-'] = ActionResume
	km['='] = ActionPause

}

func (inst *TcpController) initKeyMapToControl(km map[KeyCode]*ActionHolder) {}

func (inst *TcpController) initKeyMapToNote(km map[KeyCode]*ActionHolder) {}

func (inst *TcpController) Start() {

	inst.init()

	go func() {
		err := inst.run()
		handleError(err)
	}()

}

func (inst *TcpController) stop() {
	inst.context.Engine.Stop()
}

func (inst *TcpController) run() error {

	cfg := inst.context.Configuration.Monitor
	network := cfg.Scheme
	host := cfg.Host
	port := cfg.Port

	addr1 := fmt.Sprintf("%s:%d", host, port)

	defer inst.stop()

	conn, err := net.Dial(network, addr1)
	if err != nil {
		return err
	}

	defer conn.Close()

	return inst.innerRunReadLoop(conn)
}

func (inst *TcpController) innerHasStop() bool {
	stat := &inst.context.State
	return ((!stat.Starting) && stat.Stopped && stat.Stopping)
}

func (inst *TcpController) handleKeyCommand(cmd string) error {

	kalist, err := inst.innerParseKeyActionCommand(cmd)
	if err != nil {
		return err
	}

	for _, ka := range kalist {
		err = inst.handleKeyAction(ka)
		if err != nil {
			return err
		}
	}
	return nil
}

func (inst *TcpController) innerParseActionX(src *KeyAction, dst *ActionHolder) error {

	if inst == nil || src == nil || dst == nil {
		return fmt.Errorf("param(s) is nil")
	}

	action := src.Action
	code := src.KeyCode
	ah := inst.keymap[code]

	if ah == nil {
		return fmt.Errorf("no action info register for keycode: %d", code)
	}

	if ah.NoteAction != nil {

		// as a note-action
		na2 := new(NoteAction)
		*na2 = *ah.NoteAction

		if src.KeyDown {
			action = ActionNoteOn
		}
		if src.KeyUp {
			action = ActionNoteOff
		}

		na2.Action = action
		dst.Action = action
		dst.NoteAction = na2
	}

	dst.Action = action
	dst.KeyAction = src

	return nil
}

func (inst *TcpController) handleKeyAction(a1 *KeyAction) error {

	ah := new(ActionHolder)
	err := inst.innerParseActionX(a1, ah)
	if err != nil {
		return err
	}

	return inst.actionCtl.Execute(ah)
}

func (inst *TcpController) innerParseKeyCode(ka *KeyAction) {
	name := ka.KeyName
	if len(name) == 1 {
		code := name[0]
		ka.KeyCode = KeyCode(code)
		return
	}

	// find in map:

}

func (inst *TcpController) innerParseKeyActionCommand(cmd string) ([]*KeyAction, error) {
	const (
		s1  = '\t'
		s2  = ' '
		sep = "\n"
	)
	cmd = strings.ReplaceAll(cmd, string(s1), sep)
	cmd = strings.ReplaceAll(cmd, string(s2), sep)
	words := strings.Split(cmd, sep)
	index := 0
	k0 := new(KeyAction) // the prototype
	results := make([]*KeyAction, 0)

	for _, word := range words {
		if word == "" {
			continue
		}
		if index < 1 {
			switch word {
			case "kd":
				k0.Action = ActionKeyDown
				k0.KeyDown = true
			case "ku":
				k0.Action = ActionKeyUp
				k0.KeyUp = true
			default:
				return nil, fmt.Errorf("bad key-action-command: '%s'", word)
			}
		} else {
			k2 := k0.Clone()
			k2.KeyName = word
			inst.innerParseKeyCode(k2)
			results = append(results, k2)
		}
		index++
	}

	return results, nil
}

func (inst *TcpController) innerRunReadLoop(conn net.Conn) error {

	const (
		nl         = '\n'
		cr         = '\r'
		maxlinelen = 64
	)

	buffer := new(strings.Builder)
	var buf [1]byte

	for {
		if inst.innerHasStop() {
			break
		}

		cnt, err := conn.Read(buf[:])

		if err == nil && cnt == 1 {
			b := buf[0]
			if b == nl || b == cr {
				cmd := buffer.String()
				err = inst.handleKeyCommand(cmd)
				buffer.Reset()
			} else if buffer.Len() < maxlinelen {
				buffer.WriteByte(b)
			}
		}

		if err != nil {
			handleError(err)
			time.Sleep(time.Second)
			err = nil
		}
	}

	return nil
}
