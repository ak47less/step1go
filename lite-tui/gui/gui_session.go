package gui

import (
	"fmt"
	"io"
	"net"
)

type Session struct {
	context *Context
	conn    *net.TCPConn
	reader  io.Reader
	writer  io.Writer
}

func (inst *Session) Init(ctx *Context, conn *net.TCPConn) error {

	if ctx == nil || conn == nil {
		return fmt.Errorf("Session.Init() : param(s) is nil")
	}

	inst.context = ctx
	inst.conn = conn
	inst.reader = conn
	inst.writer = conn
	return nil
}

func (inst *Session) SendCommand(cmd string) error {
	b := []byte(cmd)
	_, err := inst.writer.Write(b)

	logger := inst.context.logger
	logger.Info("send_command: %s", cmd)

	return err
}

func (inst *Session) Close() error {
	conn := inst.conn
	if conn == nil {
		return nil
	}
	return conn.Close()
}

func (inst *Session) Run() error {

	var buffer [16]byte

	ctx := inst.context
	rdr := inst.reader
	stat := ctx.GetStatus()
	logger := ctx.logger

	for {

		if stat.stopped || stat.stopping {
			break
		}

		cnt, err := rdr.Read(buffer[:])
		if err != nil {
			return err
		}

		logger.Debug("rx %d byte(s)", cnt)
	}

	return nil
}
