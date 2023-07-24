package pressure

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
)

type Pressure struct {
	Avg10  float64
	Avg60  float64
	Avg300 float64
	Total  int64
}

type ResourcePressure struct {
	SomePressure Pressure
	FullPressure *Pressure
}

type CPUPressure struct {
	SomePressure Pressure
}

type IOPressure struct {
	SomePressure Pressure
	FullPressure Pressure
}

type MemoryPressure struct {
	SomePressure Pressure
	FullPressure Pressure
}

type AllPressures struct {
	CPU    CPUPressure
	IO     IOPressure
	Memory MemoryPressure
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

func parsePressureLine(line string) (*Pressure, error) {
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

	return &Pressure{
		Avg10: avg10, Avg60: avg60, Avg300: avg300, Total: total,
	}, nil
}

// cpu pressure is special - 'full' is always zero, and some distributions omit that line entirely.
func parseCPUPressure(s string) (some *Pressure, err error) {
	lines := strings.Split(s, "\n")
	if len(lines) < 2 {
		return nil, fmt.Errorf("expected at least 1 lines in /proc/pressure/cpu contents %q, got %d", s, len(lines))
	}

	some, err = parsePressureLine(lines[0])
	if err != nil {
		return nil, fmt.Errorf("failed to parse 'some' /proc/pressure line %q: %w", lines[0], err)
	}

	return some, nil
}

func parseNonCPUPressure(s string) (some *Pressure, full *Pressure, err error) {
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

// CurrentCPUPressure reads /proc/pressure/cpu and returns a Pressure struct
// representing its output
func CurrentCPUPressure() (some *Pressure, err error) {
	s, err := ioutil.ReadFile("/proc/pressure/cpu")
	if err != nil {
		return nil, fmt.Errorf("failed to read /proc/pressure/cpu: %w", err)
	}
	return parseCPUPressure(string(s))
}

// CurrentIOPressure reads /proc/pressure/io and returns a Pressure struct
// representing its output
func CurrentIOPressure() (some *Pressure, full *Pressure, err error) {
	s, err := ioutil.ReadFile("/proc/pressure/io")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read /proc/pressure/io: %w", err)
	}
	return parseNonCPUPressure(string(s))
}

// CurrentMemoryPressure reads /proc/pressure/memory and returns a Pressure struct
// representing its output
func CurrentMemoryPressure() (some *Pressure, full *Pressure, err error) {
	s, err := ioutil.ReadFile("/proc/pressure/memory")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read /proc/pressure/mem: %w", err)
	}
	return parseNonCPUPressure(string(s))
}

func CurrentAllPressures() (*AllPressures, error) {
	cpuSome, err := CurrentCPUPressure()
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
		CPU:    CPUPressure{SomePressure: *cpuSome},
		Memory: MemoryPressure{SomePressure: *memSome, FullPressure: *memFull},
		IO:     IOPressure{SomePressure: *ioSome, FullPressure: *ioFull},
	}, nil
}
