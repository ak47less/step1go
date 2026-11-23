package core

import (
	"strings"
	"testing"
)

func TestPropertyGetterReadsScalarValues(t *testing.T) {
	props := Properties{
		"bool":   "true",
		"byte":   "42",
		"int":    "-7",
		"string": "hello",
		"uint":   "12",
	}
	getter := props.Getter()

	if got := getter.GetBool("bool"); got != true {
		t.Fatalf("GetBool() = %v, want true", got)
	}
	if got := getter.GetByte("byte"); got != 42 {
		t.Fatalf("GetByte() = %d, want 42", got)
	}
	if got := getter.GetInt("int"); got != -7 {
		t.Fatalf("GetInt() = %d, want -7", got)
	}
	if got := getter.GetString("string"); got != "hello" {
		t.Fatalf("GetString() = %q, want %q", got, "hello")
	}
	if got := getter.GetUint("uint"); got != 12 {
		t.Fatalf("GetUint() = %d, want 12", got)
	}
	if err := getter.Error(); err != nil {
		t.Fatalf("Error() = %v, want nil", err)
	}
}

func TestPropertyGetterMissingValues(t *testing.T) {
	getter := Properties{}.Getter()

	if got := getter.GetString("missing", "fallback"); got != "fallback" {
		t.Fatalf("GetString() = %q, want %q", got, "fallback")
	}
	if got := getter.GetInt("int", 9); got != 9 {
		t.Fatalf("GetInt() = %d, want 9", got)
	}
	if got := getter.GetUint("uint", 10); got != 10 {
		t.Fatalf("GetUint() = %d, want 10", got)
	}
	if got := getter.GetByte("byte", 11); got != 11 {
		t.Fatalf("GetByte() = %d, want 11", got)
	}
	if got := getter.GetBool("bool", true); got != true {
		t.Fatalf("GetBool() = %v, want true", got)
	}

	err := getter.Error()
	if err == nil {
		t.Fatal("Error() = nil, want missing required property errors")
	}
	for _, name := range []string{"missing", "int", "uint", "byte", "bool"} {
		if !strings.Contains(err.Error(), name) {
			t.Fatalf("Error() = %q, want it to contain %q", err, name)
		}
	}
}

func TestPropertyGetterOptionalMissingValues(t *testing.T) {
	getter := Properties{}.Getter().Optional()

	if got := getter.GetString("missing", "fallback"); got != "fallback" {
		t.Fatalf("GetString() = %q, want %q", got, "fallback")
	}
	if got := getter.GetInt("int"); got != 0 {
		t.Fatalf("GetInt() = %d, want 0", got)
	}
	if err := getter.Error(); err != nil {
		t.Fatalf("Error() = %v, want nil for optional missing values", err)
	}
}

func TestPropertyGetterParseErrorsUseDefaults(t *testing.T) {
	props := Properties{
		"bool": "not-a-bool",
		"byte": "300",
		"int":  "not-an-int",
		"uint": "-1",
	}
	getter := props.Getter().Optional()

	if got := getter.GetBool("bool", true); got != true {
		t.Fatalf("GetBool() = %v, want default true", got)
	}
	if got := getter.GetByte("byte", 13); got != 13 {
		t.Fatalf("GetByte() = %d, want default 13", got)
	}
	if got := getter.GetInt("int", 14); got != 14 {
		t.Fatalf("GetInt() = %d, want default 14", got)
	}
	if got := getter.GetUint("uint", 15); got != 15 {
		t.Fatalf("GetUint() = %d, want default 15", got)
	}

	err := getter.Error()
	if err == nil {
		t.Fatal("Error() = nil, want parse errors")
	}
	for _, name := range []string{"bool", "byte", "int", "uint"} {
		if !strings.Contains(err.Error(), name) {
			t.Fatalf("Error() = %q, want it to contain %q", err, name)
		}
	}
}
