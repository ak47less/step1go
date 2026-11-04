package lite

import (
	"errors"
	"reflect"
	"testing"
)

type lifeManagerTestLife struct {
	cycle *LifeCycle
}

func (inst *lifeManagerTestLife) GetLifeCycle() *LifeCycle {
	return inst.cycle
}

func newLifeManagerTestLife(name string, order int, calls *[]string) Life {
	record := func(stage string) LifeCycleStageFunc {
		return func() error {
			*calls = append(*calls, name+":"+stage)
			return nil
		}
	}

	cycle := &LifeCycle{
		Order:     order,
		OnCreate:  record("create"),
		OnStart:   record("start"),
		OnResume:  record("resume"),
		OnRun:     record("run"),
		OnPause:   record("pause"),
		OnStop:    record("stop"),
		OnDestroy: record("destroy"),
	}
	cycle.OnInit = func(*EngineContext) error {
		*calls = append(*calls, name+":init")
		return nil
	}
	return &lifeManagerTestLife{cycle: cycle}
}

func TestLifeManagerGetMain(t *testing.T) {
	manager := NewLifeManager()

	main := manager.GetMain()
	if main == nil {
		t.Fatal("GetMain() returned nil")
	}
	if got := manager.GetMain(); got != main {
		t.Error("GetMain() returned different instances")
	}

	cycle := main.GetLifeCycle()
	if cycle == nil {
		t.Fatal("GetLifeCycle() returned nil")
	}

	if cycle.OnInit == nil {
		t.Error("OnInit is nil")
	}
	if cycle.OnCreate == nil {
		t.Error("OnCreate is nil")
	}
	if cycle.OnStart == nil {
		t.Error("OnStart is nil")
	}
	if cycle.OnResume == nil {
		t.Error("OnResume is nil")
	}
	if cycle.OnRun == nil {
		t.Error("OnRun is nil")
	}
	if cycle.OnPause == nil {
		t.Error("OnPause is nil")
	}
	if cycle.OnStop == nil {
		t.Error("OnStop is nil")
	}
	if cycle.OnDestroy == nil {
		t.Error("OnDestroy is nil")
	}
}

func TestLifeManagerInvokesLifeCycleInOrder(t *testing.T) {
	manager := NewLifeManager()
	var calls []string

	second := newLifeManagerTestLife("second", 20, &calls)
	first := newLifeManagerTestLife("first", 10, &calls)
	manager.Add(second)
	manager.Add(first)

	cycle := manager.GetMain().GetLifeCycle()
	ec := new(EngineContext)
	stages := []struct {
		name string
		call func() error
	}{
		{name: "init", call: func() error { return cycle.OnInit(ec) }},
		{name: "create", call: cycle.OnCreate},
		{name: "start", call: cycle.OnStart},
		{name: "resume", call: cycle.OnResume},
		{name: "run", call: cycle.OnRun},
		{name: "pause", call: cycle.OnPause},
		{name: "stop", call: cycle.OnStop},
		{name: "destroy", call: cycle.OnDestroy},
	}

	for _, stage := range stages {
		if err := stage.call(); err != nil {
			t.Fatalf("%s stage returned error: %v", stage.name, err)
		}
	}

	want := []string{
		"first:init", "second:init",
		"first:create", "second:create",
		"first:start", "second:start",
		"first:resume", "second:resume",
		"first:run", "second:run",
		"second:pause", "first:pause",
		"second:stop", "first:stop",
		"second:destroy", "first:destroy",
	}
	if !reflect.DeepEqual(calls, want) {
		t.Errorf("life cycle calls = %v, want %v", calls, want)
	}
}

func TestLifeManagerAddInvalidatesCache(t *testing.T) {
	manager := NewLifeManager()
	cycle := manager.GetMain().GetLifeCycle()

	// Build the initially empty cache before adding a life.
	if err := cycle.OnRun(); err != nil {
		t.Fatalf("OnRun() returned error: %v", err)
	}

	var calls []string
	manager.Add(newLifeManagerTestLife("added", 1, &calls))

	cycle = manager.GetMain().GetLifeCycle()
	if err := cycle.OnRun(); err != nil {
		t.Fatalf("OnRun() returned error: %v", err)
	}
	if want := []string{"added:run"}; !reflect.DeepEqual(calls, want) {
		t.Errorf("life cycle calls = %v, want %v", calls, want)
	}
}

