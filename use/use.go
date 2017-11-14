/*
http://www.apache.org/licenses/LICENSE-2.0.txt

Copyright 2016 Intel Corporation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package use

import (
	"path/filepath"
	"regexp"
	"time"

	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/control/plugin/cpolicy"
	"github.com/intelsdi-x/snap/core/ctypes"
)

const (
	// Name of plugin
	name = "Use"
	// Version of plugin
	version = 1
	// Type of plugin
	pluginType = plugin.CollectorPluginType
	waitTime   = 10 * time.Millisecond
)

var (
	metricLabels = []string{
		"utilization",
		"saturation",
	}
	cpure  = regexp.MustCompile(`^/intel/use/compute/.*`)
	storre = regexp.MustCompile(`^/intel/use/storage/.*`)
	memre  = regexp.MustCompile(`^/intel/use/memory/.*`)
)

// Use contains values of previous measurments
type Use struct {
	host         string
	initialized  bool
	procPath     string
	diskStatPath string
	cpuStatPath  string
	loadAvgPath  string
	memInfoPath  string
	vmStatPath   string
}

// Meta returns name, version and plugin type
func Meta() *plugin.PluginMeta {
	return plugin.NewPluginMeta(name, version, pluginType, []string{plugin.SnapGOBContentType}, []string{plugin.SnapGOBContentType})
}

// NewUseCollector returns Use struct
func NewUseCollector() *Use {
	return &Use{}
}

func (u *Use) init(table map[string]ctypes.ConfigValue) {
	procPath := "/proc"
	procPathValue, ok := table["proc_path"]
	if ok {
		procPath = procPathValue.(ctypes.ConfigValueStr).Value
	}
	u.procPath = procPath

	u.diskStatPath = filepath.Join(procPath, "diskstats")
	u.cpuStatPath = filepath.Join(procPath, "stat")
	u.loadAvgPath = filepath.Join(procPath, "loadavg")
	u.memInfoPath = filepath.Join(procPath, "meminfo")
	u.vmStatPath = filepath.Join(procPath, "vmstat")
	u.initialized = true
}

// CollectMetrics returns Use metrics
func (u *Use) CollectMetrics(mts []plugin.MetricType) ([]plugin.MetricType, error) {
	metrics := make([]plugin.MetricType, len(mts))

	for i, p := range mts {
		ns := p.Namespace().String()
		switch {
		case cpure.MatchString(ns):
			metric, err := u.computeStat(p.Namespace())
			if err != nil {
				return nil, err
			}
			metrics[i] = *metric

		case storre.MatchString(ns):
			metric, err := u.diskStat(p.Namespace())
			if err != nil {
				return nil, err
			}
			metrics[i] = *metric
		case memre.MatchString(ns):
			metric, err := memStat(p.Namespace(), u.vmStatPath, u.memInfoPath)
			if err != nil {
				return nil, err
			}
			metrics[i] = *metric
		}
		tags, err := hostTags()

		if err == nil {
			metrics[i].Tags_ = tags
		}
		metrics[i].Timestamp_ = time.Now()

	}
	return metrics, nil
}

// GetMetricTypes returns the metric types exposed by use plugin
func (u *Use) GetMetricTypes(cfg plugin.ConfigType) ([]plugin.MetricType, error) {
	if !u.initialized {
		u.init(cfg.Table())
	}

	mts := []plugin.MetricType{}

	cpu, err := getCPUMetricTypes()
	if err != nil {
		return nil, err
	}
	mts = append(mts, cpu...)
	disk, err := getDiskMetricTypes()
	if err != nil {
		return nil, err
	}
	mts = append(mts, disk...)
	mem, err := getMemMetricTypes()
	if err != nil {
		return nil, err
	}
	mts = append(mts, mem...)

	return mts, nil
}

//GetConfigPolicy returns a ConfigPolicy
func (u *Use) GetConfigPolicy() (*cpolicy.ConfigPolicy, error) {
	cp := cpolicy.New()
	config := cpolicy.NewPolicyNode()
	rule, err := cpolicy.NewStringRule("proc_path", false, "/proc")
	if err != nil {
		return nil, err
	}
	config.Add(rule)
	cp.Add([]string{"intel", "use"}, config)

	return cp, nil
}
func handleErr(e error) {
	if e != nil {
		panic(e)

	}
}
