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
	"errors"
	"path/filepath"
	"regexp"
	"time"

	"github.com/aasssddd/snap-plugin-lib-go/v1/plugin"
)

const (
	// Name of plugin
	name = "Use"
	// Version of plugin
	version = 1

	waitTime = 10 * time.Millisecond
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
	Host         string
	initialized  bool
	ProcPath     string
	DiskStatPath string
	CpuStatPath  string
	LoadAvgPath  string
	MemInfoPath  string
	VmStatPath   string
}

// NewUseCollector returns Use struct
func NewUseCollector() *Use {
	return &Use{}
}

func (u *Use) init(cfg plugin.Config) {
	procPath, err := cfg.GetString("proc_path")
	if err != nil {
		procPath = "/proc"
	}

	u.ProcPath = procPath
	u.DiskStatPath = filepath.Join(procPath, "diskstats")
	u.CpuStatPath = filepath.Join(procPath, "stat")
	u.LoadAvgPath = filepath.Join(procPath, "loadavg")
	u.MemInfoPath = filepath.Join(procPath, "meminfo")
	u.VmStatPath = filepath.Join(procPath, "vmstat")
	u.initialized = true
}

// CollectMetrics returns Use metrics
func (u *Use) CollectMetrics(mts []plugin.Metric) ([]plugin.Metric, error) {
	cfg := mts[0].Config
	if !u.initialized {
		u.init(cfg)
	}

	metrics := make([]plugin.Metric, len(mts))
	for i, p := range mts {
		ns := p.Namespace.String()
		switch {
		case cpure.MatchString(ns):
			metric, err := u.computeStat(p.Namespace)
			if err != nil {
				return nil, errors.New("Unable to get compute stat: " + err.Error())
			}
			metrics[i] = *metric

		case storre.MatchString(ns):
			metric, err := u.diskStat(p.Namespace)
			if err != nil {
				return nil, errors.New("Unable to get disk stat: " + err.Error())
			}
			metrics[i] = *metric
		case memre.MatchString(ns):
			metric, err := memStat(p.Namespace, u.VmStatPath, u.MemInfoPath)
			if err != nil {
				return nil, errors.New("Unable to get mem stat: " + err.Error())
			}
			metrics[i] = *metric
		}
		tags, err := hostTags()

		if err == nil {
			metrics[i].Tags = tags
		}
		metrics[i].Timestamp = time.Now()

	}
	return metrics, nil
}

// GetMetricTypes returns the metric types exposed by use plugin
func (u *Use) GetMetricTypes(cfg plugin.Config) ([]plugin.Metric, error) {
	if !u.initialized {
		u.init(cfg)
	}

	mts := []plugin.Metric{}

	cpu, err := getCPUMetricTypes()
	if err != nil {
		return nil, errors.New("Unable to get cpu metric types: " + err.Error())
	}
	mts = append(mts, cpu...)
	disk, err := getDiskMetricTypes()
	if err != nil {
		return nil, errors.New("Unable to get disk metric types: " + err.Error())
	}
	mts = append(mts, disk...)
	mem, err := getMemMetricTypes()
	if err != nil {
		return nil, errors.New("Unable to get mem metric types: " + err.Error())
	}
	mts = append(mts, mem...)

	return mts, nil
}

//GetConfigPolicy returns a ConfigPolicy
func (u *Use) GetConfigPolicy() (plugin.ConfigPolicy, error) {
	policy := plugin.NewConfigPolicy()
	policy.AddNewStringRule([]string{"intel", "use"}, "proc_path", false, plugin.SetDefaultString("/proc"))
	return *policy, nil
}
