package gui

import (
	"fmt"
	"net"
	"time"
)

type Server struct {
	context *Context
}

func (inst *Server) Start() error {

	var err error
	stat := inst.context.GetStatus()
	ctx := inst.context

	stat.starting = true

	go func() {
		err = inst.run()
		ctx.HandleError(err)
	}()

	return nil
}

func (inst *Server) Stop() error {
	stat := inst.context.GetStatus()
	stat.stopping = true
	return nil
}

func (inst *Server) innerUpdateStatus() {

}

func (inst *Server) run() error {

	ctx := inst.context
	stat := ctx.GetStatus()
	logger := ctx.logger

	defer func() {
		stat.stopped = true
		inst.innerUpdateStatus()
	}()

	// open

	const network = "tcp"
	host := inst.context.serverHost
	port := inst.context.serverPort
	addr1 := fmt.Sprintf("%s:%d", host, port)

	addr2, err := net.ResolveTCPAddr(network, addr1)
	if err != nil {
		return err
	}

	li, err := net.ListenTCP(network, addr2)
	if err != nil {
		return err
	}

	defer li.Close()

	// started

	stat.started = true
	inst.innerUpdateStatus()

	logger.Info("listen TCP @%s:%d", host, port)

	// accept

	for {

		if stat.stopping || stat.stopped {
			break
		}

		session := new(Session)
		conn, err := li.AcceptTCP()
		if err != nil {
			ctx.HandleError(err)
			time.Sleep(time.Second * 3)
			continue
		}

		// session.context = ctx
		// session.conn = conn
		session.Init(ctx, conn)

		err = inst.innerHandleSession(session)
		if err != nil {
			ctx.HandleError(err)
			time.Sleep(time.Second * 3)
			continue
		}

	}

	return nil
}

func (inst *Server) innerHandleSession(se *Session) error {

	ctx := inst.context
	inst.innerSetCurrentSession(se)

	go func() {
		err := se.Run()
		ctx.HandleError(err)
	}()

	return nil
}

func (inst *Server) innerSetCurrentSession(next *Session) {

	ctx := inst.context
	prev := ctx.session

	ctx.session = next
	if prev == nil {
		return
	}

	err := prev.Close()
	ctx.HandleError(err)
}
