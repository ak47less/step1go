package lite

import (
	"errors"
	"sync"
	"testing"
)

const (
	eventTestType EventType = EventTypeMax + 1
	eventTestID   EventID   = 100
)

func TestNewEventContext(t *testing.T) {
	ctx := NewEventContext()
	if ctx == nil {
		t.Fatal("NewEventContext() returned nil")
	}
	if ctx.GetDispatcher() == nil {
		t.Fatal("GetDispatcher() returned nil")
	}
	if ctx.GetRunner() == nil {
		t.Fatal("GetRunner() returned nil")
	}

	dispatcher, ok := ctx.dispatcher.(*innerEventDispatcher)
	if !ok {
		t.Fatalf("dispatcher type = %T, want *innerEventDispatcher", ctx.dispatcher)
	}
	if dispatcher.context != ctx {
		t.Error("dispatcher is not associated with the event context")
	}

	runner, ok := ctx.runner.(*innerEventRunner)
	if !ok {
		t.Fatalf("runner type = %T, want *innerEventRunner", ctx.runner)
	}
	if runner.context != ctx {
		t.Error("runner is not associated with the event context")
	}

	zero := new(EventContext)
	if zero.GetDispatcher() != nil {
		t.Error("GetDispatcher() on a zero context returned non-nil")
	}
	if zero.GetRunner() != nil {
		t.Error("GetRunner() on a zero context returned non-nil")
	}
}

func TestEventDispatcherQueuesAndRunnerDispatchesMatchingCallbacks(t *testing.T) {
	ctx := NewEventContext()
	dispatcher := ctx.GetDispatcher()
	runner := ctx.GetRunner()
	got := make([]string, 0)

	callbacks := []*EventCallback{
		{
			Type: eventTestType,
			ID:   eventTestID,
			Fn: func(event *Event) error {
				got = append(got, event.Sender.(string))
				return nil
			},
		},
		{
			Type: eventTestType + 1,
			ID:   eventTestID,
			Fn: func(*Event) error {
				t.Error("callback for a different type was invoked")
				return nil
			},
		},
		{
			Type: eventTestType,
			ID:   eventTestID + 1,
			Fn: func(*Event) error {
				t.Error("callback for a different id was invoked")
				return nil
			},
		},
	}
	for _, callback := range callbacks {
		if err := dispatcher.AddCallback(callback); err != nil {
			t.Fatalf("AddCallback() returned error: %v", err)
		}
	}

	if err := dispatcher.Dispatch(Event{
		Type:   eventTestType,
		ID:     eventTestID,
		Sender: "first",
	}); err != nil {
		t.Fatalf("Dispatch(first) returned error: %v", err)
	}
	if err := dispatcher.Dispatch(Event{
		Type:   eventTestType,
		ID:     eventTestID,
		Sender: "second",
	}); err != nil {
		t.Fatalf("Dispatch(second) returned error: %v", err)
	}
	if len(ctx.buffer) != 2 {
		t.Fatalf("buffer length before Run() = %d, want 2", len(ctx.buffer))
	}
	if len(got) != 0 {
		t.Fatalf("callbacks ran during Dispatch(): %v", got)
	}

	if err := runner.Run(); err != nil {
		t.Fatalf("Run() returned error: %v", err)
	}
	want := []string{"first", "second"}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Errorf("callback events = %v, want %v", got, want)
	}
	if len(ctx.buffer) != 0 {
		t.Errorf("buffer length after Run() = %d, want 0", len(ctx.buffer))
	}
}

func TestEventRunnerPropagatesCallbackErrorAndStops(t *testing.T) {
	ctx := NewEventContext()
	dispatcher := ctx.GetDispatcher()
	runner := ctx.GetRunner()
	wantErr := errors.New("callback failed")
	laterCalled := false

	if err := dispatcher.AddCallback(&EventCallback{
		Type: eventTestType,
		ID:   eventTestID,
		Fn: func(*Event) error {
			return wantErr
		},
	}); err != nil {
		t.Fatalf("AddCallback(first) returned error: %v", err)
	}
	if err := dispatcher.AddCallback(&EventCallback{
		Type: eventTestType,
		ID:   eventTestID,
		Fn: func(*Event) error {
			laterCalled = true
			return nil
		},
	}); err != nil {
		t.Fatalf("AddCallback(second) returned error: %v", err)
	}

	for _, sender := range []string{"first", "second"} {
		if err := dispatcher.Dispatch(Event{
			Type:   eventTestType,
			ID:     eventTestID,
			Sender: sender,
		}); err != nil {
			t.Fatalf("Dispatch(%q) returned error: %v", sender, err)
		}
	}

	if err := runner.Run(); !errors.Is(err, wantErr) {
		t.Fatalf("Run() error = %v, want %v", err, wantErr)
	}
	if laterCalled {
		t.Error("a callback after the failed callback was invoked")
	}
	if len(ctx.buffer) != 1 {
		t.Errorf("buffer length after failed Run() = %d, want 1", len(ctx.buffer))
	}
}