func TestLifeManagerRemove(t *testing.T) {
	manager := NewLifeManager()
	var calls []string
	removed := newLifeManagerTestLife("removed", 1, &calls)
	remaining := newLifeManagerTestLife("remaining", 2, &calls)
	manager.Add(removed)
	manager.Add(remaining)

	// Build the cache before removing the life.
	if err := manager.GetMain().GetLifeCycle().OnCreate(); err != nil {
		t.Fatalf("OnCreate() returned error: %v", err)
	}

	manager.Remove(removed)
	manager.Remove(nil)
	manager.Remove(newLifeManagerTestLife("unknown", 3, &calls))

	calls = nil
	if err := manager.GetMain().GetLifeCycle().OnCreate(); err != nil {
		t.Fatalf("OnCreate() returned error: %v", err)
	}
	if want := []string{"remaining:create"}; !reflect.DeepEqual(calls, want) {
		t.Errorf("life cycle calls = %v, want %v", calls, want)
	}
}

func TestLifeManagerIgnoresNilAndEmptyLife(t *testing.T) {
	manager := NewLifeManager()
	manager.Add(nil)
	manager.Add(&lifeManagerTestLife{})

	cycle := manager.GetMain().GetLifeCycle()
	if err := cycle.OnInit(new(EngineContext)); err != nil {
		t.Errorf("OnInit() returned error: %v", err)
	}
	if err := cycle.OnCreate(); err != nil {
		t.Errorf("OnCreate() returned error: %v", err)
	}
	if err := cycle.OnStart(); err != nil {
		t.Errorf("OnStart() returned error: %v", err)
	}
	if err := cycle.OnResume(); err != nil {
		t.Errorf("OnResume() returned error: %v", err)
	}
	if err := cycle.OnRun(); err != nil {
		t.Errorf("OnRun() returned error: %v", err)
	}
	if err := cycle.OnPause(); err != nil {
		t.Errorf("OnPause() returned error: %v", err)
	}
	if err := cycle.OnStop(); err != nil {
		t.Errorf("OnStop() returned error: %v", err)
	}
	if err := cycle.OnDestroy(); err != nil {
		t.Errorf("OnDestroy() returned error: %v", err)
	}
}

func TestLifeManagerPropagatesLifeCycleError(t *testing.T) {
	manager := NewLifeManager()
	wantErr := errors.New("create failed")
	var calls []string

	manager.Add(&lifeManagerTestLife{cycle: &LifeCycle{
		Order: 1,
		OnCreate: func() error {
			calls = append(calls, "failed")
			return wantErr
		},
	}})
	manager.Add(newLifeManagerTestLife("after", 2, &calls))

	err := manager.GetMain().GetLifeCycle().OnCreate()
	if !errors.Is(err, wantErr) {
		t.Fatalf("OnCreate() error = %v, want %v", err, wantErr)
	}
	if want := []string{"failed"}; !reflect.DeepEqual(calls, want) {
		t.Errorf("life cycle calls = %v, want %v", calls, want)
	}
}

func TestLifeManagerInitRequiresEngineContext(t *testing.T) {
	manager := NewLifeManager()
	if err := manager.GetMain().GetLifeCycle().OnInit(nil); err == nil {
		t.Error("OnInit(nil) returned nil error")
	}
}

