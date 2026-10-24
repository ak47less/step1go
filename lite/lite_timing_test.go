package lite

import "testing"

func TestTimelineEquals(t *testing.T) {
	newTimeline := func() *Timeline {
		return &Timeline{
			Parent: &Timeline{
				Label:  "root",
				Offset: 1,
				Length: 120,
				BPM:    120,
				Tempo:  Tempo{Upper: 4, Lower: 4},
				Loop:   true,
			},
			Label:  "timeline",
			Offset: 2,
			Length: 60,
			BPM:    90,
			Tempo:  Tempo{Upper: 3, Lower: 4},
			Loop:   false,
		}
	}

	tests := []struct {
		name   string
		mutate func(*Timeline)
		want   bool
	}{
		{
			name: "equal values",
			want: true,
		},
		{
			name:   "different parent",
			mutate: func(timeline *Timeline) { timeline.Parent = &Timeline{Label: "other root"} },
			want:   false,
		},
		{
			name:   "different label",
			mutate: func(timeline *Timeline) { timeline.Label = "other timeline" },
			want:   false,
		},
		{
			name:   "different offset",
			mutate: func(timeline *Timeline) { timeline.Offset = 3 },
			want:   false,
		},
		{
			name:   "different length",
			mutate: func(timeline *Timeline) { timeline.Length = 90 },
			want:   false,
		},
		{
			name:   "different bpm",
			mutate: func(timeline *Timeline) { timeline.BPM = 100 },
			want:   false,
		},
		{
			name:   "different tempo",
			mutate: func(timeline *Timeline) { timeline.Tempo = Tempo{Upper: 6, Lower: 8} },
			want:   false,
		},
		{
			name:   "different loop",
			mutate: func(timeline *Timeline) { timeline.Loop = true },
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := newTimeline()
			b := newTimeline()
			if tt.mutate != nil {
				tt.mutate(b)
			}

			if got := a.Equals(b); got != tt.want {
				t.Errorf("Equals() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTimelineEqualsSpecialCases(t *testing.T) {
	timeline := &Timeline{Label: "timeline"}

	if !timeline.Equals(timeline) {
		t.Error("Equals() should return true for the same timeline")
	}
	if timeline.Equals(nil) {
		t.Error("Equals() should return false for a nil timeline")
	}

	var nilTimeline *Timeline
	if nilTimeline.Equals(nil) {
		t.Error("Equals() should return false for two nil timelines")
	}
}

func TestTimelineSame(t *testing.T) {
	timeline := &Timeline{Label: "timeline"}
	equalButDistinct := &Timeline{Label: "timeline"}

	tests := []struct {
		name string
		a    *Timeline
		b    *Timeline
		want bool
	}{
		{
			name: "same instance",
			a:    timeline,
			b:    timeline,
			want: true,
		},
		{
			name: "equal but distinct",
			a:    timeline,
			b:    equalButDistinct,
			want: false,
		},
		{
			name: "two nil timelines",
			want: true,
		},
		{
			name: "one nil timeline",
			a:    timeline,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.a.Same(tt.b); got != tt.want {
				t.Errorf("Same() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTimelineHasSameRoot(t *testing.T) {
	root := &Timeline{Label: "root", BPM: 120}
	equivalentRoot := &Timeline{Label: "root", BPM: 120}
	differentRoot := &Timeline{Label: "root", BPM: 90}

	a := &Timeline{Parent: root, Label: "a"}
	b := &Timeline{Parent: root, Label: "b"}
	c := &Timeline{Parent: equivalentRoot, Label: "c"}
	d := &Timeline{Parent: differentRoot, Label: "d"}

	if !a.HasSameRoot(b) {
		t.Error("HasSameRoot() should return true for branches of the same timeline")
	}
	if a.HasSameRoot(c) {
		t.Error("HasSameRoot() should distinguish equivalent but distinct roots")
	}
	if a.HasSameRoot(d) {
		t.Error("HasSameRoot() should return false for different roots")
	}
}

func TestAbsolutePositionRelative(t *testing.T) {
	root := &Timeline{}
	child := &Timeline{Parent: root, Offset: 2.5}
	grandchild := &Timeline{Parent: child, Offset: 1.25}

	tests := []struct {
		name     string
		absolute AbsolutePosition
		base     *Timeline
		want     Seconds
		wantAt   *Timeline
	}{
		{
			name:     "root timeline without offset",
			absolute: 10,
			base:     root,
			want:     10,
			wantAt:   root,
		},
		{
			name:     "timeline with offset",
			absolute: 10,
			base:     child,
			want:     7.5,
			wantAt:   child,
		},
		{
			name:     "nested timeline offsets",
			absolute: 10,
			base:     grandchild,
			want:     6.25,
			wantAt:   grandchild,
		},
		{
			name:     "negative position",
			absolute: 1,
			base:     grandchild,
			want:     -2.75,
			wantAt:   grandchild,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.absolute.Relative(tt.base, nil)

			if got == nil {
				t.Fatal("Relative() returned nil")
			}
			if got.Position != tt.want {
				t.Errorf("Position = %v, want %v", got.Position, tt.want)
			}
			if got.At != tt.wantAt {
				t.Errorf("At = %p, want %p", got.At, tt.wantAt)
			}
		})
	}
}

func TestAbsolutePositionRelativeReusesDestination(t *testing.T) {
	base := &Timeline{Offset: 2}
	dst := &RelativePosition{
		At:       &Timeline{},
		Position: 99,
	}

	got := AbsolutePosition(10).Relative(base, dst)

	if got != dst {
		t.Fatal("Relative() did not return the supplied destination")
	}
	if got.At != base {
		t.Errorf("At = %p, want %p", got.At, base)
	}
	if got.Position != 8 {
		t.Errorf("Position = %v, want 8", got.Position)
	}
}

func TestAbsolutePositionRelativeWithNilBase(t *testing.T) {
	got := AbsolutePosition(10).Relative(nil, nil)

	if got == nil {
		t.Fatal("Relative() returned nil")
	}
	if got.At == nil {
		t.Fatal("Relative() did not create a base timeline")
	}
	if got.At.GetTotalOffset() != 0 {
		t.Errorf("base offset = %v, want 0", got.At.GetTotalOffset())
	}
	if got.Position != 10 {
		t.Errorf("Position = %v, want 10", got.Position)
	}
}

func TestRelativePositionAbsolute(t *testing.T) {
	root := &Timeline{}
	child := &Timeline{Parent: root, Offset: 2.5}
	grandchild := &Timeline{Parent: child, Offset: 1.25}

	tests := []struct {
		name string
		pos  *RelativePosition
		want AbsolutePosition
	}{
		{
			name: "nil position",
			pos:  nil,
			want: 0,
		},
		{
			name: "root timeline",
			pos:  &RelativePosition{At: root, Position: 10},
			want: 10,
		},
		{
			name: "nil timeline",
			pos:  &RelativePosition{Position: 10},
			want: 10,
		},
		{
			name: "timeline with offset",
			pos:  &RelativePosition{At: child, Position: 7.5},
			want: 10,
		},
		{
			name: "nested timeline offsets",
			pos:  &RelativePosition{At: grandchild, Position: 6.25},
			want: 10,
		},
		{
			name: "negative position",
			pos:  &RelativePosition{At: grandchild, Position: -2.75},
			want: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.pos.Absolute(); got != tt.want {
				t.Errorf("Absolute() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPositionConversionRoundTrip(t *testing.T) {
	root := &Timeline{Offset: 1}
	timeline := &Timeline{Parent: root, Offset: 2.5}
	want := AbsolutePosition(12.75)

	relative := want.Relative(timeline, nil)
	got := relative.Absolute()

	if got != want {
		t.Errorf("round trip = %v, want %v", got, want)
	}
}
