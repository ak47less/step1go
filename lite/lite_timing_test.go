package lite

import "testing"

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
