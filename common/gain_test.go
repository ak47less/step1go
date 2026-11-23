package common

import (
	"fmt"
	"math"
	"testing"
)

const gainTolerance = 1e-5

func assertGainClose(t *testing.T, got, want float32) {
	t.Helper()
	if diff := math.Abs(float64(got - want)); diff > gainTolerance {
		t.Errorf("got %f, want %f (tolerance %g)", got, want, gainTolerance)
	}
}

func TestGainPercentConversions(t *testing.T) {
	cases := []struct {
		name    string
		percent GainPercent
		rate    GainRate
		db      GainDB
	}{
		{name: "unity", percent: 100, rate: 1, db: 0},
		{name: "half", percent: 50, rate: 0.5, db: -6.0206},
		{name: "double", percent: 200, rate: 2, db: 6.0206},
		{name: "one-tenth", percent: 10, rate: 0.1, db: -20},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assertGainClose(t, float32(c.percent.Rate()), float32(c.rate))
			assertGainClose(t, float32(c.percent.DB()), float32(c.db))
		})
	}
}

func TestGainDBConversions(t *testing.T) {
	cases := []struct {
		name    string
		db      GainDB
		percent GainPercent
		rate    GainRate
	}{
		{name: "unity", db: 0, percent: 100, rate: 1},
		{name: "half", db: -6.0206, percent: 50, rate: 0.5},
		{name: "double", db: 6.0206, percent: 200, rate: 2},
		{name: "one-tenth", db: -20, percent: 10, rate: 0.1},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assertGainClose(t, float32(c.db.Rate()), float32(c.rate))
			assertGainClose(t, float32(c.db.Percent()), float32(c.percent))
		})
	}
}

func TestGainRateConversions(t *testing.T) {
	cases := []struct {
		name    string
		rate    GainRate
		percent GainPercent
		db      GainDB
	}{
		{name: "unity", rate: 1, percent: 100, db: 0},
		{name: "half", rate: 0.5, percent: 50, db: -6.0206},
		{name: "double", rate: 2, percent: 200, db: 6.0206},
		{name: "one-tenth", rate: 0.1, percent: 10, db: -20},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assertGainClose(t, float32(c.rate.Percent()), float32(c.percent))
			assertGainClose(t, float32(c.rate.DB()), float32(c.db))
		})
	}
}

func TestGainConversionsRoundTrip(t *testing.T) {
	for _, rate := range []GainRate{0.01, 0.1, 0.5, 1, 2, 10, 100} {
		t.Run(fmt.Sprintf("rate[%g]", rate), func(t *testing.T) {
			got := rate.DB().Rate()
			assertGainClose(t, float32(got), float32(rate))
		})
		t.Run(fmt.Sprintf("percent[%g]", rate), func(t *testing.T) {
			got := rate.Percent().Rate()
			assertGainClose(t, float32(got), float32(rate))
		})
	}

	for _, db := range []GainDB{-40, -20, -6.0206, 0, 6.0206, 20} {
		t.Run(fmt.Sprintf("db[%g]", db), func(t *testing.T) {
			got := db.Rate().DB()
			assertGainClose(t, float32(got), float32(db))
		})
	}
}
