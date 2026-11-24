package restful

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/ak47less/step1go/lite2/core"
	"github.com/ak47less/step1go/lite2/lib"
)

type engineResource struct {
	State    string        `json:"state"`
	Starting bool          `json:"starting"`
	Started  bool          `json:"started"`
	Stopping bool          `json:"stopping"`
	Stopped  bool          `json:"stopped"`
	Audio    audioResource `json:"audio"`
}

type audioResource struct {
	State string `json:"state"`
}

type transportResource struct {
	State     string  `json:"state"`
	Playing   bool    `json:"playing"`
	Paused    bool    `json:"paused"`
	Recording bool    `json:"recording"`
	Position  float64 `json:"position"`
	Timeline  string  `json:"timeline"`
}

type bankResource struct {
	Index       int    `json:"index"`
	Label       string `json:"label"`
	Description string `json:"description"`
}

type bankListResource struct {
	Items []bankResource `json:"items"`
}

type stepResource struct {
	Index int `json:"index"`
}

type stepListResource struct {
	Items []stepResource `json:"items"`
}

type trackResource struct {
	Index       int     `json:"index"`
	Name        string  `json:"name"`
	Label       string  `json:"label"`
	Description string  `json:"description"`
	Order       int     `json:"order"`
	Enabled     bool    `json:"enabled"`
	Solo        bool    `json:"solo"`
	Mute        bool    `json:"mute"`
	GainDB      float64 `json:"gainDb"`
	Channel     int     `json:"channel"`
}

type trackListResource struct {
	Items []trackResource `json:"items"`
}

type cellResource struct {
	Name       string            `json:"name"`
	BankIndex  int               `json:"bankIndex"`
	StepIndex  int               `json:"stepIndex"`
	TrackIndex int               `json:"trackIndex"`
	Note       *midiNoteResource `json:"note"`
}

type midiNoteResource struct {
	Number   int `json:"number"`
	Velocity int `json:"velocity"`
}

type cellListResource struct {
	Items []cellResource `json:"items"`
}

type cubeResource struct {
	Banks     cubeSizeResource  `json:"banks"`
	Steps     cubeSizeResource  `json:"steps"`
	Tracks    cubeSizeResource  `json:"tracks"`
	Timelines map[string]string `json:"timelines"`
}

type cubeSizeResource struct {
	Length   int `json:"length"`
	Capacity int `json:"capacity"`
}

func (inst *innerWebServer) handleEngine(w http.ResponseWriter, r *http.Request) {
	if !allowMethods(w, r, http.MethodGet, http.MethodHead) {
		return
	}
	writeJSON(w, http.StatusOK, inst.engine())
}

func (inst *innerWebServer) handleStartEngine(w http.ResponseWriter, r *http.Request) {
	inst.handleEngineCommand(w, r, func(ctx *lib.ActionContext) error {
		return inst.controller.StartEngine(ctx)
	})
}

func (inst *innerWebServer) handleStopEngine(w http.ResponseWriter, r *http.Request) {
	inst.handleEngineCommand(w, r, func(ctx *lib.ActionContext) error {
		return inst.controller.StopEngine(ctx)
	})
}

func (inst *innerWebServer) handleRestartEngine(w http.ResponseWriter, r *http.Request) {
	inst.handleEngineCommand(w, r, func(ctx *lib.ActionContext) error {
		return inst.controller.RestartEngine(ctx)
	})
}

func (inst *innerWebServer) handleEngineCommand(
	w http.ResponseWriter,
	r *http.Request,
	command func(*lib.ActionContext) error,
) {
	if !allowMethods(w, r, http.MethodPost) {
		return
	}
	if inst.controller == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE",
			"action controller is not available", nil)
		return
	}

	err := command(inst.controller.NewContext())
	if err != nil {
		writeActionError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, inst.engine())
}

func (inst *innerWebServer) handleTransport(w http.ResponseWriter, r *http.Request) {
	if !allowMethods(w, r, http.MethodGet, http.MethodHead) {
		return
	}
	writeJSON(w, http.StatusOK, inst.transport())
}

func (inst *innerWebServer) handlePlay(w http.ResponseWriter, r *http.Request) {
	inst.handleTransportCommand(w, r, func(ctx *lib.ActionContext) error {
		return inst.controller.Play(ctx)
	})
}

func (inst *innerWebServer) handleRecord(w http.ResponseWriter, r *http.Request) {
	inst.handleTransportCommand(w, r, func(ctx *lib.ActionContext) error {
		return inst.controller.Record(ctx)
	})
}

