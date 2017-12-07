package use

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/aasssddd/snap-plugin-lib-go/v1/plugin"
	"github.com/pkg/errors"
)

// DiskStat struct for storing disk metric Data
type DiskStat struct {
	last         int64
	current      int64
	diskName     string
	diskStatPath string
}

// Utilization returns utilization of Disk Device
func (d *DiskStat) Utilization() (float64, error) {
	var err error
	d.last, err = readStatForDisk(d.diskName, "timeio", d.diskStatPath)
	if err != nil {
		return 0.0, err
	}
	time.Sleep(waitTime)
	d.current, err = readStatForDisk(d.diskName, "timeio", d.diskStatPath)
	if err != nil {
		return 0.0, err
	}
	return float64(d.current-d.last) / 10.0, nil
}

// Saturation returns saturation of Disk Device
func (d *DiskStat) Saturation() (float64, error) {
	var err error
	d.last, err = readStatForDisk(d.diskName, "weightedtimeio", d.diskStatPath)
	if err != nil {
		return 0.0, err
	}
	d.current, err = readStatForDisk(d.diskName, "weightedtimeio", d.diskStatPath)
	if err != nil {
		return 0.0, err
	}
	// 10ms * 10 ticks
	return float64(d.current-d.last) / 100.0, nil
}

func getDiskMetricTypes() ([]plugin.Metric, error) {
	var mts []plugin.Metric

	for _, diskName := range listDisks() {
		for _, name := range metricLabels {
			mts = append(mts, plugin.Metric{Namespace: plugin.NewNamespace("intel", "use", "storage", diskName, name)})
		}

	}
	return mts, nil
}

func listDisks() []string {
	cmd := "lsblk"
	args := []string{"-d", "--noheadings", "--list", "-o", "NAME"}
	output, err := run(cmd, args)
	if err != nil {
		return []string{}
	}
	disks := []string{}
	for _, disk := range strings.Split(string(output), "\n") {
		parts := strings.Split(disk, " ")
		if len(parts) == 1 {
			// We assume a multipart string is not a supported disk
			// e.g: "lsblk: dm-0: failed to get device path"
			disks = append(disks, disk)
		}
	}

	return disks
}

func readStatForDisk(diskName string, statType string, diskStatPath string) (int64, error) {
	if diskName == "" {
		log.Warningf("Empty diskName")
		return 0, nil
	}

	lines, err := readLines(diskStatPath)
	if err != nil {
		return 0, err
	}

	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		if diskName == fields[2] {
			switch statType {
			case "timeio":
				return strconv.ParseInt((fields[12]), 10, 64)
			case "weightedtimeio":
				return strconv.ParseInt((fields[13]), 10, 64)
			}
		}
	}

	return 0, fmt.Errorf("Can't find a disk [%s].\n", diskName)
}

func (u *Use) diskStat(ns plugin.Namespace) (*plugin.Metric, error) {
	diskName := ns.Strings()[3]
	switch {
	case regexp.MustCompile(`^/intel/use/storage/.*/utilization$`).MatchString(ns.String()):
		diskStat := DiskStat{diskName: diskName, diskStatPath: u.DiskStatPath}
		metric, err := diskStat.Utilization()
		if err != nil {
			return nil, errors.Errorf("Unable to get disk utilization: %s", err.Error())
		}
		return &plugin.Metric{
			Namespace: ns,
			Data:      metric,
		}, nil
	case regexp.MustCompile(`^/intel/use/storage/.*/saturation$`).MatchString(ns.String()):
		diskStat := DiskStat{diskName: diskName, diskStatPath: u.DiskStatPath}
		metric, err := diskStat.Saturation()
		if err != nil {
			return nil, errors.Errorf("Unable to get disk saturation: " + err.Error())
		}

		return &plugin.Metric{
			Namespace: ns,
			Data:      float64(metric),
		}, nil
	}

	return nil, errors.Errorf("Unknown disk stat namespace %v", ns)
}
