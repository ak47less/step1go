package lite

import "fmt"

type CubeBuffers struct {
	BankBuffer  BankArray
	StepBuffer  StepArray
	TrackBuffer TrackArray
}

// Cube 是一个由{bank,step,track}三个维度构成的立方体
type Cube struct {
	Context *SequencerContext

	CellManager CellManager
	Buffers     CubeBuffers

	Banks  BankSlice
	Tracks TrackSlice
	Steps  StepSlice
}

func (inst *Cube) Init(ctx *SequencerContext, cfg *Configuration) error {

	if inst == nil {
		return fmt.Errorf("cube-instance is nil")
	}

	if ctx == nil && cfg == nil {
		return fmt.Errorf("param(s) is nil")
	}

	if cfg == nil {
		cfg = ctx.EngineContext.Configuration
	}

	if cfg == nil {
		return fmt.Errorf("engine.configuration is nil")
	}

	ci := new(innerCubeInitializer)
	ci.cube = inst
	return ci.init(ctx, cfg)
}

// 根据当前状态重新加载这个 cube 中的各个列表
func (inst *Cube) Reload(ibank BankIndex, all bool) error {

	var err error
	src := inst.Buffers.BankBuffer

	if all {
		for _, item := range src {
			if item == nil {
				continue
			}
			err = item.Reload()
			if err != nil {
				return err
			}
		}
		return nil
	}

	bank := src.GetItem(ibank)
	if bank == nil {
		return nil
	}
	return bank.Reload()
}

////////////////////////////////////////////////////////////////////////////////

type innerCubeInitializer struct {
	cube *Cube
}

func (inst *innerCubeInitializer) init(ctx *SequencerContext, cfg *Configuration) error {

	var err error

	cube := inst.cube
	bufs := &cube.Buffers

	bufs.BankBuffer = bufs.BankBuffer.Init(cfg.BankListCapacity)
	bufs.StepBuffer = bufs.StepBuffer.Init(cfg.StepListCapacity)
	bufs.TrackBuffer = bufs.TrackBuffer.Init(cfg.TrackListCapacity)

	cube.Context = ctx
	cube.Banks = bufs.BankBuffer.GetSlice(cfg.BankListSize)
	cube.Steps = bufs.StepBuffer.GetSlice(cfg.StepListSize)
	cube.Tracks = bufs.TrackBuffer.GetSlice(cfg.TrackListSize)

	inst.initBankHead()
	inst.initStepHead()
	inst.initTrackHead()

	err = inst.initBanks()
	if err != nil {
		return err
	}

	return nil
}

func (inst *innerCubeInitializer) initTrackHead() error {

	cube := inst.cube
	tracks := cube.Buffers.TrackBuffer

	for idx, item := range tracks {

		meta := new(TrackMeta)
		itrack := TrackIndex(idx)

		meta.Index = itrack

		item.Index = itrack
		item.Meta = meta
		item.cube = cube
	}

	return nil
}

func (inst *innerCubeInitializer) initBankHead() error {

	cube := inst.cube
	banks := cube.Buffers.BankBuffer

	for idx, item := range banks {
		item.Index = BankIndex(idx)
		item.cube = cube
	}

	return nil
}

func (inst *innerCubeInitializer) initStepHead() error {

	cube := inst.cube
	steps := cube.Buffers.StepBuffer

	for idx, item := range steps {
		item.Index = StepIndex(idx)
		item.cube = cube
	}

	return nil
}

func (inst *innerCubeInitializer) initBanks() error {

	all := inst.cube.Buffers.BankBuffer

	for _, bank := range all {
		err := inst.initBank(bank)
		if err != nil {
			return err
		}
	}

	return nil
}

func (inst *innerCubeInitializer) initBank(bank *Bank) error {

	cube := inst.cube

	step_cap := cube.Buffers.StepBuffer.Capacity()
	track_cap := cube.Buffers.TrackBuffer.Capacity()
	step_len := cube.Steps.Length()
	track_len := cube.Tracks.Length()

	bank.StepBuffer = bank.StepBuffer.Init(step_cap)
	bank.TrackBuffer = bank.TrackBuffer.Init(track_cap)

	bank.Steps = bank.StepBuffer.GetSlice(step_len)
	bank.Tracks = bank.TrackBuffer.GetSlice(track_len)

	var istep, itrack uint

	want := new(Cell)
	cnt_track := track_cap
	cnt_step := step_cap
	steps := bank.StepBuffer
	tracks := bank.TrackBuffer
	share_tracks := cube.Buffers.TrackBuffer
	cm := &cube.CellManager
	bank_idx := bank.Index

	for itrack = 0; itrack < cnt_track; itrack++ {

		track_idx := TrackIndex(itrack)
		track_ref := tracks[itrack]

		share_track_ref := share_tracks[itrack]
		meta := share_track_ref.Meta

		track_ref.cube = cube
		track_ref.Meta = meta

		for istep = 0; istep < cnt_step; istep++ {

			step_idx := StepIndex(istep)
			step_ref := steps[istep]
			step_ref.cube = cube

			want.BankIndex = bank_idx
			want.StepIndex = step_idx
			want.TrackIndex = track_idx

			have, err := cm.OpenCell(want, true)
			if err != nil {
				return err
			}

			have.ownerBank = bank
			have.ownerStep = step_ref
			have.ownerTrack = track_ref
			have.cube = cube

			track_ref.Cells.add(have)
			step_ref.Cells.add(have)
		}
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////
// EOF
