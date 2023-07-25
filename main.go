package main

import (
	"fmt"
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
	if err := run(); err != nil {
		log.Fatalln(err)
	}
}

func run() error {
	var (
		resource     resource     = resourceAll
		graphMetric  graphMetric  = avg10
		pressureType pressureType = both

		dataMux sync.Mutex
		data    []*pressure.AllPressures
	)

	fetchErr := make(chan error, 1)

	if err := ui.Init(); err != nil {
		return fmt.Errorf("failed to initialize termui: %w", err)
	}
	defer ui.Close()

	err := fetchPressures(&dataMux, &data)
	if err != nil {
		return fmt.Errorf("failed to fetch pressures: %w", err)
	}

	go func() {
		for {
			time.Sleep(fetchPeriod)
			if err := fetchPressures(&dataMux, &data); err != nil {
				fetchErr <- fmt.Errorf("failed to fetch pressures: %w", err)
			}
		}
	}()

	uiEvents := ui.PollEvents()
	for {
		dataMux.Lock()
		ui.Clear()
		render(resource, graphMetric, pressureType, data)
		dataMux.Unlock()

		select {
		case err := <-fetchErr:
			return err
		case <-time.After(renderPeriod):
		case e := <-uiEvents:
			switch e.ID {
			case "q", "<C-c>":
				return nil
			case "a":
				resource = resourceAll
				pressureType = both
			case "c":
				resource = resourceCPU
				pressureType = some
			case "m":
				resource = resourceMemory
				pressureType = both
			case "i":
				resource = resourceIO
				pressureType = both
			case "1":
				graphMetric = avg10
			case "6":
				graphMetric = avg60
			case "3":
				graphMetric = avg300
			case "s":
				pressureType = some
			case "f":
				if resource != resourceCPU {
					pressureType = full
				}
			case "b":
				if resource != resourceCPU {
					pressureType = both
				}
			}
		}
	}
}

func fetchPressures(mux *sync.Mutex, data *[]*pressure.AllPressures) error {
	mux.Lock()
	defer mux.Unlock()

	pressures, err := pressure.CurrentAllPressures()
	if err != nil {
		return err
	}
	*data = append(*data, pressures)

	if len(*data) > maxDataLength {
		*data = (*data)[1:]
	}
	return nil
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
	if resource == resourceCPU {
		pressureTypeSelector.TabNames = []string{"[s]ome"}
	}
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
	if pressureType == both && resource != resourceCPU {
		tableHeight = 7
	}
	x, y := 4, 5

	var title string
	var titleStyle ui.Style

	somePressures := make([]pressure.Pressure, len(data))
	fullPressures := make([]pressure.Pressure, len(data))

	switch resource {
	case resourceCPU:
		title = "CPU"
		titleStyle = ui.NewStyle(ui.ColorCyan, ui.ColorBlack, ui.ModifierBold)

		for i, p := range data {
			somePressures[i] = p.CPU.SomePressure
		}
		fullPressures = nil
	case resourceMemory:
		title = "Memory"
		titleStyle = ui.NewStyle(ui.ColorBlue, ui.ColorBlack, ui.ModifierBold)

		for i, p := range data {
			somePressures[i] = p.Memory.SomePressure
			fullPressures[i] = p.Memory.FullPressure
		}
	case resourceIO:
		title = "IO"
		titleStyle = ui.NewStyle(ui.ColorMagenta, ui.ColorBlack, ui.ModifierBold)

		for i, p := range data {
			somePressures[i] = p.IO.SomePressure
			fullPressures[i] = p.IO.FullPressure
		}
	}

	var table *widgets.Table
	switch pressureType {
	case some:
		table = newPressureTable(title, somePressures, nil)
	case full:
		table = newPressureTable(title, nil, fullPressures)
	case both:
		table = newPressureTable(title, somePressures, fullPressures)
	}

	table.TitleStyle = titleStyle
	table.SetRect(x, y, x+tableWidth, y+tableHeight)

	y += tableHeight + 1

	graphHeight := 32
	graph := newPressureGraph(somePressures, fullPressures, graphMetric)
	graph.SetRect(x, y, x+tableWidth, y+graphHeight)

	return []ui.Drawable{table, graph}
}

func renderAllPane(data []*pressure.AllPressures, pressureType pressureType) []ui.Drawable {
	cpuTableHeight, nonCPUTableHeight := 5, 5
	if pressureType == both {
		nonCPUTableHeight = 7
	}

	tableWidth := 90
	x, y := 4, 5

	cpuSomePressure := make([]pressure.Pressure, len(data))
	for i, p := range data {
		cpuSomePressure[i] = p.CPU.SomePressure
	}

	memSomePressure := make([]pressure.Pressure, len(data))
	memFullPressure := make([]pressure.Pressure, len(data))
	for i, p := range data {
		memSomePressure[i] = p.Memory.SomePressure
		memFullPressure[i] = p.Memory.FullPressure
	}

	ioSomePressure := make([]pressure.Pressure, len(data))
	ioFullPressure := make([]pressure.Pressure, len(data))
	for i, p := range data {
		ioSomePressure[i] = p.IO.SomePressure
		ioFullPressure[i] = p.IO.FullPressure
	}

	cpuTable := newPressureTable("CPU", cpuSomePressure, nil)
	cpuTable.SetRect(x, y, x+tableWidth, y+cpuTableHeight)
	cpuTable.TitleStyle = ui.NewStyle(ui.ColorCyan, ui.ColorBlack, ui.ModifierBold)
	y += cpuTableHeight + 1

	memTable := newPressureTable("Memory", memSomePressure, memFullPressure)
	memTable.SetRect(x, y, x+tableWidth, y+nonCPUTableHeight)
	memTable.TitleStyle = ui.NewStyle(ui.ColorBlue, ui.ColorBlack, ui.ModifierBold)
	y += nonCPUTableHeight + 1

	ioTable := newPressureTable("IO", ioSomePressure, ioFullPressure)
	ioTable.SetRect(x, y, x+tableWidth, y+nonCPUTableHeight)
	ioTable.TitleStyle = ui.NewStyle(ui.ColorMagenta, ui.ColorBlack, ui.ModifierBold)

	return []ui.Drawable{cpuTable, memTable, ioTable}
}