func (inst *innerWebServer) handlePause(w http.ResponseWriter, r *http.Request) {
	inst.handleTransportCommand(w, r, func(ctx *lib.ActionContext) error {
		return inst.controller.Pause(ctx)
	})
}

func (inst *innerWebServer) handleResume(w http.ResponseWriter, r *http.Request) {
	inst.handleTransportCommand(w, r, func(ctx *lib.ActionContext) error {
		return inst.controller.Resume(ctx)
	})
}

func (inst *innerWebServer) handleTransportStop(w http.ResponseWriter, r *http.Request) {
	inst.handleTransportCommand(w, r, func(ctx *lib.ActionContext) error {
		return inst.controller.Stop(ctx)
	})
}

func (inst *innerWebServer) handleTransportCommand(
	w http.ResponseWriter,
	r *http.Request,
	command func(*lib.ActionContext) error,
) {
	if !allowMethods(w, r, http.MethodPost) {
		return
	}
	if inst.controller == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE",
			"action controller is not available", nil)
		return
	}

	err := command(inst.controller.NewContext())
	if err != nil {
		writeActionError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, inst.transport())
}

type seekRequest struct {
	Position *float64 `json:"position"`
	Timeline string   `json:"timeline"`
}

func (inst *innerWebServer) handleSeek(w http.ResponseWriter, r *http.Request) {
	if !allowMethods(w, r, http.MethodPut) {
		return
	}

	var request seekRequest
	if err := decodeJSON(w, r, &request); err != nil {
		writeAPIError(w, http.StatusBadRequest, "INVALID_JSON", err.Error(), nil)
		return
	}
	if request.Position == nil || math.IsNaN(*request.Position) || math.IsInf(*request.Position, 0) {
		writeAPIError(w, http.StatusBadRequest, "INVALID_PARAMETER",
			"position is required and must be a finite number", nil)
		return
	}
	if *request.Position < 0 {
		writeAPIError(w, http.StatusBadRequest, "INVALID_PARAMETER",
			"position must be greater than or equal to 0", map[string]any{"field": "position"})
		return
	}

	timeline, err := inst.resolveTimeline(request.Timeline)
	if err != nil {
		writeAPIError(w, http.StatusUnprocessableEntity, "INVALID_TIMELINE", err.Error(), nil)
		return
	}
	if timeline.Length > 0 && *request.Position > float64(timeline.Length) {
		writeAPIError(w, http.StatusUnprocessableEntity, "POSITION_OUT_OF_RANGE",
			"position is outside the timeline", map[string]any{
				"position": *request.Position,
				"length":   float64(timeline.Length),
			})
		return
	}
	if inst.controller == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE",
			"action controller is not available", nil)
		return
	}

	position := &lib.RelativePosition{
		Ref: timeline,
		Pos: lib.Seconds(*request.Position),
	}
	if err := inst.controller.Seek(inst.controller.NewContext(), position); err != nil {
		writeActionError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, inst.transport())
}

func (inst *innerWebServer) handleStatus(w http.ResponseWriter, r *http.Request) {
	if !allowMethods(w, r, http.MethodGet, http.MethodHead) {
		return
	}

	keys := parseStatusKeys(r.URL.Query()["keys"])
	all := inst.statusValues()
	status := make(map[string]any)
	if len(keys) == 0 {
		status = all
	} else {
		for _, key := range keys {
			if value, ok := all[key]; ok {
				status[key] = value
			}
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"status": status})
}

func (inst *innerWebServer) handleCube(w http.ResponseWriter, r *http.Request) {
	if !allowMethods(w, r, http.MethodGet, http.MethodHead) {
		return
	}

	inst.cubeMu.RLock()
	defer inst.cubeMu.RUnlock()

	cube, ok := inst.cube()
	if !ok {
		writeAPIError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE",
			"sequencer is not initialized", nil)
		return
	}

	writeJSON(w, http.StatusOK, cubeResource{
		Banks: cubeSizeResource{
			Length:   len(cube.Banks),
			Capacity: int(cube.Buffers.BankBuffer.Capacity()),
		},
		Steps: cubeSizeResource{
			Length:   len(cube.Steps),
			Capacity: int(cube.Buffers.StepBuffer.Capacity()),
		},
		Tracks: cubeSizeResource{
			Length:   len(cube.Tracks),
			Capacity: int(cube.Buffers.TrackBuffer.Capacity()),
		},
		Timelines: inst.timelineNames(),
	})
}

