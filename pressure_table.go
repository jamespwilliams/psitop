package main

import (
	"fmt"
	"strings"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/jamespwilliams/psitop/pressure"
)

func newPressureTable(title string, somePressures, fullPressures []pressure.Pressure) *widgets.Table {
	table := widgets.NewTable()
	table.Title = title
	table.TextStyle = ui.NewStyle(ui.ColorWhite)
	table.RowSeparator = true
	table.RowStyles = map[int]ui.Style{0: ui.NewStyle(ui.ColorWhite, ui.ColorBlack, ui.ModifierBold)}
	table.FillRow = true
	table.TextAlignment = ui.AlignCenter

	table.Rows = [][]string{
		{"", "avg10", "avg60", "avg300"},
	}

	allPressures := map[string][]pressure.Pressure{"some": somePressures, "full": fullPressures}
	for _, pressureType := range []string{"some", "full"} {
		pressures := allPressures[pressureType]

		if len(pressures) == 0 {
			continue
		}

		current := pressures[len(pressures)-1]
		var previous *pressure.Pressure
		if len(pressures) > 2 {
			previous = &pressures[len(pressures)-2]
		}

		table.Rows = append(table.Rows, getPressureRow(fmt.Sprintf("[%s](mod:bold)", pressureType), current, previous))
	}

	return table
}

func getPressureRow(title string, current pressure.Pressure, previous *pressure.Pressure) []string {
	row := []string{title}

	previousAvg10, previousAvg60, previousAvg300 := 0.0, 0.0, 0.0
	if previous != nil {
		previousAvg10, previousAvg60, previousAvg300 = previous.Avg10, previous.Avg60, previous.Avg300
	}

	row = append(row, formatLoadAverage(current.Avg10, previousAvg10))
	row = append(row, formatLoadAverage(current.Avg60, previousAvg60))
	row = append(row, formatLoadAverage(current.Avg300, previousAvg300))

	return row
}

// formatLoadAverage accepts a pair of load averages, one being the current load average, and the other
// being the previous load average, and formats them, ready for insertion into a widgets.Table.
//
// This includes applying colors and modifiers to the text (for example, making load averages that have
// changed since the last render bold).
func formatLoadAverage(currentMetric, previousMetric float64) string {
	var styles []string

	if currentMetric != previousMetric {
		styles = append(styles, "mod:bold")
	}

	if currentMetric == 0.0 {
		styles = append(styles, "fg:white")
	} else if currentMetric < float64(numCPUs) {
		styles = append(styles, "fg:green")
	} else if currentMetric < 2*float64(numCPUs) {
		styles = append(styles, "fg:yellow")
	} else {
		styles = append(styles, "fg:red")
	}

	styleSpecifier := strings.Join(styles, ",")

	return fmt.Sprintf("[%.2f](%s)", currentMetric, styleSpecifier)
}
