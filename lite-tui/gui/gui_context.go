package gui

import (
	"fmt"
	"io"

	"github.com/ak47less/step1go/common"
)

type Context struct {
	server  *Server
	session *Session
	status  *Status

	serverProtocol string
	serverHost     string
	serverPort     int

	logger common.Logger

	clist []io.Closer
}

func (inst *Context) Init() error {

	inst.serverProtocol = "tcp"
	inst.serverHost = "127.0.0.1"
	inst.serverPort = 33309

	inst.status = new(Status)
	inst.logger = common.Log()

	inst.GetServer()
	inst.logger.SetTag("lite-gui")

	return nil
}

func (inst *Context) GetServer() *Server {
	ser := inst.server
	if ser == nil {
		ser = new(Server)
		ser.context = inst
		inst.server = ser
	}
	return ser
}

func (inst *Context) GetSession() (*Session, error) {
	se := inst.session
	if se == nil {
		return nil, fmt.Errorf("no active session")
	}
	return se, nil
}

func (inst *Context) GetStatus() *Status {

	st := inst.status
	if st == nil {
		st = new(Status)
		inst.status = st
	}

	st.host = inst.serverHost
	st.port = inst.serverPort

	return st
}

func (inst *Context) HandleError(err error) {
	if err == nil {
		return
	}
	logger := inst.logger
	logger.Error(err.Error())
}

func (inst *Context) GetLogger() common.Logger {
	return inst.logger
}
