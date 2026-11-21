package core

import "github.com/ak47less/step1go/common"

var theErrorHandler innerErrorHandler

////////////////////////////////////////////////////////////////////////////////

func handleError(err error) {
	theErrorHandler.handle(err)
}

func HandleError(err error) {
	theErrorHandler.handle(err)
}

////////////////////////////////////////////////////////////////////////////////

type innerErrorHandler struct {
	logger Logger
}

func (inst *innerErrorHandler) handle(err error) {

	if err == nil {
		return
	}

	if inst == nil {
		return
	}

	inst.innerGetLogger().Error("Error: %s", err.Error())

}

func (inst *innerErrorHandler) innerGetLogger() Logger {
	l := inst.logger
	if l == nil {
		l = common.Log()
		inst.logger = l
	}
	return l
}

////////////////////////////////////////////////////////////////////////////////
