package core

import "fmt"

type Properties map[string]string

func (p Properties) Init() Properties {
	if p == nil {
		p = make(Properties)
	}
	return p
}

func (p Properties) GetValue(name string) string {
	if p == nil {
		return ""
	}
	return p[name]
}

func (p Properties) SetValue(name, value string) Properties {
	if p == nil {
		p = make(Properties)
	}
	p[name] = value
	return p
}

func (p Properties) GetValueOptional(name string, def string) string {
	value := p.GetValue(name)
	if value == "" {
		return def
	}
	return value
}

func (p Properties) GetValueRequired(name string) (string, error) {
	value := p.GetValue(name)
	if value == "" {
		return "", fmt.Errorf("no property with name '%s'", name)
	}
	return value, nil
}