func TestEventCallbackCanDispatchAnotherEvent(t *testing.T) {
	ctx := NewEventContext()
	dispatcher := ctx.GetDispatcher()
	got := make([]EventID, 0)

	if err := dispatcher.AddCallback(&EventCallback{
		Type: eventTestType,
		ID:   eventTestID,
		Fn: func(event *Event) error {
			got = append(got, event.ID)
			if event.ID == eventTestID {
				return dispatcher.Dispatch(Event{
					Type: eventTestType,
					ID:   eventTestID + 1,
				})
			}
			return nil
		},
	}); err != nil {
		t.Fatalf("AddCallback(first) returned error: %v", err)
	}
	if err := dispatcher.AddCallback(&EventCallback{
		Type: eventTestType,
		ID:   eventTestID + 1,
		Fn: func(event *Event) error {
			got = append(got, event.ID)
			return nil
		},
	}); err != nil {
		t.Fatalf("AddCallback(second) returned error: %v", err)
	}

	if err := dispatcher.Dispatch(Event{Type: eventTestType, ID: eventTestID}); err != nil {
		t.Fatalf("Dispatch() returned error: %v", err)
	}
	if err := ctx.GetRunner().Run(); err != nil {
		t.Fatalf("Run() returned error: %v", err)
	}
	if len(got) != 2 || got[0] != eventTestID || got[1] != eventTestID+1 {
		t.Errorf("callback ids = %v, want [%d %d]", got, eventTestID, eventTestID+1)
	}
}

func TestEventDispatcherAddAndRemoveCallback(t *testing.T) {
	ctx := NewEventContext()
	dispatcher := ctx.GetDispatcher()

	if err := dispatcher.AddCallback(nil); err == nil {
		t.Error("AddCallback(nil) returned nil error")
	}
	if err := dispatcher.AddCallback(&EventCallback{}); err == nil {
		t.Error("AddCallback() accepted a nil function")
	}
	if err := dispatcher.RemoveCallback(nil); err == nil {
		t.Error("RemoveCallback(nil) returned nil error")
	}

	called := 0
	callback := &EventCallback{
		Type: eventTestType,
		ID:   eventTestID,
		Fn: func(*Event) error {
			called++
			return nil
		},
	}
	if err := dispatcher.AddCallback(callback); err != nil {
		t.Fatalf("AddCallback() returned error: %v", err)
	}
	if err := dispatcher.RemoveCallback(callback); err != nil {
		t.Fatalf("RemoveCallback() returned error: %v", err)
	}
	if err := dispatcher.RemoveCallback(callback); err == nil {
		t.Error("RemoveCallback() removed an already removed callback")
	}

	if err := dispatcher.Dispatch(Event{Type: eventTestType, ID: eventTestID}); err != nil {
		t.Fatalf("Dispatch() returned error: %v", err)
	}
	if err := ctx.GetRunner().Run(); err != nil {
		t.Fatalf("Run() returned error: %v", err)
	}
	if called != 0 {
		t.Errorf("removed callback was invoked %d times", called)
	}
}

func TestEventRunnerRunContext(t *testing.T) {
	first := NewEventContext()
	second := NewEventContext()
	called := false

	if err := second.GetDispatcher().AddCallback(&EventCallback{
		Type: eventTestType,
		ID:   eventTestID,
		Fn: func(*Event) error {
			called = true
			return nil
		},
	}); err != nil {
		t.Fatalf("AddCallback() returned error: %v", err)
	}
	if err := second.GetDispatcher().Dispatch(Event{
		Type: eventTestType,
		ID:   eventTestID,
	}); err != nil {
		t.Fatalf("Dispatch() returned error: %v", err)
	}

	if err := first.GetRunner().RunContext(second); err != nil {
		t.Fatalf("RunContext() returned error: %v", err)
	}
	if !called {
		t.Error("RunContext() did not process the supplied context")
	}
	if len(second.buffer) != 0 {
		t.Errorf("supplied context buffer length = %d, want 0", len(second.buffer))
	}
	if err := first.GetRunner().RunContext(nil); err == nil {
		t.Error("RunContext(nil) returned nil error")
	}
}

func TestEventContextConcurrentDispatch(t *testing.T) {
	ctx := NewEventContext()
	const eventCount = 100
	var callbackCount int
	var mutex sync.Mutex

	if err := ctx.GetDispatcher().AddCallback(&EventCallback{
		Type: eventTestType,
		ID:   eventTestID,
		Fn: func(*Event) error {
			mutex.Lock()
			callbackCount++
			mutex.Unlock()
			return nil
		},
	}); err != nil {
		t.Fatalf("AddCallback() returned error: %v", err)
	}

	var waitGroup sync.WaitGroup
	for index := 0; index < eventCount; index++ {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			if err := ctx.GetDispatcher().Dispatch(Event{
				Type: eventTestType,
				ID:   eventTestID,
			}); err != nil {
				t.Errorf("Dispatch() returned error: %v", err)
			}
		}()
	}
	waitGroup.Wait()

	if err := ctx.GetRunner().Run(); err != nil {
		t.Fatalf("Run() returned error: %v", err)
	}
	if callbackCount != eventCount {
		t.Errorf("callback count = %d, want %d", callbackCount, eventCount)
	}
}
