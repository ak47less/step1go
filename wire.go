package step1go

type PortRef struct {
	Element ElementName

	Port PortName
}

type WireConfig struct {
	Source PortRef

	Destination PortRef
}

type Wire struct {
	Source Output

	Destination Input
}
