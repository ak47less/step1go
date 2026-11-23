package common

////////////////////////////////////////////////////////////////////////////////

// 以 百分比 的形式表示增益, 100(%) 代表 0dB
type GainPercent float32

func (value GainPercent) DB() GainDB {
	return 0 // todo : no impl
}

func (value GainPercent) Rate() GainRate {
	return 0 // todo : no impl
}

////////////////////////////////////////////////////////////////////////////////

// 以对数的形式表示增益, 0(dB) 代表 100%
type GainDB float32

func (value GainDB) Percent() GainPercent {
	return 0 // todo : no impl
}

func (value GainDB) Rate() GainRate {
	return 0 // todo : no impl
}

////////////////////////////////////////////////////////////////////////////////

// 以比率的形式表示增益, 1.0 代表 100%
type GainRate float32

func (value GainRate) DB() GainDB {
	return 0 // todo : no impl
}

func (value GainRate) Percent() GainPercent {
	return 0 // todo : no impl
}

////////////////////////////////////////////////////////////////////////////////
// EOF
