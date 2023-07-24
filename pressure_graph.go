package main

import (
	"math"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/jamespwilliams/psitop/pressure"
)

func newPressureGraph(somePressures, fullPressures []pressure.Pressure, graphMetric graphMetric) *widgets.Plot {
	somePressures = somePressures[max(len(somePressures)-70, 0):]
	fullPressures = fullPressures[max(len(fullPressures)-70, 0):]

	var someData, fullData []float64
	for _, p := range somePressures {
		switch graphMetric {
		case avg10:
			someData = append(someData, p.Avg10)
		case avg60:
			someData = append(someData, p.Avg60)
		default:
			someData = append(someData, p.Avg300)
		}
	}

	for _, p := range fullPressures {
		switch graphMetric {
		case avg10:
			fullData = append(fullData, p.Avg10)
		case avg60:
			fullData = append(fullData, p.Avg60)
		default:
			fullData = append(fullData, p.Avg300)
		}
	}

	p := widgets.NewPlot()
	p.Title = "Pressure"
	p.Data = [][]float64{}
	p.LineColors = []ui.Color{}

	if len(someData) > 0 {
		p.Data = append(p.Data, someData)
		p.LineColors = append(p.LineColors, ui.ColorGreen)
	}

	if len(fullData) > 0 {
		p.Data = append(p.Data, fullData)
		p.LineColors = append(p.LineColors, ui.ColorYellow)
	}

	p.MaxVal = calculatePressureGraphMaxVal(p.Data)

	return p
}

func calculatePressureGraphMaxVal(data [][]float64) float64 {
	var max float64
	for _, line := range data {
		for _, point := range line {
			if point > max {
				max = point
			}
		}
	}

	maxVal := math.Pow(2, math.Ceil(math.Log2(max)))
	if maxVal < 1.0 {
		return 1.0
	}
	return maxVal
}