func TestLifeManagerAddReplaysCurrentLifeCycleState(t *testing.T) {
	tests := []struct {
		name    string
		prepare func(*testing.T, *LifeCycle)
		want    []string
	}{
		{
			name: "not started",
		},
		{
			name: "initialized",
			prepare: func(t *testing.T, cycle *LifeCycle) {
				if err := cycle.OnInit(new(EngineContext)); err != nil {
					t.Fatalf("OnInit() returned error: %v", err)
				}
			},
			want: []string{"added:init"},
		},
		{
			name: "created",
			prepare: func(t *testing.T, cycle *LifeCycle) {
				if err := cycle.OnInit(new(EngineContext)); err != nil {
					t.Fatalf("OnInit() returned error: %v", err)
				}
				if err := cycle.OnCreate(); err != nil {
					t.Fatalf("OnCreate() returned error: %v", err)
				}
			},
			want: []string{"added:init", "added:create"},
		},
		{
			name: "started",
			prepare: func(t *testing.T, cycle *LifeCycle) {
				if err := cycle.OnInit(new(EngineContext)); err != nil {
					t.Fatalf("OnInit() returned error: %v", err)
				}
				if err := cycle.OnCreate(); err != nil {
					t.Fatalf("OnCreate() returned error: %v", err)
				}
				if err := cycle.OnStart(); err != nil {
					t.Fatalf("OnStart() returned error: %v", err)
				}
			},
			want: []string{"added:init", "added:create", "added:start", "added:resume", "added:run"},
		},
		{
			name: "paused",
			prepare: func(t *testing.T, cycle *LifeCycle) {
				if err := cycle.OnInit(new(EngineContext)); err != nil {
					t.Fatalf("OnInit() returned error: %v", err)
				}
				if err := cycle.OnCreate(); err != nil {
					t.Fatalf("OnCreate() returned error: %v", err)
				}
				if err := cycle.OnStart(); err != nil {
					t.Fatalf("OnStart() returned error: %v", err)
				}
				if err := cycle.OnPause(); err != nil {
					t.Fatalf("OnPause() returned error: %v", err)
				}
			},
			want: []string{"added:init", "added:create", "added:start"},
		},
		{
			name: "resumed",
			prepare: func(t *testing.T, cycle *LifeCycle) {
				if err := cycle.OnInit(new(EngineContext)); err != nil {
					t.Fatalf("OnInit() returned error: %v", err)
				}
				if err := cycle.OnCreate(); err != nil {
					t.Fatalf("OnCreate() returned error: %v", err)
				}
				if err := cycle.OnStart(); err != nil {
					t.Fatalf("OnStart() returned error: %v", err)
				}
				if err := cycle.OnPause(); err != nil {
					t.Fatalf("OnPause() returned error: %v", err)
				}
				if err := cycle.OnResume(); err != nil {
					t.Fatalf("OnResume() returned error: %v", err)
				}
			},
			want: []string{"added:init", "added:create", "added:start", "added:resume", "added:run"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewLifeManager()
			cycle := manager.GetMain().GetLifeCycle()
			if tt.prepare != nil {
				tt.prepare(t, cycle)
			}

			var calls []string
			manager.Add(newLifeManagerTestLife("added", 1, &calls))

			if !reflect.DeepEqual(calls, tt.want) {
				t.Errorf("life cycle calls = %v, want %v", calls, tt.want)
			}
		})
	}
}

func TestLifeManagerAddPassesEngineContextToInit(t *testing.T) {
	manager := NewLifeManager()
	ec := new(EngineContext)
	if err := manager.GetMain().GetLifeCycle().OnInit(ec); err != nil {
		t.Fatalf("OnInit() returned error: %v", err)
	}

	var got *EngineContext
	manager.Add(&lifeManagerTestLife{cycle: &LifeCycle{
		OnInit: func(ec *EngineContext) error {
			got = ec
			return nil
		},
	}})

	if got != ec {
		t.Errorf("OnInit() context = %p, want %p", got, ec)
	}
}

func TestLifeManagerAddContinuesAfterCallbackError(t *testing.T) {
	wantErr := errors.New("callback failed")
	var calls []string
	ctx := &innerLifeCycleContext{
		ec:           new(EngineContext),
		onInitDone:   true,
		onCreateDone: true,
		onStartDone:  true,
	}
	caller := &innerItemLifeCaller{context: ctx}
	life := &lifeManagerTestLife{cycle: &LifeCycle{
		OnInit: func(*EngineContext) error {
			calls = append(calls, "init")
			return wantErr
		},
		OnCreate: func() error {
			calls = append(calls, "create")
			return wantErr
		},
		OnStart: func() error {
			calls = append(calls, "start")
			return nil
		},
		OnResume: func() error {
			calls = append(calls, "resume")
			return nil
		},
		OnRun: func() error {
			calls = append(calls, "run")
			return nil
		},
	}}

	caller.call(life)

	want := []string{"init", "create", "start", "resume", "run"}
	if !reflect.DeepEqual(calls, want) {
		t.Errorf("life cycle calls = %v, want %v", calls, want)
	}
}

func TestInnerItemLifeCallerHandlesNilInputs(t *testing.T) {
	var calls []string
	life := newLifeManagerTestLife("added", 1, &calls)
	caller := &innerItemLifeCaller{}

	caller.call(life)
	caller.call(nil)
	caller.call(&lifeManagerTestLife{})

	if len(calls) != 0 {
		t.Errorf("life cycle calls = %v, want no calls", calls)
	}
}