func (inst *innerWebServer) handleBanks(w http.ResponseWriter, r *http.Request) {
	if !allowMethods(w, r, http.MethodGet, http.MethodHead) {
		return
	}

	inst.cubeMu.RLock()
	defer inst.cubeMu.RUnlock()

	cube, ok := inst.cube()
	if !ok {
		writeAPIError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE",
			"sequencer is not initialized", nil)
		return
	}

	items := make([]bankResource, 0, len(cube.Banks))
	for index, bank := range cube.Banks {
		items = append(items, makeBankResource(index, bank))
	}
	writeJSON(w, http.StatusOK, bankListResource{Items: items})
}

func (inst *innerWebServer) handleBank(w http.ResponseWriter, r *http.Request) {
	if !allowMethods(w, r, http.MethodGet, http.MethodHead, http.MethodPatch) {
		return
	}
	index, ok := parsePathIndex(w, r, "bankIndex")
	if !ok {
		return
	}

	if r.Method == http.MethodPatch {
		inst.patchBank(w, r, index)
		return
	}

	inst.cubeMu.RLock()
	defer inst.cubeMu.RUnlock()

	cube, ok := inst.cube()
	if !ok {
		writeAPIError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE",
			"sequencer is not initialized", nil)
		return
	}
	if index >= len(cube.Banks) {
		writeAPIError(w, http.StatusNotFound, "NOT_FOUND", "bank not found", map[string]any{"index": index})
		return
	}
	writeJSON(w, http.StatusOK, makeBankResource(index, cube.Banks[index]))
}

type bankPatch struct {
	Label       *string `json:"label"`
	Description *string `json:"description"`
}

func (inst *innerWebServer) patchBank(w http.ResponseWriter, r *http.Request, index int) {
	var patch bankPatch
	if err := decodeJSON(w, r, &patch); err != nil {
		writeAPIError(w, http.StatusBadRequest, "INVALID_JSON", err.Error(), nil)
		return
	}
	if patch.Label == nil && patch.Description == nil {
		writeAPIError(w, http.StatusBadRequest, "INVALID_PARAMETER",
			"at least one bank field is required", nil)
		return
	}

	inst.cubeMu.Lock()
	defer inst.cubeMu.Unlock()

	cube, ok := inst.cube()
	if !ok {
		writeAPIError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE",
			"sequencer is not initialized", nil)
		return
	}
	if index >= len(cube.Banks) {
		writeAPIError(w, http.StatusNotFound, "NOT_FOUND", "bank not found", map[string]any{"index": index})
		return
	}

	bank := cube.Banks[index]
	if patch.Label != nil {
		bank.Label = *patch.Label
	}
	if patch.Description != nil {
		bank.Description = *patch.Description
	}
	writeJSON(w, http.StatusOK, makeBankResource(index, bank))
}

func (inst *innerWebServer) handleSteps(w http.ResponseWriter, r *http.Request) {
	if !allowMethods(w, r, http.MethodGet, http.MethodHead) {
		return
	}

	inst.cubeMu.RLock()
	defer inst.cubeMu.RUnlock()

	cube, ok := inst.cube()
	if !ok {
		writeAPIError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE",
			"sequencer is not initialized", nil)
		return
	}

	items := make([]stepResource, 0, len(cube.Steps))
	for index := range cube.Steps {
		items = append(items, stepResource{Index: index})
	}
	writeJSON(w, http.StatusOK, stepListResource{Items: items})
}

func (inst *innerWebServer) handleStep(w http.ResponseWriter, r *http.Request) {
	if !allowMethods(w, r, http.MethodGet, http.MethodHead) {
		return
	}
	index, ok := parsePathIndex(w, r, "stepIndex")
	if !ok {
		return
	}

	inst.cubeMu.RLock()
	defer inst.cubeMu.RUnlock()

	cube, ok := inst.cube()
	if !ok {
		writeAPIError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE",
			"sequencer is not initialized", nil)
		return
	}
	if index >= len(cube.Steps) {
		writeAPIError(w, http.StatusNotFound, "NOT_FOUND", "step not found", map[string]any{"index": index})
		return
	}
	writeJSON(w, http.StatusOK, stepResource{Index: index})
}

func (inst *innerWebServer) handleTracks(w http.ResponseWriter, r *http.Request) {
	if !allowMethods(w, r, http.MethodGet, http.MethodHead) {
		return
	}

	inst.cubeMu.RLock()
	defer inst.cubeMu.RUnlock()

	cube, ok := inst.cube()
	if !ok {
		writeAPIError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE",
			"sequencer is not initialized", nil)
		return
	}

	items := make([]trackResource, 0, len(cube.Tracks))
	for index, track := range cube.Tracks {
		items = append(items, makeTrackResource(index, track))
	}
	writeJSON(w, http.StatusOK, trackListResource{Items: items})
}

