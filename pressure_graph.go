package main

import (
	"math"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/jamespwilliams/psitop/pressure"
)

func newPressureGraph(pressures []pressure.ResourcePressure, pressureType pressureType, graphMetric graphMetric) *widgets.Plot {
	pressures = pressures[max(len(pressures)-70, 0):]

	var someData, fullData []float64
	for _, p := range pressures {
		switch graphMetric {
		case avg10:
			someData = append(someData, p.SomePressure.Avg10)
			fullData = append(fullData, p.FullPressure.Avg10)
		case avg60:
			someData = append(someData, p.SomePressure.Avg60)
			fullData = append(fullData, p.FullPressure.Avg60)
		default:
			someData = append(someData, p.SomePressure.Avg300)
			fullData = append(fullData, p.FullPressure.Avg300)
		}
	}

	p := widgets.NewPlot()
	p.Title = "Pressure"

	switch pressureType {
	case some:
		p.Data = [][]float64{someData}
		p.LineColors = []ui.Color{ui.ColorGreen}
	case full:
		p.Data = [][]float64{fullData}
		p.LineColors = []ui.Color{ui.ColorYellow}
	default:
		p.Data = [][]float64{someData, fullData}
		p.LineColors = []ui.Color{ui.ColorGreen, ui.ColorYellow}
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
