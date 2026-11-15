package pressure

import (
	"strings"
	"testing"
)

func TestParseCPUPressure(t *testing.T) {
	t.Run("with trailing newline and full line", func(t *testing.T) {
		input := "some avg10=1.25 avg60=0.50 avg300=0.10 total=12345\nfull avg10=0.00 avg60=0.00 avg300=0.00 total=0\n"
		got, err := parseCPUPressure(input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Avg10 != 1.25 || got.Avg60 != 0.50 || got.Avg300 != 0.10 || got.Total != 12345 {
			t.Fatalf("unexpected pressure: %+v", got)
		}
	})

	t.Run("without full line or trailing newline", func(t *testing.T) {
		input := "some avg10=0.01 avg60=0.02 avg300=0.03 total=42"
		got, err := parseCPUPressure(input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Avg10 != 0.01 || got.Avg60 != 0.02 || got.Avg300 != 0.03 || got.Total != 42 {
			t.Fatalf("unexpected pressure: %+v", got)
		}
	})

	t.Run("empty input returns error", func(t *testing.T) {
		if _, err := parseCPUPressure("   \n"); err == nil {
			t.Fatal("expected error for empty input")
		}
	})

	t.Run("missing field returns error", func(t *testing.T) {
		input := "some avg10=0.01 avg60=0.02 total=42\n"
		if _, err := parseCPUPressure(input); err == nil {
			t.Fatal("expected error for missing avg300")
		}
	})
}

func TestParseNonCPUPressure(t *testing.T) {
	t.Run("valid some and full without trailing newline", func(t *testing.T) {
		input := "some avg10=2.00 avg60=1.50 avg300=1.00 total=100\nfull avg10=0.50 avg60=0.25 avg300=0.10 total=20"
		some, full, err := parseNonCPUPressure(input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if some.Avg10 != 2.00 || some.Total != 100 {
			t.Fatalf("unexpected some pressure: %+v", some)
		}
		if full.Avg10 != 0.50 || full.Total != 20 {
			t.Fatalf("unexpected full pressure: %+v", full)
		}
	})

	t.Run("single line returns error", func(t *testing.T) {
		input := "some avg10=2.00 avg60=1.50 avg300=1.00 total=100\n"
		if _, _, err := parseNonCPUPressure(input); err == nil {
			t.Fatal("expected error when full line is missing")
		}
	})

	t.Run("invalid full line mentions full in error", func(t *testing.T) {
		input := "some avg10=2.00 avg60=1.50 avg300=1.00 total=100\nfull avg10=bad avg60=0.25 avg300=0.10 total=20\n"
		_, _, err := parseNonCPUPressure(input)
		if err == nil {
			t.Fatal("expected error for invalid full line")
		}
		if !strings.Contains(err.Error(), "'full'") {
			t.Fatalf("expected error to mention 'full', got: %v", err)
		}
	})
}