func (inst *innerWebServer) handleTrack(w http.ResponseWriter, r *http.Request) {
	if !allowMethods(w, r, http.MethodGet, http.MethodHead, http.MethodPatch) {
		return
	}
	index, ok := parsePathIndex(w, r, "trackIndex")
	if !ok {
		return
	}

	if r.Method == http.MethodPatch {
		inst.patchTrack(w, r, index)
		return
	}

	inst.cubeMu.RLock()
	defer inst.cubeMu.RUnlock()

	cube, ok := inst.cube()
	if !ok {
		writeAPIError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE",
			"sequencer is not initialized", nil)
		return
	}
	if index >= len(cube.Tracks) {
		writeAPIError(w, http.StatusNotFound, "NOT_FOUND", "track not found", map[string]any{"index": index})
		return
	}
	writeJSON(w, http.StatusOK, makeTrackResource(index, cube.Tracks[index]))
}

type trackPatch struct {
	Name        *string  `json:"name"`
	Label       *string  `json:"label"`
	Description *string  `json:"description"`
	Order       *int     `json:"order"`
	Enabled     *bool    `json:"enabled"`
	Solo        *bool    `json:"solo"`
	Mute        *bool    `json:"mute"`
	GainDB      *float64 `json:"gainDb"`
	Channel     *int     `json:"channel"`
}

func (patch trackPatch) empty() bool {
	return patch.Name == nil &&
		patch.Label == nil &&
		patch.Description == nil &&
		patch.Order == nil &&
		patch.Enabled == nil &&
		patch.Solo == nil &&
		patch.Mute == nil &&
		patch.GainDB == nil &&
		patch.Channel == nil
}

func (inst *innerWebServer) patchTrack(w http.ResponseWriter, r *http.Request, index int) {
	var patch trackPatch
	if err := decodeJSON(w, r, &patch); err != nil {
		writeAPIError(w, http.StatusBadRequest, "INVALID_JSON", err.Error(), nil)
		return
	}
	if patch.empty() {
		writeAPIError(w, http.StatusBadRequest, "INVALID_PARAMETER",
			"at least one track field is required", nil)
		return
	}
	if patch.GainDB != nil && (math.IsNaN(*patch.GainDB) || math.IsInf(*patch.GainDB, 0)) {
		writeAPIError(w, http.StatusBadRequest, "INVALID_PARAMETER",
			"gainDb must be a finite number", map[string]any{"field": "gainDb"})
		return
	}
	if patch.Channel != nil && (*patch.Channel < -1 || *patch.Channel > 127) {
		writeAPIError(w, http.StatusBadRequest, "INVALID_PARAMETER",
			"channel must be between -1 and 127", map[string]any{"field": "channel"})
		return
	}

	inst.cubeMu.Lock()
	defer inst.cubeMu.Unlock()

	cube, ok := inst.cube()
	if !ok {
		writeAPIError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE",
			"sequencer is not initialized", nil)
		return
	}
	if index >= len(cube.Tracks) {
		writeAPIError(w, http.StatusNotFound, "NOT_FOUND", "track not found", map[string]any{"index": index})
		return
	}

	track := cube.Tracks[index]
	if patch.Name != nil {
		track.Name = lib.TrackName(*patch.Name)
	}
	if patch.Label != nil {
		track.Label = *patch.Label
	}
	if patch.Description != nil {
		track.Description = *patch.Description
	}
	if patch.Order != nil {
		track.Order = *patch.Order
	}
	if patch.Enabled != nil {
		track.Enabled = *patch.Enabled
	}
	if patch.Solo != nil {
		track.Solo = *patch.Solo
	}
	if patch.Mute != nil {
		track.Mute = *patch.Mute
	}
	if patch.GainDB != nil {
		track.Gain = lib.GainDB(*patch.GainDB)
	}
	if patch.Channel != nil {
		track.Channel = int8(*patch.Channel)
	}
	writeJSON(w, http.StatusOK, makeTrackResource(index, track))
}

