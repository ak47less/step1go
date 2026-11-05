package lite

import "testing"

func TestCube(t *testing.T) {

	var err error

	sc := new(SequencerContext)
	cfg := new(Configuration)
	cube := new(Cube)

	cfg.BankListCapacity = 9
	cfg.StepListCapacity = 32
	cfg.TrackListCapacity = 12

	cfg.BankListSize = 4
	cfg.StepListSize = 16
	cfg.TrackListSize = 8

	err = cube.Init(sc, cfg)

	if err != nil {
		t.Error(err)
	}
}
