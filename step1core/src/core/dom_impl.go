package core

import "github.com/ak47less/step1go/step1core"

func GetDomImplementation() step1core.DOMImplementation {
	return new(innerDOMImplementation)
}

type innerDOMImplementation struct {
}

// Codec implements [step1core.DOMImplementation].
func (i *innerDOMImplementation) Codec() step1core.DOMDocumentCodec {
	panic("unimplemented")
}

// Factory implements [step1core.DOMImplementation].
func (i *innerDOMImplementation) Factory() step1core.DOMFactory {
	panic("unimplemented")
}
