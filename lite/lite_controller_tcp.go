package lite

import (
	"fmt"
	"net"
	"time"
)

type TcpController struct {
	context   *EngineContext
	monitor   *Monitor
	keymap    map[rune]Action
	actionCtl ActionController
}

func (inst *TcpController) init() {

	ec := inst.context
	ctrl := NewActionController(ec)
	km := make(map[rune]Action)

	inst.initKeyMap(km)

	inst.keymap = km
	inst.actionCtl = ctrl
}

func (inst *TcpController) initKeyMap(km map[rune]Action) {

	km['p'] = ActionPlay
	km['s'] = ActionStop
	km['r'] = ActionRecord
	km['-'] = ActionResume
	km['='] = ActionPause

}

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

func (inst *TcpController) handleKeyInput(keycode byte) error {
	act, err := inst.innerGetActionByCode(keycode)
	if err != nil {
		return err
	}
	return inst.actionCtl.Execute(act)
}

func (inst *TcpController) innerGetActionByCode(keycode byte) (Action, error) {

	table := inst.keymap
	ch := rune(keycode)

	if 'A' <= ch && ch <= 'Z' {
		ch = ch - 'A' + 'a'
	}

	if table != nil {
		act := table[ch]
		if act > 0 {
			return act, nil
		}
	}

	return 0, fmt.Errorf("no item [%s] in key-map", string(ch))
}

func (inst *TcpController) innerRunReadLoop(conn net.Conn) error {

	var buffer [1]byte

	for {
		if inst.innerHasStop() {
			break
		}
		cnt, err := conn.Read(buffer[:])
		if err == nil && cnt == 1 {
			code := buffer[0]
			err = inst.handleKeyInput(code)
			handleError(err)
		} else {
			time.Sleep(time.Second)
		}
	}

	return nil
}
