package common

import "math"

////////////////////////////////////////////////////////////////////////////////

// 以 百分比 的形式表示增益, 100(%) 代表 0dB
type GainPercent float32

func (value GainPercent) DB() GainDB {
	return value.Rate().DB()
}

func (value GainPercent) Rate() GainRate {
	return GainRate(value / 100)
}

////////////////////////////////////////////////////////////////////////////////

// 以对数的形式表示增益, 0(dB) 代表 100%
type GainDB float32

func (value GainDB) Percent() GainPercent {
	return value.Rate().Percent()
}

func (value GainDB) Rate() GainRate {
	return GainRate(math.Pow(10, float64(value)/20))
}

////////////////////////////////////////////////////////////////////////////////

// 以比率的形式表示增益, 1.0 代表 100%
type GainRate float32

func (value GainRate) DB() GainDB {
	return GainDB(20 * math.Log10(float64(value)))
}

func (value GainRate) Percent() GainPercent {
	return GainPercent(value * 100)
}

////////////////////////////////////////////////////////////////////////////////
// EOF
