package main

import (
	"math"
	"os"
	"sync"

	"github.com/shirou/gopsutil/v3/process"
)

// procCache retains *process.Process objects between ticks so that
// p.Percent(0) can measure a real CPU delta (it needs two calls on the same
// object — the first seeds the baseline, the second returns the delta).
var (
	procCache   = map[int32]*process.Process{}
	procCacheMu sync.Mutex
)

// CollectProcesses returns a snapshot of all running processes.
// Real CPU% values appear from the second tick onward (~1 s after start).
func CollectProcesses() processesMsg {
	selfPID := int32(os.Getpid())

	pids, err := process.Pids()
	if err != nil {
		return processesMsg{Err: err}
	}

	procCacheMu.Lock()
	defer procCacheMu.Unlock()

	// Build a set of currently-live PIDs so we can evict stale cache entries.
	livePIDs := make(map[int32]struct{}, len(pids))
	for _, pid := range pids {
		livePIDs[pid] = struct{}{}
	}
	for pid := range procCache {
		if _, alive := livePIDs[pid]; !alive {
			delete(procCache, pid)
		}
	}

	rows := make([]ProcessRow, 0, len(pids))
	for _, pid := range pids {
		if pid == selfPID {
			continue
		}

		// Reuse cached object so CPU baseline persists between ticks.
		p, ok := procCache[pid]
		if !ok {
			p, err = process.NewProcess(pid)
			if err != nil {
				continue
			}
			procCache[pid] = p
		}

		name, err := p.Name()
		if err != nil || name == "" {
			continue // kernel/zombie process we can't read
		}

		// Percent(0) is non-blocking: call 1 seeds baseline (returns 0),
		// call 2+ returns real delta since last call.
		cpuPct, err := p.Percent(0)
		if err != nil || math.IsNaN(cpuPct) || math.IsInf(cpuPct, 0) || cpuPct < 0 {
			cpuPct = 0
		}

		memInfo, err := p.MemoryInfo()
		var memMB float64
		if err == nil && memInfo != nil {
			memMB = float64(memInfo.RSS) / (1 << 20)
		}

		threads, err := p.NumThreads()
		if err != nil {
			threads = 0
		}

		// Username may fail without elevated privileges — show N/A, don't crash.
		username, err := p.Username()
		if err != nil || username == "" {
			username = "N/A"
		}
		// Strip domain prefix on Windows (DOMAIN\user → user)
		for i := len(username) - 1; i >= 0; i-- {
			if username[i] == '\\' {
				username = username[i+1:]
				break
			}
		}

		rows = append(rows, ProcessRow{
			PID:     pid,
			Name:    name,
			CPU:     cpuPct,
			MemMB:   memMB,
			Threads: threads,
			User:    username,
		})
	}

	return processesMsg{Procs: rows}
}
