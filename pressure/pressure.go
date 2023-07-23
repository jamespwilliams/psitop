package pressure

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
)

type PressureSnapshot struct {
	Avg10  float64
	Avg60  float64
	Avg300 float64
	Total  int64
}

type ResourcePressure struct {
	SomePressure PressureSnapshot
	FullPressure PressureSnapshot
}

type AllPressures struct {
	CPU    ResourcePressure
	IO     ResourcePressure
	Memory ResourcePressure
}

// parseSpaceSeparatedKeyValues parses a set of space-separated key=value pairs
// into a map
func parseSpaceSeparatedKeyValues(kvs string) (map[string]string, error) {
	m := make(map[string]string)
	for _, kv := range strings.Split(kvs, " ") {
		key, val, found := strings.Cut(kv, "=")
		if !found {
			return nil, fmt.Errorf("field %q didn't conform to expected key=value format", kv)
		}

		m[key] = val
	}
	return m, nil
}

func parsePressureLine(line string) (*PressureSnapshot, error) {
	// chop off the initial some/full tag:
	_, line, _ = strings.Cut(line, " ")

	fields, err := parseSpaceSeparatedKeyValues(line)
	if err != nil {
		return nil, fmt.Errorf("failed to parse /proc/pressure line %q: %w", line, err)
	}

	for _, expectedKey := range []string{"avg10", "avg60", "avg300", "total"} {
		if _, ok := fields[expectedKey]; !ok {
			return nil, fmt.Errorf("/proc/pressure line %q is missing expected field %q", line, expectedKey)
		}
	}

	avg10, err := strconv.ParseFloat(fields["avg10"], 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse avg10 value %q as a float: %w", fields["avg10"], err)
	}

	avg60, err := strconv.ParseFloat(fields["avg60"], 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse avg60 value %q as a float: %w", fields["avg60"], err)
	}

	avg300, err := strconv.ParseFloat(fields["avg300"], 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse avg300 value %q as a float: %w", fields["avg300"], err)
	}

	total, err := strconv.ParseInt(fields["total"], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse total value %q as an int: %w", fields["total"], err)
	}

	return &PressureSnapshot{
		Avg10: avg10, Avg60: avg60, Avg300: avg300, Total: total,
	}, nil
}

func ParseProcPressure(s string) (some *PressureSnapshot, full *PressureSnapshot, err error) {
	lines := strings.Split(s, "\n")
	if len(lines) < 2 {
		return nil, nil, fmt.Errorf("expected at least 2 lines in /proc/pressure contents %q, got %d", s, len(lines))
	}

	some, err = parsePressureLine(lines[0])
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse 'some' /proc/pressure line %q: %w", lines[0], err)
	}

	full, err = parsePressureLine(lines[1])
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse 'some' /proc/pressure line %q: %w", lines[1], err)
	}

	return some, full, nil
}

func currentPressure(kind string) (some *PressureSnapshot, full *PressureSnapshot, err error) {
	filename := fmt.Sprintf("/proc/pressure/%s", kind)

	s, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read %q: %w", filename, err)
	}

	return ParseProcPressure(string(s))
}

// CurrentCPUPressure reads /proc/pressure/cpu and returns a Pressure struct
// representing its output
func CurrentCPUPressure() (some *PressureSnapshot, full *PressureSnapshot, err error) {
	return currentPressure("cpu")
}

// CurrentIOPressure reads /proc/pressure/io and returns a Pressure struct
// representing its output
func CurrentIOPressure() (some *PressureSnapshot, full *PressureSnapshot, err error) {
	return currentPressure("io")
}

// CurrentMemoryPressure reads /proc/pressure/memory and returns a Pressure struct
// representing its output
func CurrentMemoryPressure() (some *PressureSnapshot, full *PressureSnapshot, err error) {
	return currentPressure("memory")
}

func CurrentAllPressures() (*AllPressures, error) {
	cpuSome, cpuFull, err := CurrentCPUPressure()
	if err != nil {
		return nil, fmt.Errorf("failed to read cpu pressure: %w", err)
	}

	memSome, memFull, err := CurrentMemoryPressure()
	if err != nil {
		return nil, fmt.Errorf("failed to read memory pressure: %w", err)
	}

	ioSome, ioFull, err := CurrentIOPressure()
	if err != nil {
		return nil, fmt.Errorf("failed to read io pressure: %w", err)
	}

	return &AllPressures{
		CPU:    ResourcePressure{SomePressure: *cpuSome, FullPressure: *cpuFull},
		Memory: ResourcePressure{SomePressure: *memSome, FullPressure: *memFull},
		IO:     ResourcePressure{SomePressure: *ioSome, FullPressure: *ioFull},
	}, nil
}
