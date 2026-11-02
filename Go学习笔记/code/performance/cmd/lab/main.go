package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"runtime/trace"
	"time"

	"example.com/perflab/internal/workload"
)

func writeProfile(path, name string) (err error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, f.Close()) }()
	profile := pprof.Lookup(name)
	if profile == nil {
		return fmt.Errorf("profile %q unavailable", name)
	}
	return profile.WriteTo(f, 0)
}

func run() (err error) {
	mode := flag.String("mode", "cpu", "cpu, alloc, or lock")
	duration := flag.Duration("duration", 500*time.Millisecond, "CPU workload duration")
	n := flag.Int("n", 100, "allocation count or rounds per lock worker")
	size := flag.Int("size", 4096, "allocation size")
	keep := flag.Int("keep", 16, "retain one in every N allocations")
	workers := flag.Int("workers", 4, "lock workload workers")
	pause := flag.Duration("pause", 100*time.Microsecond, "simulated work inside lock")
	cpuFile := flag.String("cpu", "", "CPU profile output")
	heapFile := flag.String("heap", "", "heap profile output")
	traceFile := flag.String("trace", "", "execution trace output")
	mutexFile := flag.String("mutex", "", "mutex profile output")
	blockFile := flag.String("block", "", "block profile output")
	flag.Parse()
	if *mode != "cpu" && *mode != "alloc" && *mode != "lock" {
		return errors.New("mode must be cpu, alloc, or lock")
	}
	if *duration <= 0 || *duration > time.Minute || *n < 1 || *n > 1000000 ||
		*size < 1 || *size > 1<<20 || *keep < 1 || *workers < 1 || *workers > 64 ||
		*pause < 0 || *pause > time.Second {
		return errors.New("workload parameters outside documented teaching bounds")
	}
	if *mode == "alloc" && uint64(*n)*uint64(*size) > 256<<20 {
		return errors.New("teaching allocation workload exceeds 256 MiB total allocations")
	}
	activeProfiles := 0
	for _, path := range []string{*cpuFile, *heapFile, *traceFile, *mutexFile, *blockFile} {
		if path != "" {
			activeProfiles++
		}
	}
	if activeProfiles > 1 {
		return errors.New("choose one recording mode per run to reduce measurement interference")
	}
	var recording *os.File
	var stop func()
	defer func() {
		if stop != nil {
			stop()
		}
		if recording != nil {
			err = errors.Join(err, recording.Close())
		}
	}()
	if *cpuFile != "" || *traceFile != "" {
		path := *cpuFile
		if *traceFile != "" {
			path = *traceFile
		}
		recording, err = os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
		if err != nil {
			return err
		}
		if *cpuFile != "" {
			if err := pprof.StartCPUProfile(recording); err != nil {
				return err
			}
			stop = pprof.StopCPUProfile
		} else {
			if err := trace.Start(recording); err != nil {
				return err
			}
			stop = trace.Stop
		}
	}
	if *mutexFile != "" {
		old := runtime.SetMutexProfileFraction(1)
		defer runtime.SetMutexProfileFraction(old)
	}
	if *blockFile != "" {
		runtime.SetBlockProfileRate(1)
		defer runtime.SetBlockProfileRate(0)
	}
	ctx, task := trace.NewTask(context.Background(), "lab-"+*mode)
	var retained [][]byte
	var checksum uint64
	var completed int
	trace.WithRegion(ctx, "workload", func() {
		switch *mode {
		case "cpu":
			deadline := time.Now().Add(*duration)
			for time.Now().Before(deadline) {
				checksum += workload.Checksum(20000)
				completed++
			}
		case "alloc":
			retained = workload.Allocate(*n, *size, *keep)
			completed = *n
		case "lock":
			completed = workload.Contended(*workers, *n, *pause)
		}
	})
	task.End()
	if *heapFile != "" {
		runtime.GC()
		if err := writeProfile(*heapFile, "heap"); err != nil {
			return err
		}
	}
	if *mutexFile != "" {
		if err := writeProfile(*mutexFile, "mutex"); err != nil {
			return err
		}
	}
	if *blockFile != "" {
		if err := writeProfile(*blockFile, "block"); err != nil {
			return err
		}
	}
	fmt.Printf("mode=%s completed=%d retained=%d checksum=%d\n", *mode, completed, len(retained), checksum)
	runtime.KeepAlive(retained)
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "lab:", err)
		os.Exit(1)
	}
}
