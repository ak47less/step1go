package step1core

////////////////////////////////////////////////////////////////////////////////

type PropertyType string

const (
	PropertyTypeFloat  PropertyType = "float"
	PropertyTypeInt    PropertyType = "int"
	PropertyTypeUint   PropertyType = "uint"
	PropertyTypeString PropertyType = "string"
)

////////////////////////////////////////////////////////////////////////////////

type PropertyGetter interface {
	GetText(def string) string
}

type PropertySetter interface {
	SetText(value string) error
}

type PropertyGetSetter interface {
	PropertyGetter
	PropertySetter
}

type Property struct {
	Type PropertyType

	Name string

	Label string

	Description string

	Min    PropertyGetter
	Max    PropertyGetter
	Getter PropertyGetter

	Setter PropertySetter
}

////////////////////////////////////////////////////////////////////////////////

type IntGetSetter interface {
	PropertyGetSetter

	GetInt(def int) int

	SetInt(value int) error
}

type FloatGetSetter interface {
	PropertyGetSetter

	GetFloat(def float32) float32

	SetFloat(value float32) error
}

////////////////////////////////////////////////////////////////////////////////
// EOF