func (inst *innerWebServer) handleCells(w http.ResponseWriter, r *http.Request) {
	if !allowMethods(w, r, http.MethodGet, http.MethodHead) {
		return
	}

	bankIndex, bankSet, ok := parseQueryIndex(w, r, "bankIndex")
	if !ok {
		return
	}
	stepIndex, stepSet, ok := parseQueryIndex(w, r, "stepIndex")
	if !ok {
		return
	}
	trackIndex, trackSet, ok := parseQueryIndex(w, r, "trackIndex")
	if !ok {
		return
	}

	inst.cubeMu.RLock()
	defer inst.cubeMu.RUnlock()

	cube, ok := inst.cube()
	if !ok {
		writeAPIError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE",
			"sequencer is not initialized", nil)
		return
	}
	if (bankSet && bankIndex >= len(cube.Banks)) ||
		(stepSet && stepIndex >= len(cube.Steps)) ||
		(trackSet && trackIndex >= len(cube.Tracks)) {
		writeAPIError(w, http.StatusNotFound, "NOT_FOUND", "cell filter is outside the cube", nil)
		return
	}

	cells := cube.Cells.ListAllCells()
	items := make([]cellResource, 0, len(cells))
	for _, cell := range cells {
		if bankSet && int(cell.BankIndex) != bankIndex {
			continue
		}
		if stepSet && int(cell.StepIndex) != stepIndex {
			continue
		}
		if trackSet && int(cell.TrackIndex) != trackIndex {
			continue
		}
		items = append(items, makeCellResource(cell))
	}
	writeJSON(w, http.StatusOK, cellListResource{Items: items})
}

func (inst *innerWebServer) handleCell(w http.ResponseWriter, r *http.Request) {
	if !allowMethods(w, r, http.MethodGet, http.MethodHead, http.MethodPatch) {
		return
	}

	if r.Method == http.MethodPatch {
		inst.patchCell(w, r)
		return
	}

	inst.cubeMu.RLock()
	defer inst.cubeMu.RUnlock()

	cube, ok := inst.cube()
	if !ok {
		writeAPIError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE",
			"sequencer is not initialized", nil)
		return
	}

	name := lib.CellName(r.PathValue("cellName"))
	cell, err := cube.Cells.GetCellByName(name)
	if err != nil {
		writeAPIError(w, http.StatusNotFound, "NOT_FOUND", "cell not found", map[string]any{"name": name})
		return
	}
	writeJSON(w, http.StatusOK, makeCellResource(cell))
}

type cellPatch struct {
	Note json.RawMessage `json:"note"`
}

type midiNotePatch struct {
	Number   *int `json:"number"`
	Velocity *int `json:"velocity"`
}

func (inst *innerWebServer) patchCell(w http.ResponseWriter, r *http.Request) {
	var patch cellPatch
	if err := decodeJSON(w, r, &patch); err != nil {
		writeAPIError(w, http.StatusBadRequest, "INVALID_JSON", err.Error(), nil)
		return
	}
	if len(patch.Note) == 0 {
		writeAPIError(w, http.StatusBadRequest, "INVALID_PARAMETER",
			"note is required", map[string]any{"field": "note"})
		return
	}

	var note *lib.MidiNote
	if !bytes.Equal(bytes.TrimSpace(patch.Note), []byte("null")) {
		var request midiNotePatch
		if err := decodeRawJSON(patch.Note, &request); err != nil {
			writeAPIError(w, http.StatusBadRequest, "INVALID_JSON", err.Error(), nil)
			return
		}
		switch {
		case request.Number == nil:
			writeAPIError(w, http.StatusBadRequest, "INVALID_PARAMETER",
				"note.number is required", map[string]any{"field": "note.number"})
			return
		case *request.Number < 0 || *request.Number > 127:
			writeAPIError(w, http.StatusBadRequest, "INVALID_PARAMETER",
				"note.number must be between 0 and 127", map[string]any{"field": "note.number"})
			return
		case request.Velocity == nil:
			writeAPIError(w, http.StatusBadRequest, "INVALID_PARAMETER",
				"note.velocity is required", map[string]any{"field": "note.velocity"})
			return
		case *request.Velocity < 0 || *request.Velocity > 127:
			writeAPIError(w, http.StatusBadRequest, "INVALID_PARAMETER",
				"note.velocity must be between 0 and 127", map[string]any{"field": "note.velocity"})
			return
		}
		note = &lib.MidiNote{
			Number:   core.MidiNoteNumber(*request.Number),
			Velocity: core.MidiNoteVelocity(*request.Velocity),
		}
	}

	inst.cubeMu.Lock()
	defer inst.cubeMu.Unlock()

	cube, ok := inst.cube()
	if !ok {
		writeAPIError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE",
			"sequencer is not initialized", nil)
		return
	}

	name := lib.CellName(r.PathValue("cellName"))
	cell, err := cube.Cells.GetCellByName(name)
	if err != nil {
		writeAPIError(w, http.StatusNotFound, "NOT_FOUND", "cell not found", map[string]any{"name": name})
		return
	}

	cell.Note = note
	writeJSON(w, http.StatusOK, makeCellResource(cell))
}

