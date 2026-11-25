package restful

import (
	"fmt"
	"math"
	"net/http"
	"strings"

	"github.com/ak47less/step1go/lite2/core"
	"github.com/ak47less/step1go/lite2/lib"
)

type midiEventsRequest struct {
	Events []midiEventRequest `json:"events"`
}

type midiEventRequest struct {
	Type    string                `json:"type"`
	Time    *float64              `json:"time"`
	Channel *int                  `json:"channel"`
	Note    *midiEventNoteRequest `json:"note"`
	CC      *midiEventCCRequest   `json:"cc"`
}

type midiEventNoteRequest struct {
	Number   *int `json:"number"`
	Velocity *int `json:"velocity"`
}

type midiEventCCRequest struct {
	Key   *int `json:"key"`
	Value *int `json:"value"`
}

type midiEventsResponse struct {
	Accepted int               `json:"accepted"`
	Results  []midiEventResult `json:"results"`
}

type midiEventResult struct {
	Index int    `json:"index"`
	Type  string `json:"type"`
	OK    bool   `json:"ok"`
}

func (inst *innerWebServer) handleMIDIEvents(w http.ResponseWriter, r *http.Request) {
	if !allowMethods(w, r, http.MethodPost) {
		return
	}

	var request midiEventsRequest
	if err := decodeJSON(w, r, &request); err != nil {
		writeAPIError(w, http.StatusBadRequest, "INVALID_JSON", err.Error(), nil)
		return
	}
	if len(request.Events) == 0 {
		writeAPIError(w, http.StatusBadRequest, "INVALID_PARAMETER",
			"events must contain at least one event", nil)
		return
	}

	events := make([]*lib.MidiEvent, len(request.Events))
	results := make([]midiEventResult, len(request.Events))
	for index, item := range request.Events {
		event, eventType, err := makeMIDIEvent(item, index)
		if err != nil {
			writeAPIError(w, http.StatusBadRequest, "INVALID_PARAMETER", err.Error(), nil)
			return
		}
		events[index] = event
		results[index] = midiEventResult{
			Index: index,
			Type:  eventType,
			OK:    true,
		}
	}

	instrument, err := inst.midiInputInstrument()
	if err != nil {
		writeMIDIInputError(w, err)
		return
	}
	if err := instrument.Dispatch(events...); err != nil {
		writeAPIError(w, http.StatusUnprocessableEntity, "MIDI_DISPATCH_FAILED",
			err.Error(), nil)
		return
	}

	writeJSON(w, http.StatusOK, midiEventsResponse{
		Accepted: len(events),
		Results:  results,
	})
}

func makeMIDIEvent(request midiEventRequest, index int) (*lib.MidiEvent, string, error) {
	eventType := strings.ToLower(strings.TrimSpace(request.Type))
	if request.Channel == nil {
		return nil, "", midiParameterError(index, "channel", "is required")
	}
	if *request.Channel < 0 || *request.Channel > 15 {
		return nil, "", midiParameterError(index, "channel", "must be in range 0..15")
	}

	time := float64(-1)
	if request.Time != nil {
		if math.IsNaN(*request.Time) || math.IsInf(*request.Time, 0) {
			return nil, "", midiParameterError(index, "time", "must be finite")
		}
		time = *request.Time
	}

	event := &lib.MidiEvent{
		Time:    lib.AbsolutePosition(time),
		Channel: core.MidiChannelNum(*request.Channel),
	}

	switch eventType {
	case "note_on", "note_off":
		if request.Note == nil {
			return nil, "", midiParameterError(index, "note", "is required for note events")
		}
		if request.CC != nil {
			return nil, "", midiParameterError(index, "cc", "is not valid for note events")
		}
		if request.Note.Number == nil {
			return nil, "", midiParameterError(index, "note.number", "is required")
		}
		if *request.Note.Number < 0 || *request.Note.Number > 127 {
			return nil, "", midiParameterError(index, "note.number", "must be in range 0..127")
		}
		if request.Note.Velocity == nil {
			return nil, "", midiParameterError(index, "note.velocity", "is required")
		}
		if *request.Note.Velocity < 0 || *request.Note.Velocity > 127 {
			return nil, "", midiParameterError(index, "note.velocity", "must be in range 0..127")
		}
		event.Note = &lib.MidiNote{
			Number:   core.MidiNoteNumber(*request.Note.Number),
			Velocity: core.MidiNoteVelocity(*request.Note.Velocity),
		}
		if eventType == "note_on" {
			event.Type = lib.MET_NOTE_ON
		} else {
			event.Type = lib.MET_NOTE_OFF
		}
	case "cc":
		if request.CC == nil {
			return nil, "", midiParameterError(index, "cc", "is required for cc events")
		}
		if request.Note != nil {
			return nil, "", midiParameterError(index, "note", "is not valid for cc events")
		}
		if request.CC.Key == nil {
			return nil, "", midiParameterError(index, "cc.key", "is required")
		}
		if *request.CC.Key < 0 || *request.CC.Key > 127 {
			return nil, "", midiParameterError(index, "cc.key", "must be in range 0..127")
		}
		if request.CC.Value == nil {
			return nil, "", midiParameterError(index, "cc.value", "is required")
		}
		if *request.CC.Value < 0 || *request.CC.Value > 127 {
			return nil, "", midiParameterError(index, "cc.value", "must be in range 0..127")
		}
		event.Type = lib.MET_CC
		event.CC = &lib.MidiCC{
			Key:   *request.CC.Key,
			Value: *request.CC.Value,
		}
	default:
		return nil, "", midiParameterError(index, "type",
			"must be one of note_on, note_off or cc")
	}

	return event, eventType, nil
}

func midiParameterError(index int, field string, reason string) error {
	return fmt.Errorf("events[%d].%s %s", index, field, reason)
}

type midiInputError struct {
	status  int
	code    string
	message string
}

func (err *midiInputError) Error() string {
	return err.message
}

func (inst *innerWebServer) midiInputInstrument() (lib.Instrument, error) {
	if inst.context == nil || inst.context.EngineContext == nil {
		return nil, &midiInputError{
			status:  http.StatusServiceUnavailable,
			code:    "SERVICE_UNAVAILABLE",
			message: "engine context is not available",
		}
	}

	engineContext := inst.context.EngineContext
	if !engineContext.Started {
		return nil, &midiInputError{
			status:  http.StatusConflict,
			code:    "INVALID_STATE",
			message: "engine must be started before dispatching MIDI events",
		}
	}

	if engineContext.InstrumentHolder != nil {
		instrument, err := engineContext.InstrumentHolder.GetInstrument()
		if err != nil {
			return nil, &midiInputError{
				status:  http.StatusServiceUnavailable,
				code:    "SERVICE_UNAVAILABLE",
				message: fmt.Sprintf("MIDI input instrument is unavailable: %v", err),
			}
		}
		if instrument != nil {
			return instrument, nil
		}
	}
	if engineContext.Instrument != nil {
		return engineContext.Instrument, nil
	}
	return nil, &midiInputError{
		status:  http.StatusServiceUnavailable,
		code:    "SERVICE_UNAVAILABLE",
		message: "MIDI input instrument is not available",
	}
}

func writeMIDIInputError(w http.ResponseWriter, err error) {
	inputErr, ok := err.(*midiInputError)
	if !ok {
		writeAPIError(w, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE",
			err.Error(), nil)
		return
	}
	writeAPIError(w, inputErr.status, inputErr.code, inputErr.message, nil)
}
