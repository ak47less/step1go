package restful

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/ak47less/step1go/lite2/lib"
	"github.com/ak47less/step1go/lite2/webaudioapi"
)

type fakeMIDIInstrument struct {
	events []*lib.MidiEvent
	err    error
}

func (inst *fakeMIDIInstrument) Dispatch(events ...*lib.MidiEvent) error {
	inst.events = append(inst.events, events...)
	return inst.err
}

func (inst *fakeMIDIInstrument) GetOutput() webaudioapi.AudioNode {
	return nil
}

func TestMIDIEvents(t *testing.T) {
	web, _ := newTestWebServer(t)
	instrument := new(fakeMIDIInstrument)
	web.context.EngineContext.Instrument = instrument
	handler := web.handler()

	body := `{
		"events": [
			{
				"type": "note_on",
				"time": -1,
				"channel": 0,
				"note": {"number": 60, "velocity": 100}
			},
			{
				"type": "note_off",
				"time": 0.5,
				"channel": 0,
				"note": {"number": 60, "velocity": 0}
			},
			{
				"type": "cc",
				"channel": 1,
				"cc": {"key": 74, "value": 127}
			}
		]
	}`
	response := performRequest(handler, http.MethodPost, "/api/v1/midi/events", body)
	if response.Code != http.StatusOK {
		t.Fatalf("MIDI events status = %d, body = %s", response.Code, response.Body.String())
	}

	var result midiEventsResponse
	if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
		t.Fatalf("decode MIDI response: %v", err)
	}
	if result.Accepted != 3 || len(result.Results) != 3 {
		t.Fatalf("unexpected MIDI response: %+v", result)
	}
	for index, item := range result.Results {
		if item.Index != index || !item.OK {
			t.Fatalf("unexpected result %d: %+v", index, item)
		}
	}

	if len(instrument.events) != 3 {
		t.Fatalf("dispatched event count = %d, want 3", len(instrument.events))
	}
	if event := instrument.events[0]; event.Type != lib.MET_NOTE_ON ||
		event.Time != -1 || event.Channel != 0 || event.Note == nil ||
		event.Note.Number != 60 || event.Note.Velocity != 100 {
		t.Fatalf("unexpected note_on event: %+v", event)
	}
	if event := instrument.events[1]; event.Type != lib.MET_NOTE_OFF ||
		event.Time != 0.5 || event.Channel != 0 || event.Note == nil ||
		event.Note.Number != 60 || event.Note.Velocity != 0 {
		t.Fatalf("unexpected note_off event: %+v", event)
	}
	if event := instrument.events[2]; event.Type != lib.MET_CC ||
		event.Time != -1 || event.Channel != 1 || event.CC == nil ||
		event.CC.Key != 74 || event.CC.Value != 127 {
		t.Fatalf("unexpected cc event: %+v", event)
	}
}

