package step1go

type Clock interface {
	Next(c *Bus)
}
