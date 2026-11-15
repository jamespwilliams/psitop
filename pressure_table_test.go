package main

import (
	"strings"
	"testing"

	"github.com/jamespwilliams/psitop/pressure"
)

func TestNewPressureTableComparesSecondSampleWithFirst(t *testing.T) {
	samples := []pressure.Pressure{
		{Avg10: 1.00, Avg60: 2.00, Avg300: 3.00},
		{Avg10: 1.00, Avg60: 2.50, Avg300: 3.00},
	}

	table := newPressureTable("CPU", samples, nil)
	if len(table.Rows) != 2 {
		t.Fatalf("expected 2 rows (header + some), got %d", len(table.Rows))
	}

	row := table.Rows[1]
	// avg10 unchanged from sample[0] -> not bold
	if strings.Contains(row[1], "mod:bold") {
		t.Fatalf("expected unchanged avg10 not to be bold, got %q", row[1])
	}
	// avg60 changed from sample[0] -> bold
	if !strings.Contains(row[2], "mod:bold") {
		t.Fatalf("expected changed avg60 to be bold, got %q", row[2])
	}
}

func TestCalculatePressureGraphMaxVal(t *testing.T) {
	tests := []struct {
		name string
		data [][]float64
		want float64
	}{
		{name: "empty", data: nil, want: 1.0},
		{name: "all zeros", data: [][]float64{{0, 0}}, want: 1.0},
		{name: "sub-unit", data: [][]float64{{0.3, 0.7}}, want: 1.0},
		{name: "power of two", data: [][]float64{{4.0}}, want: 4.0},
		{name: "rounds up to power of two", data: [][]float64{{2.5, 5.1}}, want: 8.0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := calculatePressureGraphMaxVal(tc.data); got != tc.want {
				t.Fatalf("calculatePressureGraphMaxVal() = %v, want %v", got, tc.want)
			}
		})
	}
}
