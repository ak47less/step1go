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
