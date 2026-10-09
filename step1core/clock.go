package step1core

type Clock interface {

	// 更新bus中的时钟
	Next(c *Bus)
}
