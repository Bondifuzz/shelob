package generateInput

import (
	"encoding/base64"
	"testing"
)

func TestCheckStringFormatByteHasBoundedPositiveLength(t *testing.T) {
	value, ok := CheckStringFormat("byte").(string)
	if !ok {
		t.Fatalf("CheckStringFormat(byte) returned %T, want string", value)
	}

	decoded, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		t.Fatalf("byte format returned invalid base64: %v", err)
	}
	if len(decoded) == 0 || len(decoded) > 1024 {
		t.Fatalf("decoded byte length = %d, want 1..1024", len(decoded))
	}
}

func TestCheckStringFormatDefaultHasBoundedPositiveLength(t *testing.T) {
	value, ok := CheckStringFormat("unknown-format").(string)
	if !ok {
		t.Fatalf("CheckStringFormat(default) returned %T, want string", value)
	}
	if len(value) == 0 || len(value) > 1024 {
		t.Fatalf("default string length = %d, want 1..1024", len(value))
	}
}

func TestCheckIntegerFormatDefaultIsBounded(t *testing.T) {
	value, ok := CheckIntegerFormat("").(int)
	if !ok {
		t.Fatalf("CheckIntegerFormat(default) returned %T, want int", value)
	}
	if value < -1_000_000 || value > 1_000_000 {
		t.Fatalf("integer = %d, want within [-1000000, 1000000]", value)
	}
}
