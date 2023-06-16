package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// https://stackoverflow.com/questions/31879817/golang-os-exec-realtime-memory-usage
func calculateMemory(pid int) (uint64, error) {
	f, err := os.Open(fmt.Sprintf("/proc/%d/smaps", pid))
	if err != nil {
		return 0, err
	}
	defer f.Close()

	res := uint64(0)
	pfx := []byte("Pss:")
	r := bufio.NewScanner(f)
	for r.Scan() {
		line := r.Bytes()
		if bytes.HasPrefix(line, pfx) {
			var size uint64
			_, err := fmt.Sscanf(string(line[4:]), "%d", &size)
			if err != nil {
				return 0, err
			}
			res += size
		}
	}
	if err := r.Err(); err != nil {
		return 0, err
	}

	return res, nil
}

func init() {
	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s <command>\n", os.Args[0])
		os.Exit(1)
	}
}

func humanReadableKBCountSI(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d kB", b)
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(b)/float64(div), "MGTPE"[exp])
}

func main() {
	args := []string{"-c", strings.Join(os.Args[1:], " ")}

	c := exec.Command("sh", args...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Start()

	var memTotal uint64
	go func() {
		for {
			mem, _ := calculateMemory(c.Process.Pid)
			if memTotal < mem {
				memTotal = mem
			}
			time.Sleep(time.Second / 100)
		}
	}()

	c.Wait()

	fmt.Printf("\nMax memory usage: %s\n", humanReadableKBCountSI(memTotal))
}
