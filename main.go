package main

import (
	"log"
	"runtime"
	"sync"
	"time"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/jamespwilliams/psitop/pressure"
)

var numCPUs = runtime.NumCPU()

const (
	renderPeriod  = 1000 * time.Millisecond
	fetchPeriod   = 500 * time.Millisecond
	maxDataLength = 100
)

func main() {
	var (
		resource     resource     = resourceAll
		graphMetric  graphMetric  = avg10
		pressureType pressureType = both

		dataMux sync.Mutex
		data    []*pressure.AllPressures
	)

	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	fetchPressures(&dataMux, &data)
	go func() {
		for {
			time.Sleep(fetchPeriod)
			fetchPressures(&dataMux, &data)
		}
	}()

	uiEvents := ui.PollEvents()
	for {
		dataMux.Lock()
		ui.Clear()
		render(resource, graphMetric, pressureType, data)
		dataMux.Unlock()

		select {
		case <-time.After(renderPeriod):
		case e := <-uiEvents:
			switch e.ID {
			case "q", "<C-c>":
				return
			case "a":
				resource = resourceAll
			case "c":
				resource = resourceCPU
			case "m":
				resource = resourceMemory
			case "i":
				resource = resourceIO
			case "1":
				graphMetric = avg10
			case "6":
				graphMetric = avg60
			case "3":
				graphMetric = avg300
			case "s":
				pressureType = some
			case "f":
				pressureType = full
			case "b":
				pressureType = both
			}
		}
	}
}

func fetchPressures(mux *sync.Mutex, data *[]*pressure.AllPressures) {
	mux.Lock()
	defer mux.Unlock()

	pressures, err := pressure.CurrentAllPressures()
	if err != nil {
		log.Fatalf(err.Error())
	}
	*data = append(*data, pressures)

	if len(*data) > maxDataLength {
		*data = (*data)[1:]
	}
}

func render(resource resource, graphMetric graphMetric, pressureType pressureType, data []*pressure.AllPressures) {
	selectors := renderSelectors(resource, pressureType, graphMetric)

	var pane []ui.Drawable
	if resource == resourceAll {
		pane = renderAllPane(data, pressureType)
	} else {
		pane = renderSingleResourcePane(data, resource, pressureType, graphMetric)
	}

	var items []ui.Drawable
	items = append(items, selectors...)
	items = append(items, pane...)
	ui.Render(items...)
}

func renderSelectors(resource resource, pressureType pressureType, graphMetric graphMetric) []ui.Drawable {
	resourceSelector := widgets.NewTabPane("[a]ll", "[c]pu", "[m]emory", "[i]o")
	resourceSelector.PaddingLeft = 1
	resourceSelector.Title = "Resource"
	resourceSelector.TitleStyle = ui.NewStyle(ui.ColorWhite, ui.ColorBlack, ui.ModifierBold)
	resourceSelector.SetRect(1, 1, 36, 4)
	resourceSelector.Border = true
	resourceSelector.ActiveTabIndex = resource.tabIndex()
	resourceSelector.ActiveTabStyle.Modifier = ui.ModifierBold

	pressureTypeSelector := widgets.NewTabPane("[s]ome", "[f]ull", "[b]oth")
	pressureTypeSelector.PaddingLeft = 1
	pressureTypeSelector.Title = "Some/full"
	pressureTypeSelector.TitleStyle = ui.NewStyle(ui.ColorWhite, ui.ColorBlack, ui.ModifierBold)
	pressureTypeSelector.SetRect(37, 1, 65, 4)
	pressureTypeSelector.Border = true
	pressureTypeSelector.ActiveTabIndex = pressureType.tabIndex()
	pressureTypeSelector.ActiveTabStyle.Modifier = ui.ModifierBold

	graphMetricSelector := widgets.NewTabPane("avg[1]0", "avg[6]0", "avg[3]00")
	graphMetricSelector.PaddingLeft = 1
	graphMetricSelector.Title = "Graph metric"
	graphMetricSelector.TitleStyle = ui.NewStyle(ui.ColorWhite, ui.ColorBlack, ui.ModifierBold)
	graphMetricSelector.SetRect(67, 1, 98, 4)
	graphMetricSelector.Border = true
	graphMetricSelector.ActiveTabIndex = graphMetric.tabIndex()
	graphMetricSelector.ActiveTabStyle.Modifier = ui.ModifierBold

	return []ui.Drawable{resourceSelector, pressureTypeSelector, graphMetricSelector}
}

func renderSingleResourcePane(data []*pressure.AllPressures, resource resource, pressureType pressureType, graphMetric graphMetric) []ui.Drawable {
	tableWidth, tableHeight := 90, 5
	if pressureType == both {
		tableHeight = 7
	}
	x, y := 4, 5

	var title string
	var titleStyle ui.Style

	pressures := make([]pressure.ResourcePressure, len(data))

	switch resource {
	case resourceCPU:
		title = "CPU"
		titleStyle = ui.NewStyle(ui.ColorCyan, ui.ColorBlack, ui.ModifierBold)

		for i, p := range data {
			pressures[i] = p.CPU
		}
	case resourceMemory:
		title = "Memory"
		titleStyle = ui.NewStyle(ui.ColorBlue, ui.ColorBlack, ui.ModifierBold)

		for i, p := range data {
			pressures[i] = p.Memory
		}
	case resourceIO:
		title = "IO"
		titleStyle = ui.NewStyle(ui.ColorMagenta, ui.ColorBlack, ui.ModifierBold)

		for i, p := range data {
			pressures[i] = p.IO
		}
	}

	table := newPressureTable(title, pressures, pressureType)
	table.TitleStyle = titleStyle
	table.SetRect(x, y, x+tableWidth, y+tableHeight)

	y += tableHeight + 1

	graphHeight := 32
	graph := newPressureGraph(pressures, pressureType, graphMetric)
	graph.SetRect(x, y, x+tableWidth, y+graphHeight)

	return []ui.Drawable{table, graph}
}

func renderAllPane(data []*pressure.AllPressures, pressureType pressureType) []ui.Drawable {
	tableHeight := 5
	if pressureType == both {
		tableHeight = 7
	}

	tableWidth := 90
	x, y := 4, 5

	cpuPressures := make([]pressure.ResourcePressure, len(data))
	for i, p := range data {
		cpuPressures[i] = p.CPU
	}

	memPressures := make([]pressure.ResourcePressure, len(data))
	for i, p := range data {
		memPressures[i] = p.Memory
	}

	ioPressures := make([]pressure.ResourcePressure, len(data))
	for i, p := range data {
		ioPressures[i] = p.IO
	}

	cpuTable := newPressureTable("CPU", cpuPressures, pressureType)
	cpuTable.SetRect(x, y, x+tableWidth, y+tableHeight)
	cpuTable.TitleStyle = ui.NewStyle(ui.ColorCyan, ui.ColorBlack, ui.ModifierBold)
	y += tableHeight + 1

	memTable := newPressureTable("Memory", memPressures, pressureType)
	memTable.SetRect(x, y, x+tableWidth, y+tableHeight)
	memTable.TitleStyle = ui.NewStyle(ui.ColorBlue, ui.ColorBlack, ui.ModifierBold)
	y += tableHeight + 1

	ioTable := newPressureTable("IO", ioPressures, pressureType)
	ioTable.SetRect(x, y, x+tableWidth, y+tableHeight)
	ioTable.TitleStyle = ui.NewStyle(ui.ColorMagenta, ui.ColorBlack, ui.ModifierBold)

	return []ui.Drawable{cpuTable, memTable, ioTable}
}
