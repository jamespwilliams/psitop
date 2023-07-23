package main

type resource string
type pressureType string
type graphMetric int

const (
	resourceAll    resource = "all"
	resourceCPU             = "resourceCPU"
	resourceMemory          = "resourceMemory"
	resourceIO              = "resourceIO"

	avg10  graphMetric = 10
	avg60  graphMetric = 60
	avg300 graphMetric = 300

	some pressureType = "some"
	full              = "full"
	both              = "both"
)

func (k resource) tabIndex() int {
	switch k {
	case resourceAll:
		return 0
	case resourceCPU:
		return 1
	case resourceMemory:
		return 2
	case resourceIO:
		return 3
	default:
		return -1
	}
}

func (g graphMetric) tabIndex() int {
	switch g {
	case avg10:
		return 0
	case avg60:
		return 1
	case avg300:
		return 2
	default:
		return -1
	}
}

func (sf pressureType) tabIndex() int {
	switch sf {
	case some:
		return 0
	case full:
		return 1
	case both:
		return 2
	default:
		return -1
	}
}
