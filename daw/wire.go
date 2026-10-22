package step1core

type PortRef struct {
	Device DeviceName

	Port PortName
}

type WireConfig struct {
	Source PortRef

	Destination PortRef
}

type Wire struct {
	Priority int

	Source Output

	Destination Input
}
