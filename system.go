package main

import (
	"fmt"

	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
)

// CollectSysStats gathers hostname, human-readable uptime, and RAM usage.
// Returned as a sysStatsMsg ready to be dispatched as a tea.Msg.
func CollectSysStats() sysStatsMsg {
	msg := sysStatsMsg{}

	// Hostname + uptime
	info, err := host.Info()
	if err != nil {
		msg.Err = err
		return msg
	}
	msg.Hostname = info.Hostname
	msg.Uptime = formatUptime(info.Uptime)

	// RAM
	vmStat, err := mem.VirtualMemory()
	if err != nil {
		msg.Err = err
		return msg
	}
	const gb = 1 << 30
	msg.MemUsed = float64(vmStat.Used) / gb
	msg.MemTotal = float64(vmStat.Total) / gb

	return msg
}

// formatUptime converts seconds to a "Xd Yh Zm" string.
func formatUptime(seconds uint64) string {
	days := seconds / 86400
	seconds %= 86400
	hours := seconds / 3600
	seconds %= 3600
	minutes := seconds / 60

	switch {
	case days > 0:
		return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
	case hours > 0:
		return fmt.Sprintf("%dh %dm", hours, minutes)
	default:
		return fmt.Sprintf("%dm", minutes)
	}
}
