package lib

type StepIndex int

type StepCount uint

type Step struct {
	Cube *Cube
}

type StepList []*Step

type StepArray StepList

type StepSlice StepList
