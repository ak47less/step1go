package lite

import "fmt"

////////////////////////////////////////////////////////////////////////////////

type innerInputInstrument struct {
	context *Context
}

// Dispatch implements [Instrument].
func (inst *innerInputInstrument) Dispatch(msg *MidiMessage) error {

	if inst == nil {
		return fmt.Errorf("innerInputInstrument: this is nil")
	}
	if msg == nil {
		return fmt.Errorf("innerInputInstrument: param 'msg' (MidiMessage) is nil")
	}

	ctx := inst.context
	target, err := ctx.CurrentInstrumentHolder.GetInstrument(ctx)

	if err != nil {
		return err
	}

	return target.Dispatch(msg)
}

// GetInfo implements [Instrument].
func (inst *innerInputInstrument) GetInfo() *InstrumentInfo {
	return &InstrumentInfo{
		Name:   "InputInstrument",
		Label:  "Input-Instrument",
		Preset: "default",
	}
}

func (inst *innerInputInstrument) _impl() Instrument {
	return inst
}

////////////////////////////////////////////////////////////////////////////////

type AutoInputInstrumentHolder struct {
}

// GetInstrument implements [InstrumentHolder].
func (inst *AutoInputInstrumentHolder) GetInstrument(ctx *Context) (Instrument, error) {

	if ctx == nil {
		return nil, fmt.Errorf("param: engine.context is nil")
	}

	ii := ctx.InputInstrument
	if ii == nil {
		// load
		ii2, err := inst.loadII(ctx)
		if err != nil {
			return nil, err
		}
		ctx.InputInstrument = ii2
		ii = ii2
	}

	return ii, nil
}

func (inst *AutoInputInstrumentHolder) loadII(ctx *Context) (Instrument, error) {
	if ctx == nil {
		return nil, fmt.Errorf("param: engine.context is nil")
	}
	ii := &innerInputInstrument{
		context: ctx,
	}
	return ii, nil
}

func (inst *AutoInputInstrumentHolder) _impl() InstrumentHolder {
	return inst
}

////////////////////////////////////////////////////////////////////////////////
// EOF