func (inst *innerWebServer) cube() (*lib.Cube, bool) {
	if inst.context == nil || inst.context.SequencerContext == nil {
		return nil, false
	}
	cube := &inst.context.SequencerContext.Cube
	if cube.Cells == nil || cube.Banks == nil || cube.Steps == nil || cube.Tracks == nil {
		return nil, false
	}
	return cube, true
}

func (inst *innerWebServer) engine() engineResource {
	ec := inst.context.EngineContext
	res := engineResource{}
	if ec == nil {
		res.State = "stopped"
		res.Audio.State = "closed"
		return res
	}

	res.Starting = ec.Starting
	res.Started = ec.Started
	res.Stopping = ec.Stopping
	res.Stopped = ec.Stopped
	switch {
	case ec.Starting:
		res.State = "starting"
	case ec.Stopping:
		res.State = "stopping"
	case ec.Started:
		res.State = "started"
	default:
		res.State = "stopped"
	}
	if ec.AudioContext == nil {
		res.Audio.State = "closed"
	} else {
		res.Audio.State = string(ec.AudioContext.GetState())
	}
	return res
}

func (inst *innerWebServer) transport() transportResource {
	tc := inst.context.TransportContext
	res := transportResource{
		State:    "stopped",
		Timeline: "@document",
	}
	if tc == nil {
		return res
	}

	res.Playing = tc.Playing
	res.Paused = tc.Paused
	res.Recording = tc.Recording
	switch {
	case tc.Paused:
		res.State = "paused"
	case tc.Recording:
		res.State = "recording"
	case tc.Playing:
		res.State = "playing"
	}

	if inst.context.Transport != nil {
		position := inst.context.Transport.GetCurrentPosition()
		if position != nil {
			res.Position = float64(position.Pos)
			if position.Ref != nil && position.Ref.Name != "" {
				res.Timeline = position.Ref.Name
			}
		}
	}
	return res
}

