package common

import "testing"

func TestLogger(t *testing.T) {

	logger := Log()
	vbool := false
	vfloat := 3.141592
	vint := 12345
	vstr := "foo,bar"

	logger.SetGate(LogLevelMin)
	logger.SetTag("test4log")

	logger.Fatal("__fatal")
	logger.Error("__error")
	logger.Warn("__warn")
	logger.Info("__info  [format_values v_bool:%v v_float:%f v_int:%d v_str:'%s']", vbool, vfloat, vint, vstr)
	logger.Debug("__debug")
	logger.Trace("__trace")

}
