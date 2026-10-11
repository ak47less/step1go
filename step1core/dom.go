package step1core

type DOM struct {
	Document DOMDocument

	Factory DOMFactory

	Codec DOMDocumentCodec

	Implementation DOMImplementation
}

type DOMImplementation interface {
	Codec() DOMDocumentCodec

	Factory() DOMFactory
}

type DOMNode interface {
	GetOwner() DOMDocument
}

type DOMFactory interface {
	CreateDocument(url string) DOMDocument

	CreateElement(doc DOMDocument, namespace string, prefix string, tag string) (DOMElement, error)

	// for detail elements

	CreateWorkElement(doc DOMDocument) (WorkElement, error)

	CreateHeadElement(doc DOMDocument) (HeadElement, error)

	CreateTimelineElement(doc DOMDocument) (TimelineElement, error)

	CreateTrackElement(doc DOMDocument) (TrackElement, error)

	CreateTrackListElement(doc DOMDocument) (TrackListElement, error)
}

type DOMDocument interface {
	DOMNode

	Implementation() DOMImplementation

	GetURL() string

	GetRoot() (WorkElement, error)

	SetRoot(root WorkElement) error
}

type DOMDocumentCodec interface {
	Encode(doc DOMDocument) ([]byte, error)
	Decode(data []byte, doc DOMDocument) error
}

type DOMElement interface {
	DOMNode

	AddChild(child *DOMElement) error

	ListChildren() []DOMElement
}

////////////////////////////////////////////////////////////////////////////////
// 以下是一些具体的 elements

type WorkElement interface {
	DOMElement
}

type HeadElement interface {
	DOMElement
}

type TrackListElement interface {
	DOMElement
}

type TrackElement interface {
	DOMElement
}

type TimelineElement interface {
	DOMElement
}

////////////////////////////////////////////////////////////////////////////////
// EOF