func (inst *innerWebServer) statusValues() map[string]any {
	engine := inst.engine()
	transport := inst.transport()
	return map[string]any{
		"engine.starting":     boolInt(engine.Starting),
		"engine.started":      boolInt(engine.Started),
		"engine.stopping":     boolInt(engine.Stopping),
		"engine.stopped":      boolInt(engine.Stopped),
		"audio.state":         engine.Audio.State,
		"transport.playing":   boolInt(transport.Playing),
		"transport.paused":    boolInt(transport.Paused),
		"transport.recording": boolInt(transport.Recording),
		"transport.position":  transport.Position,
	}
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func parseStatusKeys(values []string) []string {
	keys := make([]string, 0)
	seen := make(map[string]struct{})
	for _, value := range values {
		for _, key := range strings.Split(value, ",") {
			key = strings.TrimSpace(key)
			if key == "" {
				continue
			}
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			keys = append(keys, key)
		}
	}
	return keys
}

func parseQueryIndex(w http.ResponseWriter, r *http.Request, name string) (int, bool, bool) {
	values := r.URL.Query()[name]
	if len(values) == 0 || values[0] == "" {
		return 0, false, true
	}
	value, err := strconv.Atoi(values[0])
	if err != nil || value < 0 {
		writeAPIError(w, http.StatusBadRequest, "INVALID_QUERY_PARAMETER",
			fmt.Sprintf("%s must be a non-negative integer", name),
			map[string]any{"field": name, "value": values[0]})
		return 0, false, false
	}
	return value, true, true
}

func (inst *innerWebServer) resolveTimeline(name string) (*lib.Timeline, error) {
	ec := inst.context.EngineContext
	if ec == nil {
		return nil, fmt.Errorf("engine context is not initialized")
	}
	if name == "" {
		name = "@document"
	}

	tls := &ec.Timelines

	switch strings.TrimPrefix(name, "@") {
	case "root":
		if tls.Root != nil {
			return tls.Root, nil
		}
	case "document":
		if tls.Document != nil {
			return tls.Document, nil
		}
	case "engine":
		if tls.Engine != nil {
			return tls.Engine, nil
		}
	case "loop":
		if tls.Loop != nil {
			return tls.Loop, nil
		}
	case "frame":
		if tls.Frame != nil {
			return tls.Frame, nil
		}
	case "transport":
		if tls.Transport != nil {
			return tls.Transport, nil
		}
	case "step":
		if tls.Step != nil {
			return tls.Step, nil
		}
	}
	return nil, fmt.Errorf("timeline %q does not exist", name)
}

func (inst *innerWebServer) timelineNames() map[string]string {
	return map[string]string{
		"document":  "@document",
		"transport": "@transport",
		"step":      "@step",
		"frame":     "@frame",
	}
}

func makeBankResource(index int, bank *lib.Bank) bankResource {
	if bank == nil {
		return bankResource{Index: index}
	}
	return bankResource{
		Index:       index,
		Label:       bank.Label,
		Description: bank.Description,
	}
}

func makeTrackResource(index int, track *lib.Track) trackResource {
	if track == nil {
		return trackResource{Index: index, Channel: -1}
	}
	return trackResource{
		Index:       index,
		Name:        string(track.Name),
		Label:       track.Label,
		Description: track.Description,
		Order:       track.Order,
		Enabled:     track.Enabled,
		Solo:        track.Solo,
		Mute:        track.Mute,
		GainDB:      float64(track.Gain),
		Channel:     int(track.Channel),
	}
}

func makeCellResource(cell *lib.Cell) cellResource {
	if cell == nil {
		return cellResource{}
	}
	resource := cellResource{
		Name:       string(cell.Name),
		BankIndex:  int(cell.BankIndex),
		StepIndex:  int(cell.StepIndex),
		TrackIndex: int(cell.TrackIndex),
	}
	if cell.Note != nil {
		resource.Note = &midiNoteResource{
			Number:   int(cell.Note.Number),
			Velocity: int(cell.Note.Velocity),
		}
	}
	return resource
}

type batchActionRequest struct {
	Actions []batchAction `json:"actions"`
}

type batchAction struct {
	Name   string          `json:"name"`
	Params json.RawMessage `json:"params"`
	All    *bool           `json:"all"`
	Keys   []string        `json:"keys"`
}

type batchResult struct {
	Index int              `json:"index"`
	Name  string           `json:"name"`
	OK    bool             `json:"ok"`
	Error *apiErrorPayload `json:"error,omitempty"`
}

type batchResponse struct {
	Error   *apiErrorPayload `json:"error,omitempty"`
	Results []batchResult    `json:"results"`
	Status  map[string]any   `json:"status,omitempty"`
}

type statusActionParams struct {
	All  *bool     `json:"all"`
	Keys *[]string `json:"keys"`
}

type seekActionParams struct {
	Position *float64 `json:"position"`
	Timeline string   `json:"timeline"`
}

func (inst *innerWebServer) handleActions(w http.ResponseWriter, r *http.Request) {
	if !allowMethods(w, r, http.MethodPost) {
		return
	}

	var request batchActionRequest
	if err := decodeJSON(w, r, &request); err != nil {
		writeAPIError(w, http.StatusBadRequest, "INVALID_JSON", err.Error(), nil)
		return
	}
	if len(request.Actions) == 0 {
		writeAPIError(w, http.StatusBadRequest, "INVALID_PARAMETER",
			"actions must contain at least one action", nil)
		return
	}

	results := make([]batchResult, 0, len(request.Actions))
	status := make(map[string]any)
	hasStatus := false

	for index, action := range request.Actions {
		result := batchResult{
			Index: index,
			Name:  action.Name,
			OK:    true,
		}
		err := inst.executeBatchAction(action, &hasStatus, status)
		if err != nil {
			statusCode, payload := actionError(err)
			result.OK = false
			result.Error = &payload
			results = append(results, result)
			writeJSON(w, statusCode, batchResponse{
				Error:   &payload,
				Results: results,
			})
			return
		}
		results = append(results, result)
	}

	response := batchResponse{Results: results}
	if hasStatus {
		response.Status = status
	}
	writeJSON(w, http.StatusOK, response)
}

func (inst *innerWebServer) executeBatchAction(
	action batchAction,
	hasStatus *bool,
	status map[string]any,
) error {
	name := strings.TrimSpace(action.Name)
	switch name {
	case "start_engine", "stop_engine", "restart_engine",
		"play", "record", "pause", "resume", "stop":
		if inst.controller == nil {
			return newRequestError(http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE",
				"action controller is not available")
		}
		ctx := inst.controller.NewContext()
		switch name {
		case "start_engine":
			return inst.controller.StartEngine(ctx)
		case "stop_engine":
			return inst.controller.StopEngine(ctx)
		case "restart_engine":
			return inst.controller.RestartEngine(ctx)
		case "play":
			return inst.controller.Play(ctx)
		case "record":
			return inst.controller.Record(ctx)
		case "pause":
			return inst.controller.Pause(ctx)
		case "resume":
			return inst.controller.Resume(ctx)
		default:
			return inst.controller.Stop(ctx)
		}
	case "seek":
		if inst.controller == nil {
			return newRequestError(http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE",
				"action controller is not available")
		}
		var params seekActionParams
		if len(action.Params) == 0 || string(action.Params) == "null" {
			return newRequestError(http.StatusBadRequest, "INVALID_PARAMETER",
				"seek requires params")
		}
		if err := decodeRawJSON(action.Params, &params); err != nil {
			return newRequestError(http.StatusBadRequest, "INVALID_PARAMETER", err.Error())
		}
		if params.Position == nil || math.IsNaN(*params.Position) || math.IsInf(*params.Position, 0) ||
			*params.Position < 0 {
			return newRequestError(http.StatusBadRequest, "INVALID_PARAMETER",
				"position must be a finite number greater than or equal to 0")
		}
		timeline, err := inst.resolveTimeline(params.Timeline)
		if err != nil {
			return newRequestError(http.StatusUnprocessableEntity, "INVALID_TIMELINE", err.Error())
		}
		if timeline.Length > 0 && *params.Position > float64(timeline.Length) {
			return newRequestError(http.StatusUnprocessableEntity, "POSITION_OUT_OF_RANGE",
				"position is outside the timeline")
		}
		position := &lib.RelativePosition{
			Ref: timeline,
			Pos: lib.Seconds(*params.Position),
		}
		return inst.controller.Seek(inst.controller.NewContext(), position)
	case "status":
		values, keys, err := inst.statusActionValues(action)
		if err != nil {
			return err
		}
		*hasStatus = true
		if all, ok := values["all"].(bool); ok && all {
			for key, value := range inst.statusValues() {
				status[key] = value
			}
			return nil
		}
		allValues := inst.statusValues()
		for _, key := range keys {
			if value, ok := allValues[key]; ok {
				status[key] = value
			}
		}
		return nil
	default:
		return newRequestError(http.StatusBadRequest, "UNKNOWN_ACTION",
			fmt.Sprintf("unknown action %q", action.Name))
	}
}

func (inst *innerWebServer) statusActionValues(action batchAction) (map[string]any, []string, error) {
	var params statusActionParams
	if len(action.Params) > 0 && string(action.Params) != "null" {
		if err := decodeRawJSON(action.Params, &params); err != nil {
			return nil, nil, newRequestError(http.StatusBadRequest, "INVALID_PARAMETER", err.Error())
		}
	}

	allFlag := action.All
	if params.All != nil {
		allFlag = params.All
	}
	keys := action.Keys
	if params.Keys != nil {
		keys = *params.Keys
	}

	all := true
	if allFlag != nil {
		all = *allFlag
	} else if action.Keys != nil || params.Keys != nil {
		all = false
	}
	return map[string]any{"all": all}, parseStatusKeys(keys), nil
}

type requestError struct {
	status  int
	code    string
	message string
}

func newRequestError(status int, code, message string) *requestError {
	return &requestError{status: status, code: code, message: message}
}

func (err *requestError) Error() string {
	return err.message
}

func actionError(err error) (int, apiErrorPayload) {
	var requestErr *requestError
	if errors.As(err, &requestErr) {
		return requestErr.status, apiErrorPayload{Code: requestErr.code, Message: requestErr.message}
	}

	message := strings.ToLower(err.Error())
	switch {
	case strings.Contains(message, "not started"),
		strings.Contains(message, "already"),
		strings.Contains(message, "state"),
		strings.Contains(message, "paused"),
		strings.Contains(message, "playing"),
		strings.Contains(message, "recording"):
		return http.StatusConflict, apiErrorPayload{
			Code:    "INVALID_STATE",
			Message: err.Error(),
		}
	case strings.Contains(message, "audio"),
		strings.Contains(message, "driver"),
		strings.Contains(message, "unavailable"):
		return http.StatusServiceUnavailable, apiErrorPayload{
			Code:    "SERVICE_UNAVAILABLE",
			Message: err.Error(),
		}
	default:
		return http.StatusUnprocessableEntity, apiErrorPayload{
			Code:    "ACTION_FAILED",
			Message: err.Error(),
		}
	}
}

func writeActionError(w http.ResponseWriter, err error) {
	status, payload := actionError(err)
	writeJSON(w, status, apiErrorResponse{Error: payload})
}

func decodeRawJSON(data []byte, dst any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		return fmt.Errorf("invalid params: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return fmt.Errorf("params must contain one JSON object")
	}
	return nil
}
