package use

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"
	"github.com/pkg/errors"
)

// MemInfo struct for storing IO Data
type MemInfo struct {
	MemTotal    float64
	MemFree     float64
	SwapIn      float64
	SwapOut     float64
	vmStatPath  string
	memInfoPath string
}

// Utilization returns utilization of Memory
func (m *MemInfo) Utilization() (float64, error) {
	memInfo, err := readStatForMemInfo(m.memInfoPath)

	if err != nil {
		return 0.0, err
	}

	m.MemFree = float64(memInfo["MemFree"])
	m.MemTotal = float64(memInfo["MemTotal"])

	if m.MemTotal >= 0 {
		return 100.0 - (m.MemFree / m.MemTotal * 100), nil
	}
	return 0.0, errors.Errorf("Error Total Memory is lower or equal 0")
}

// Saturation returns saturation of Memory
func (m *MemInfo) Saturation() (float64, error) {
	memInfo, err := readStatForVMStat(m.vmStatPath)
	if err != nil {
		return 0.0, err
	}

	m.SwapIn = float64(memInfo["SwapIn"])
	m.SwapOut = float64(memInfo["SwapOut"])
	if m.SwapOut > 0 {
		return m.SwapIn / m.SwapOut * 100.00, nil
	}
	return 0.0, nil
}

func getMemMetricTypes() ([]plugin.Metric, error) {
	var mts []plugin.Metric
	for _, name := range metricLabels {

		mts = append(mts, plugin.Metric{Namespace: plugin.NewNamespace("intel", "use", "memory", name)})
	}
	return mts, nil
}

func memStat(ns plugin.Namespace, vmStatPath string, memInfoPath string) (*plugin.Metric, error) {
	switch {
	case regexp.MustCompile(`^/intel/use/memory/utilization$`).MatchString(ns.String()):
		m := MemInfo{vmStatPath: vmStatPath, memInfoPath: memInfoPath}
		metric, err := m.Utilization()
		if err != nil {
			return nil, errors.Errorf("Unable to get memory utilization: %s", err.Error())
		}
		return &plugin.Metric{
			Namespace: ns,
			Data:      metric,
		}, nil

	case regexp.MustCompile(`^/intel/use/memory/saturation$`).MatchString(ns.String()):
		m := MemInfo{vmStatPath: vmStatPath, memInfoPath: memInfoPath}
		metric, err := m.Saturation()
		if err != nil {
			return nil, errors.Errorf("Unable to get memory saturation: %s", err.Error())
		}

		return &plugin.Metric{
			Namespace: ns,
			Data:      metric,
		}, nil
	}
	return nil, fmt.Errorf("Unknown memory namespace processing %v", ns)
}

func readStatForMemInfo(memInfoPath string) (map[string]int64, error) {
	lines, err := readLines(memInfoPath)
	ret := make(map[string]int64, 2)
	if err != nil {
		return ret, err
	}
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) > 0 {
			key := strings.TrimSpace(fields[0])
			value := strings.TrimSpace(fields[1])
			switch key {
			case "MemTotal:":
				ret["MemTotal"], err = strconv.ParseInt(value, 10, 64)
				if err != nil {
					return ret, errors.Errorf("Unable to parse int from mem total %s: %s", value, err.Error())
				}
			case "MemFree:":
				ret["MemFree"], err = strconv.ParseInt(value, 10, 64)
				if err != nil {
					return ret, errors.Errorf("Unable to parse int from mem free %s: %s", value, err.Error())
				}
			}
		}
	}
	return ret, nil
}

func readStatForVMStat(vmStatPath string) (map[string]int64, error) {
	filename := vmStatPath
	ret := make(map[string]int64, 2)
	lines, err := readLines(filename)
	if err != nil {
		return ret, err
	}

	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) > 0 {
			key := strings.TrimSpace(fields[0])
			value := strings.TrimSpace(fields[1])
			switch key {
			case "pswpin":
				ret["SwapIn"], err = strconv.ParseInt(value, 10, 64)
				if err != nil {
					return nil, err
				}
			case "pswpout":
				ret["SwapOut"], err = strconv.ParseInt(value, 10, 64)
				if err != nil {
					return nil, err
				}
			}

		}
	}
	return ret, nil
}
