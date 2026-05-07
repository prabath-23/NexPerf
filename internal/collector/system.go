package collector

import (
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
)

func CollectSystem() (SystemSummary, error) {
	cpuPercent, err := cpu.Percent(150*time.Millisecond, false)
	if err != nil {
		return SystemSummary{}, err
	}

	vm, err := mem.VirtualMemory()
	if err != nil {
		return SystemSummary{}, err
	}

	usage, err := disk.Usage("/")
	if err != nil {
		return SystemSummary{}, err
	}

	hostname := ""
	if info, err := host.Info(); err == nil {
		hostname = info.Hostname
	}

	percent := 0.0
	if len(cpuPercent) > 0 {
		percent = cpuPercent[0]
	}

	return SystemSummary{
		CPUPercent: percent,
		Memory: MemoryInfo{
			Total:   vm.Total,
			Used:    vm.Used,
			Percent: vm.UsedPercent,
		},
		Disk: DiskInfo{
			Total:   usage.Total,
			Used:    usage.Used,
			Percent: usage.UsedPercent,
		},
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
		Hostname:  hostname,
		Timestamp: time.Now(),
	}, nil
}
