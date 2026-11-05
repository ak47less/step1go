package main

import (
	"math/rand"

	"github.com/ak47less/step1go/lite"
)

type ExampleCubeDataLoader struct {
	ec *lite.Context
}

// GetInfo implements [lite.Part].
func (inst *ExampleCubeDataLoader) GetInfo() *lite.PartInfo {
	i := new(lite.PartInfo)
	i.Name = "ExampleCubeDataLoader"
	i.Part = inst
	return i
}

// GetLifeCycle implements [lite.Life].
func (inst *ExampleCubeDataLoader) GetLifeCycle() *lite.LifeCycle {
	lc := new(lite.LifeCycle)
	lc.OnInit = inst.oninit
	lc.OnStart = inst.onstart
	return lc
}

func (inst *ExampleCubeDataLoader) oninit(ec *lite.Context) error {
	inst.ec = ec
	return nil
}

func (inst *ExampleCubeDataLoader) onstart() error {

	cube := inst.ec.SequencerContext.Cube

	cm := &cube.CellManager
	banks := cube.Banks
	steps := cube.Steps
	tracks := cube.Tracks
	want := new(lite.Cell)

	for _, bank := range banks {
		for _, track := range tracks {
			for _, step := range steps {

				want.BankIndex = bank.Index
				want.StepIndex = step.Index
				want.TrackIndex = track.Index

				have, err := cm.OpenCell(want, false)

				if err == nil && have != nil {
					inst.innerGenRandomNoteForCell(have)
				}
			}
		}
	}

	return nil
}

func (inst *ExampleCubeDataLoader) innerGenRandomNoteForCell(cell *lite.SimpleCell) error {

	var data [4]byte
	rand.Read(data[:])
	note := new(lite.Note)

	b0 := data[0]
	b1 := data[1]

	k := lite.MidiNoteNumber(b0)
	v := lite.MidiNoteVelocity(b1)

	if 20 < k && k < 100 {
		if 50 < v && v < 125 {
			note.Key = k
			note.Velocity = v
			cell.Note = note
			cell.Enabled = true
		}
	}

	return nil
}

func (inst *ExampleCubeDataLoader) _impl() (lite.Part, lite.Life) {
	return inst, inst
}
