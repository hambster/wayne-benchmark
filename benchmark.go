package main

/**
 * CPU Stress Testing Tool
 * @Author: hambster@gmail.com
 **/

import (
	"crypto/sha512"
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"
)

const (
	EnvNumThreads           = "NUM_THREADS"
	EnvNumThreadsDefault    = 2
	ModeIsolation           = "isolation"
	ModeShared              = "shared"
	ModeSpinLock            = "spinlock"
	IterationDefault        = 10000
	NumRandomGen            = 128
	RandomWriteRatioDefault = float64(0.05)
	Bytes2WriteSizeDefault  = 16 * 1024 //16KB
)

func usage() {
	cliName := os.Args[0]
	fmt.Printf(`
	command:
		%s [%s|%s|%s] [number-of-iteration] [random-write-ratio] [path-to-write] [bytes-to-write]

	note:
		the maximum value of random-write-ratio is 100, which means you will definite write data in every iteration.

	example:

		# populate 16 threads to stress CPU with 1000000 iteration without interacting with each other
		# and generate 20%% random write 1KB data in /tmp with
		export NUM_THREADS=16
		%s %s 1000000 20 /tmp 1024

		# populate 16 threads to stress CPU with 1000000 iteration with interacting with each other
		# and generate 30%% random write 2KB in /mnt/resource
		export NUM_THREADS=16
		%s %s 1000000 30 /mnt/resource 2048

		# populate 16 threads to stress CPU with 1000000 iteration with interacting with each other
		# and generate 90%% random write 2KB in /tmp
		export NUM_THREADS=16
		taskset 0xFFF %s %s 1000000 90 /tmp 2048

		# populate 16 threads to stress CPU with 1000000 iteration with interacting with each other using spinlock
		# and generate 90%% random write 2KB in /tmp
		export NUM_THREADS=16
		taskset 0xFFF %s %s 1000000 90 /tmp 2048

`,
		cliName,
		ModeIsolation, ModeShared, ModeSpinLock,
		cliName, ModeIsolation,
		cliName, ModeShared,
		cliName, ModeShared,
		cliName, ModeSpinLock,
	)
	os.Exit(1)
}

func checkMode(mode string) bool {
	if ModeIsolation != mode && ModeShared != mode && ModeSpinLock != mode {
		return false
	}

	return true
}

func checkIteration(iteration string) bool {
	if tmp, err := strconv.ParseInt(iteration, 10, 32); nil == err && 0 < tmp {
		return true
	}

	return false
}

func checkRandomWriteRatio(ratio string) bool {
	if tmp, err := strconv.ParseFloat(ratio, 32); nil == err && 0.0 <= tmp && 100.0 >= tmp {
		return true
	}

	return false
}

func checkPath2Write(dirPath string) bool {
	if info, err := os.Stat(dirPath); nil == err && info.IsDir() {
		return true
	}

	return false
}

func consumeCPU(randSrc *rand.Rand, randomWriteRatio float64, fileHandle *os.File, bytes2Write []byte, lock *SpinLock) string {
	sha512Handle := sha512.New512_256()
	isWrite := false
	if 100.0 == randomWriteRatio || (randomWriteRatio/100.0) >= randSrc.Float64() {
		isWrite = true
	}

	for idx := 0; idx < NumRandomGen; idx++ {
		sha512Handle.Write([]byte(fmt.Sprintf("%f", randSrc.Float64())))
	}

	if isWrite {
		fileHandle.Seek(0, os.SEEK_SET)
		fileHandle.Write(bytes2Write)
	}

	if nil != lock {
		lock.Lock()
		defer lock.Unlock()
	}

	return fmt.Sprintf("%x", sha512Handle.Sum(nil))
}