func TestMIDIEventsValidation(t *testing.T) {
	web, _ := newTestWebServer(t)
	instrument := new(fakeMIDIInstrument)
	web.context.EngineContext.Instrument = instrument
	handler := web.handler()

	tests := []struct {
		name   string
		body   string
		code   string
		status int
	}{
		{
			name:   "empty",
			body:   `{"events":[]}`,
			code:   "INVALID_PARAMETER",
			status: http.StatusBadRequest,
		},
		{
			name:   "missing channel",
			body:   `{"events":[{"type":"note_on","note":{"number":60,"velocity":100}}]}`,
			code:   "INVALID_PARAMETER",
			status: http.StatusBadRequest,
		},
		{
			name:   "channel out of range",
			body:   `{"events":[{"type":"note_on","channel":16,"note":{"number":60,"velocity":100}}]}`,
			code:   "INVALID_PARAMETER",
			status: http.StatusBadRequest,
		},
		{
			name:   "unknown type",
			body:   `{"events":[{"type":"pitch_bend","channel":0}]}`,
			code:   "INVALID_PARAMETER",
			status: http.StatusBadRequest,
		},
		{
			name:   "missing note",
			body:   `{"events":[{"type":"note_on","channel":0}]}`,
			code:   "INVALID_PARAMETER",
			status: http.StatusBadRequest,
		},
		{
			name:   "velocity out of range",
			body:   `{"events":[{"type":"note_on","channel":0,"note":{"number":60,"velocity":128}}]}`,
			code:   "INVALID_PARAMETER",
			status: http.StatusBadRequest,
		},
		{
			name:   "note payload on cc",
			body:   `{"events":[{"type":"cc","channel":0,"note":{"number":60,"velocity":100}}]}`,
			code:   "INVALID_PARAMETER",
			status: http.StatusBadRequest,
		},
		{
			name:   "missing cc value",
			body:   `{"events":[{"type":"cc","channel":0,"cc":{"key":74}}]}`,
			code:   "INVALID_PARAMETER",
			status: http.StatusBadRequest,
		},
		{
			name:   "unknown field",
			body:   `{"events":[{"type":"note_on","channel":0,"note":{"number":60,"velocity":100},"value":1}]}`,
			code:   "INVALID_JSON",
			status: http.StatusBadRequest,
		},
		{
			name:   "invalid later event",
			body:   `{"events":[{"type":"note_on","channel":0,"note":{"number":60,"velocity":100}},{"type":"note_off","channel":0}]}`,
			code:   "INVALID_PARAMETER",
			status: http.StatusBadRequest,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response := performRequest(handler, http.MethodPost, "/api/v1/midi/events", test.body)
			if response.Code != test.status {
				t.Fatalf("status = %d, want %d, body = %s", response.Code, test.status, response.Body.String())
			}
			if code := decodeErrorCode(t, response); code != test.code {
				t.Fatalf("error code = %q, want %q", code, test.code)
			}
		})
	}
	if len(instrument.events) != 0 {
		t.Fatalf("invalid batches dispatched %d events, want 0", len(instrument.events))
	}
}

func TestMIDIEventsErrors(t *testing.T) {
	t.Run("method not allowed", func(t *testing.T) {
		web, _ := newTestWebServer(t)
		response := performRequest(web.handler(), http.MethodGet, "/api/v1/midi/events", "")
		if response.Code != http.StatusMethodNotAllowed {
			t.Fatalf("status = %d, want %d", response.Code, http.StatusMethodNotAllowed)
		}
	})

	t.Run("engine not started", func(t *testing.T) {
		web, _ := newTestWebServer(t)
		web.context.EngineContext.Started = false
		web.context.EngineContext.Instrument = new(fakeMIDIInstrument)
		response := performRequest(web.handler(), http.MethodPost, "/api/v1/midi/events",
			`{"events":[{"type":"note_on","channel":0,"note":{"number":60,"velocity":100}}]}`)
		if response.Code != http.StatusConflict {
			t.Fatalf("status = %d, want %d", response.Code, http.StatusConflict)
		}
		if code := decodeErrorCode(t, response); code != "INVALID_STATE" {
			t.Fatalf("error code = %q, want INVALID_STATE", code)
		}
	})

	t.Run("instrument unavailable", func(t *testing.T) {
		web, _ := newTestWebServer(t)
		response := performRequest(web.handler(), http.MethodPost, "/api/v1/midi/events",
			`{"events":[{"type":"note_on","channel":0,"note":{"number":60,"velocity":100}}]}`)
		if response.Code != http.StatusServiceUnavailable {
			t.Fatalf("status = %d, want %d", response.Code, http.StatusServiceUnavailable)
		}
	})

	t.Run("dispatch failed", func(t *testing.T) {
		web, _ := newTestWebServer(t)
		web.context.EngineContext.Instrument = &fakeMIDIInstrument{
			err: errors.New("instrument rejected event"),
		}
		response := performRequest(web.handler(), http.MethodPost, "/api/v1/midi/events",
			`{"events":[{"type":"note_on","channel":0,"note":{"number":60,"velocity":100}}]}`)
		if response.Code != http.StatusUnprocessableEntity {
			t.Fatalf("status = %d, want %d", response.Code, http.StatusUnprocessableEntity)
		}
		if code := decodeErrorCode(t, response); code != "MIDI_DISPATCH_FAILED" {
			t.Fatalf("error code = %q, want MIDI_DISPATCH_FAILED", code)
		}
	})
}
