package core

import "errors"

type PropertyGetter interface {
	Error() error

	Properties() Properties

	Optional() PropertyGetter

	Required() PropertyGetter

	// getter for types

	GetInt(name string, default_values ...int) int

	GetUint(name string, default_values ...uint) uint

	GetString(name string, default_values ...string) string

	GetByte(name string, default_values ...byte) byte

	GetBool(name string, default_values ...bool) bool
}

//////////////////////////////////////////////////////////////////////////////

type innerPropertyGetter struct {
	err      error
	required bool
	table    Properties
}

// GetBool implements [PropertyGetter].
func (inst *innerPropertyGetter) GetBool(name string, default_values ...bool) bool {
	panic("unimplemented")
}

// GetByte implements [PropertyGetter].
func (inst *innerPropertyGetter) GetByte(name string, default_values ...byte) byte {
	panic("unimplemented")
}

// GetInt implements [PropertyGetter].
func (inst *innerPropertyGetter) GetInt(name string, default_values ...int) int {
	panic("unimplemented")
}

// GetString implements [PropertyGetter].
func (inst *innerPropertyGetter) GetString(name string, default_values ...string) string {
	panic("unimplemented")
}

// GetUint implements [PropertyGetter].
func (inst *innerPropertyGetter) GetUint(name string, default_values ...uint) uint {
	panic("unimplemented")
}

// Properties implements [PropertyGetter].
func (inst *innerPropertyGetter) Properties() Properties {
	return inst.table
}

// Error implements [PropertyGetter].
func (inst *innerPropertyGetter) Error() error {
	return inst.err
}

// Optional implements [PropertyGetter].
func (inst *innerPropertyGetter) Optional() PropertyGetter {
	inst.required = false
	return inst
}

// Required implements [PropertyGetter].
func (inst *innerPropertyGetter) Required() PropertyGetter {
	inst.required = true
	return inst
}

func (inst *innerPropertyGetter) init(table Properties) PropertyGetter {

	if inst == nil {
		inst = new(innerPropertyGetter)
	}

	if table == nil {
		table = make(Properties)
	}

	inst.table = table
	inst.err = nil
	inst.required = true

	return inst
}

func (inst *innerPropertyGetter) innerAddError(err error) {
	if inst == nil || err == nil {
		return
	}
	older := inst.err
	if older == nil {
		inst.err = err
	} else {
		inst.err = errors.Join(older, err)
	}
}

func (inst *innerPropertyGetter) _impl() PropertyGetter {
	return inst
}

//////////////////////////////////////////////////////////////////////////////
// EOF
