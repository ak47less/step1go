package lite

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/ak47less/step1go/common"
)

type TcpController struct {
	context *EngineContext
	monitor *Monitor
	keymap  *common.ComputerKeyMap

	inner ActionController
}

func (inst *TcpController) init() {

	ec := inst.context
	ctrl := NewActionController(ec)
	kmb := new(common.ComputerKeyMapBuilder)

	inst.initKeyMap(kmb)

	inst.keymap = kmb.Build()
	inst.inner = ctrl
}

func (inst *TcpController) initKeyMap(b *common.ComputerKeyMapBuilder) {

	// transport

	b.MapKeyName("p").ToControl(common.ActionPlay)
	b.MapKeyName("s").ToControl(common.ActionStop)
	b.MapKeyName("r").ToControl(common.ActionRecord)
	b.MapKeyName("l").ToControl(common.ActionLoop)

	b.MapKeyCode(common.CKeyCodeSpace).ToControl(common.ActionResume)
	b.MapKeyCode(common.CKeyCodeSpace).ToControl(common.ActionPause)

	// 12-tones

	b.MapKeyCode(common.CKeyCode1).ToNote(60)
	b.MapKeyCode(common.CKeyCode2).ToNote(61)
	b.MapKeyCode(common.CKeyCode3).ToNote(62)
	b.MapKeyCode(common.CKeyCode4).ToNote(63)
	b.MapKeyCode(common.CKeyCode5).ToNote(64)
	b.MapKeyCode(common.CKeyCode6).ToNote(65)
	b.MapKeyCode(common.CKeyCode7).ToNote(66)
	b.MapKeyCode(common.CKeyCode8).ToNote(67)
	b.MapKeyCode(common.CKeyCode9).ToNote(68)
	b.MapKeyCode(common.CKeyCode0).ToNote(69)
	b.MapKeyName(common.CKeyNameMinus).ToNote(70)
	b.MapKeyName(common.CKeyNameEqual).ToNote(71)

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

func (inst *TcpController) innerParseKeyCommand(cmd string) ([]*common.ComputerKeyMapping, error) {
	const (
		ch1 = ' '
		ch2 = '\t'
		sep = "\n"
	)
	cmd = strings.ReplaceAll(cmd, string(ch1), sep)
	cmd = strings.ReplaceAll(cmd, string(ch2), sep)

	words := strings.Split(cmd, sep)
	mp1 := new(common.ComputerKeyMapping)
	result := make([]*common.ComputerKeyMapping, 0)
	index := 0

	for _, w := range words {
		if len(w) == 0 {
			continue
		}
		if index == 0 {
			inst.innerInitMapping(mp1, w)
		} else {
			inst.innerUpdateMapping(mp1, w)
			mp2 := inst.innerCloneMapping(mp1)
			result = append(result, mp2)
		}
		index++
	}

	return result, nil
}

func (inst *TcpController) innerInitMapping(mp *common.ComputerKeyMapping, word string) {

	if inst == nil || mp == nil {
		return
	}

	mp.FlagKeyUp = false
	mp.FlagKeyDown = false

	switch word {

	// key-down

	case "kd":
		mp.FlagKeyDown = true
	case "down":
		mp.FlagKeyDown = true

		// key-up

	case "ku":
		mp.FlagKeyUp = true
	case "up":
		mp.FlagKeyUp = true
	}
}

func (inst *TcpController) innerUpdateMapping(mp *common.ComputerKeyMapping, word string) {

	name := common.ComputerKeyName(word)
	mp.Name = name
	mp.Code = name.Code()
}

func (inst *TcpController) innerCloneMapping(src *common.ComputerKeyMapping) *common.ComputerKeyMapping {
	if inst == nil || src == nil {
		return nil
	}
	dst := new(common.ComputerKeyMapping)
	*dst = *src
	return dst
}

func (inst *TcpController) handleKeyCommand(cmd string) error {

	maplist, err := inst.innerParseKeyCommand(cmd)
	if err != nil {
		return err
	}

	for _, km := range maplist {
		err = inst.handleKeyMapping(km)
		if err != nil {
			return err
		}
	}

	return nil
}

func (inst *TcpController) handleKeyMapping(want *common.ComputerKeyMapping) error {

	if inst == nil || want == nil {
		return nil
	}

	have, err := inst.keymap.Find(want)
	if err != nil {
		return err
	}

	return inst.inner.Execute(have.Action)
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
