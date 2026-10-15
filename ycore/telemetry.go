package ycore

import (
	"fmt"
	"runtime"
	"sync"
)

var (
	teleMu    sync.Mutex
	telemetry *Telemetry = NewTelemetry()
)

type Telemetry struct {
	TotalEstimatedGPUMemory uint64
}

func NewTelemetry() *Telemetry {
	return &Telemetry{}
}

func UpdateGPUMem(bytes uint64) {
	teleMu.Lock()
	defer teleMu.Unlock()

	telemetry.TotalEstimatedGPUMemory += bytes
}

func PrintMemoryUsage() {
	bToMb := func(b uint64) uint64 {
		return b / 1024 / 1024
	}
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	fmt.Printf("Alloc = %v MiB", bToMb(m.Alloc))
	fmt.Printf("Alloc = %v MiB", bToMb(m.HeapAlloc))
	fmt.Printf("\tTotalAlloc = %v MiB", bToMb(m.TotalAlloc))
	fmt.Printf("\tSys = %v MiB", bToMb(m.Sys))
	fmt.Printf("\tGPU = %v MiB", bToMb(telemetry.TotalEstimatedGPUMemory))
	fmt.Printf("\tNumGC = %v\n", m.NumGC)
}
