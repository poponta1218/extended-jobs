// vim: set ft=go:
package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

type Process struct {
	pid   int
	ppid  int
	uid   int
	flag  int
	state string
	cmd   string
	cwd   string
}

func GetValidShells() []string {
	file, err := os.Open("/etc/shells")
	if err != nil {
		return []string{}
	}
	defer file.Close()

	var shells []string
	s := bufio.NewScanner(file)
	for s.Scan() {
		if strings.HasPrefix(s.Text(), "#") {
			continue
		}
		shells = append(shells, s.Text())
	}

	return shells
}

func GetProcInfo(pid int) Process {
	var p Process
	procDir := fmt.Sprintf("/proc/%d", pid)

	// Parse /proc/[pid]/status
	// only parse Uid field, which is in the 9th line
	if statusFile, err := os.Open(procDir + "/status"); err == nil {
		s := bufio.NewScanner(statusFile)
		for i := 0; i < 9; i++ {
			s.Scan()
		}
		parts := strings.Fields(s.Text())
		if len(parts) > 1 {
			p.uid, _ = strconv.Atoi(parts[1])
		}
		statusFile.Close()
	}

	// Parse /proc/[pid]/stat
	if statFile, err := os.Open(procDir + "/stat"); err == nil {
		s := bufio.NewScanner(statFile)
		if s.Scan() {
			parts := strings.Fields(s.Text())
			if len(parts) > 8 {
				p.pid, _ = strconv.Atoi(parts[0])
				p.state = parseState(parts[2])
				p.ppid, _ = strconv.Atoi(parts[3])
				p.flag = parseFlag(parts[8])
			}
		}
		statFile.Close()
	}

	// Parse /proc/[pid]/cmdline
	if cmdFile, err := os.Open(procDir + "/cmdline"); err == nil {
		s := bufio.NewScanner(cmdFile)
		if s.Scan() {
			// Remove null bytes and trailing whitespaces
			p.cmd = strings.TrimSuffix(strings.ReplaceAll(strings.TrimSpace(s.Text()), "\x00", " "), " ")
		}
		cmdFile.Close()
	}

	// Resolve /proc/[pid]/cwd
	if cwdPath, err := os.Readlink(procDir + "/cwd"); err == nil {
		p.cwd = cwdPath
	}

	return p
}

func parseState(state string) string {
	switch state[0] {
	case 'R', 'D', 'S': // Running, Uninterruptible sleep, Interruptible sleep
		return "RUNNING"
	case 'T': // Stopped
		return "STOPPED"
	case 'Z': // Zombie
		return "DEFUNCT"
	default:
		return "UNKNOWN"
	}
}

func parseFlag(flagStr string) int {
	flag, _ := strconv.Atoi(flagStr)
	hexStr := fmt.Sprintf("%x", flag)
	if len(hexStr) < 3 {
		hexStr = strings.Repeat("0", 3-len(hexStr)) + hexStr
	}
	suffix, _ := strconv.ParseInt("0x"+hexStr[len(hexStr)-3:], 0, 64)
	return map[int]int{
		0x000: 0, // Normal flag
		0x040: 1, // Forked but didn't exec
		0x100: 4, // Traced
		0x140: 5, // Forked but didn't exec and traced
	}[int(suffix)]
}

func SortProcsByPid(procs []Process) []Process {
	sort.Slice(procs, func(i, j int) bool {
		return procs[i].pid < procs[j].pid
	})
	return procs
}

func FilterProcs(procs []Process, filterFunc func(Process) bool) []Process {
	var filtered []Process
	for _, p := range procs {
		if filterFunc(p) {
			filtered = append(filtered, p)
		}
	}
	return filtered
}

func GetMaxWidth(procs []Process, init int, fieldFunc func(Process) string) int {
	max := init
	for _, p := range procs {
		length := len(fieldFunc(p))
		if length > max {
			max = length
		}
	}
	return max
}

func DisplayProcsInfo(procs []Process) {
	widthPID := GetMaxWidth(procs, 3, func(p Process) string { return strconv.Itoa(p.pid) })
	widthState := GetMaxWidth(procs, 4, func(p Process) string { return p.state })
	widthCmd := GetMaxWidth(procs, 7, func(p Process) string { return p.cmd })
	widthCwd := GetMaxWidth(procs, 8, func(p Process) string { return p.cwd })

	fmt.Printf("%*s %*s %*s %-*s\n", widthPID, "PID", widthState, "STAT", widthCmd, "COMMAND", widthCwd, "DIRECTORY")
	for _, p := range procs {
		fmt.Printf("%*d %*s %*s %-*s\n", widthPID, p.pid, widthState, p.state, widthCmd, p.cmd, widthCwd, p.cwd)
	}
}

func main() {
	selfPID := os.Getpid() // Get current process's PID
	selfUID := os.Getuid() // Get current user's UID

	entries, err := os.ReadDir("/proc")
	if err != nil {
		panic(err)
	}

	var procs []Process
	for _, entry := range entries {
		if pid, err := strconv.Atoi(entry.Name()); err == nil && pid != selfPID {
			procs = append(procs, GetProcInfo(pid))
		}
	}
	procs = SortProcsByPid(procs)

	procs = FilterProcs(procs, func(p Process) bool { return p.uid == selfUID })
	shells := GetValidShells()
	filteredShells := FilterProcs(procs, func(p Process) bool {
		for _, shell := range shells {
			exe := strings.Split(p.cmd, " ")[0]
			if strings.HasPrefix(p.cmd, "-") || strings.HasPrefix(exe, shell) {
				return true
			}
		}
		return false
	})

	parentPIDs := map[int]bool{1: true}
	for _, p := range filteredShells {
		parentPIDs[p.pid] = true
	}

	procs = FilterProcs(procs, func(p Process) bool { return parentPIDs[p.ppid] })
	procs = FilterProcs(procs, func(p Process) bool { return p.flag == 0 })

	DisplayProcsInfo(procs)
}
