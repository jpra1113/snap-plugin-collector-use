package use

import (
	"bufio"
	"bytes"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/host"
)

func readLines(filename string) ([]string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return []string{""}, errors.Errorf("Unable to open file %s: %s", filename, err.Error())
	}
	defer f.Close()

	var ret []string
	var offset uint

	r := bufio.NewReader(f)
	n := -1
	for i := 0; i < n+int(offset) || n < 0; i++ {
		line, err := r.ReadString('\n')
		ret = append(ret, strings.Trim(line, "\n"))
		if err != nil {
			break
		}
		if i < int(offset) {
			continue
		}
	}

	return ret, nil

}

func run(cmd string, args []string) ([]byte, error) {
	command := exec.Command(cmd, args...)
	var b bytes.Buffer
	command.Stdout = &b
	if err := command.Run(); err != nil {
		return nil, errors.Errorf("Failed to run cmd %s: %s", cmd, err.Error())
	}
	return b.Bytes(), nil
}

func readInt(filename string) (int64, error) {
	f, err := os.Open(filename)
	if err != nil {
		return 0, errors.Errorf("Unable to open file %s: %s", filename, err.Error())
	}
	defer f.Close()

	r := bufio.NewReader(f)

	// The int files that this is concerned with should only be one liners.
	line, err := r.ReadString('\n')
	if err != nil {
		return 0, errors.Errorf("Unable to read string: %s", err.Error())
	}

	trimmedLine := strings.TrimSpace(line)
	i, err := strconv.ParseInt(trimmedLine, 10, 32)
	if err != nil {
		return 0, errors.Errorf("Unable to parse int from line %s: %s", trimmedLine, err.Error())
	}

	return i, nil
}

func hostTags() (map[string]string, error) {
	tags := make(map[string]string)

	hostInfo, err := host.Info()
	if err != nil {
		return tags, errors.Errorf("Unable to get host info: %s", err.Error())
	}

	tags["hostname"] = hostInfo.Hostname
	tags["os"] = hostInfo.OS
	tags["platform"] = hostInfo.Platform
	tags["platform_family"] = hostInfo.PlatformFamily
	tags["platform_version"] = hostInfo.PlatformVersion
	tags["virtualization_role"] = hostInfo.VirtualizationRole
	tags["virtualization_system"] = hostInfo.VirtualizationSystem

	return tags, nil

}