func runBenchmark(mode string, numThreads int, numIteration int, randomWriteRatio float64, path2Write string, bytes2WriteSize int) {
	bytes2Write := make([]byte, bytes2WriteSize)
	for idx := 0; idx < bytes2WriteSize; idx++ {
		bytes2Write[idx] = byte(idx % 128)
	}

	startTime := time.Now()
	wg := &sync.WaitGroup{}
	if ModeIsolation == mode {
		for idx := 0; idx < numThreads; idx++ {
			wg.Add(1)
			go func(workerIdx int) {
				defer wg.Done()
				rndHandle := rand.New(rand.NewSource(time.Now().Unix() + rand.Int63n(0xFFFFFFF)))
				fileHandle, err := os.OpenFile(fmt.Sprintf("%s/%d", path2Write, workerIdx), os.O_RDWR|os.O_CREATE, 0755)
				if nil != err {
					fmt.Printf("failed to create file\n")
					return
				}

				for idxIter := 0; idxIter < numIteration; idxIter++ {
					consumeCPU(rndHandle, randomWriteRatio, fileHandle, bytes2Write, nil)
				}
			}(idx)
		}
	} else {
		sharedChan := make(chan string, numThreads)
		var lock *SpinLock
		if ModeSpinLock == mode {
			lock = &SpinLock{}
			fmt.Printf("SpinLock: Enabled\n")
		} else {
			fmt.Printf("SpinLock: Disabled\n")
		}

		for idx := 0; idx < numThreads; idx++ {
			wg.Add(1)
			go func(workerIdx int) {
				defer wg.Done()
				rndHandle := rand.New(rand.NewSource(time.Now().Unix() + rand.Int63n(0xFFFFFFF)))
				fileHandle, err := os.OpenFile(fmt.Sprintf("%s/%d", path2Write, workerIdx), os.O_RDWR|os.O_CREATE, 0755)
				if nil != err {
					fmt.Printf("failed to create file\n")
					return
				}

				for idxIter := 0; idxIter < numIteration; idxIter++ {
					if 0 != idxIter {
						<-sharedChan
					}

					tmp := consumeCPU(rndHandle, randomWriteRatio, fileHandle, bytes2Write, lock)
					sharedChan <- tmp
				}
			}(idx)
		}
	}

	wg.Wait()
	fmt.Printf("Time eclipsed: %s\n", time.Now().Sub(startTime))
}

func main() {
	if 6 != len(os.Args) ||
		!checkMode(os.Args[1]) ||
		!checkIteration(os.Args[2]) ||
		!checkRandomWriteRatio(os.Args[3]) ||
		!checkPath2Write(os.Args[4]) {
		usage()
	}

	numThreads := EnvNumThreadsDefault
	if tmp, err := strconv.ParseInt(os.Getenv(EnvNumThreads), 10, 32); nil == err && 0 < tmp {
		numThreads = int(tmp)
	}
	fmt.Printf("Num of threads = %d\n", numThreads)
	runtime.GOMAXPROCS(numThreads)

	numIteration := IterationDefault
	if tmp, err := strconv.ParseInt(os.Args[2], 10, 32); nil == err && 0 < tmp {
		numIteration = int(tmp)
	}
	fmt.Printf("Num of iteration = %d\n", numIteration)

	randomWriteRatio := RandomWriteRatioDefault
	if tmp, err := strconv.ParseFloat(os.Args[3], 64); nil == err && 0.0 <= tmp && 100.0 >= tmp {
		randomWriteRatio = tmp
	}
	fmt.Printf("RandowmWriteRatio = %f%%\n", randomWriteRatio)

	path2Write := os.Args[4]
	fmt.Printf("Write I/O Dir = %s\n", path2Write)

	bytes2WriteSize := Bytes2WriteSizeDefault
	if tmp, err := strconv.ParseInt(os.Args[5], 10, 32); nil == err && 0 < tmp {
		bytes2WriteSize = int(tmp)
	}
	fmt.Printf("Write I/O with Size = %d bytes\n", bytes2WriteSize)

	runBenchmark(os.Args[1], numThreads, numIteration, randomWriteRatio, path2Write, bytes2WriteSize)
}
