package core

import (
	"errors"
	"fmt"
	"strconv"
)

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

	GetFloat32(name string, default_values ...float32) float32

	GetFloat64(name string, default_values ...float64) float64
}

//////////////////////////////////////////////////////////////////////////////

type innerPropertyGetter struct {
	err      error
	required bool
	table    Properties
}

// GetFloat32 implements [PropertyGetter].
func (inst *innerPropertyGetter) GetFloat32(name string, default_values ...float32) float32 {
	panic("unimplemented")
}

// GetFloat64 implements [PropertyGetter].
func (inst *innerPropertyGetter) GetFloat64(name string, default_values ...float64) float64 {
	panic("unimplemented")
}

// GetBool implements [PropertyGetter].
func (inst *innerPropertyGetter) GetBool(name string, default_values ...bool) bool {
	str, ok := inst.innerGetValue(name)
	if ok {
		value, err := strconv.ParseBool(str)
		inst.innerAddParseError(name, err)
		if err == nil {
			return value
		}
	}
	for _, value := range default_values {
		return value
	}
	return false
}

// GetByte implements [PropertyGetter].
func (inst *innerPropertyGetter) GetByte(name string, default_values ...byte) byte {
	str, ok := inst.innerGetValue(name)
	if ok {
		value, err := strconv.ParseInt(str, 10, 8)
		inst.innerAddParseError(name, err)
		if err == nil {
			return byte(value)
		}
	}
	for _, value := range default_values {
		return value
	}
	return 0
}

// GetInt implements [PropertyGetter].
func (inst *innerPropertyGetter) GetInt(name string, default_values ...int) int {
	str, ok := inst.innerGetValue(name)
	if ok {
		value, err := strconv.ParseInt(str, 10, 0)
		inst.innerAddParseError(name, err)
		if err == nil {
			return int(value)
		}
	}
	for _, value := range default_values {
		return value
	}
	return 0
}

// GetString implements [PropertyGetter].
func (inst *innerPropertyGetter) GetString(name string, default_values ...string) string {
	str, ok := inst.innerGetValue(name)
	if ok {
		return str
	}
	for _, value := range default_values {
		return value
	}
	return ""
}

// GetUint implements [PropertyGetter].
func (inst *innerPropertyGetter) GetUint(name string, default_values ...uint) uint {
	str, ok := inst.innerGetValue(name)
	if ok {
		value, err := strconv.ParseUint(str, 10, 0)
		inst.innerAddParseError(name, err)
		if err == nil {
			return uint(value)
		}
	}
	for _, value := range default_values {
		return value
	}
	return 0
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

func (inst *innerPropertyGetter) innerGetValue(name string) (string, bool) {
	value, err := inst.table.GetValueRequired(name)
	if err != nil {
		if inst.required {
			inst.innerAddError(err)
		}
		return value, false
	}
	return value, true
}

func (inst *innerPropertyGetter) innerAddParseError(name string, err error) {
	if err == nil {
		return
	}
	inst.innerAddError(fmt.Errorf("parse property error, name:%s error:%w", name, err))
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
