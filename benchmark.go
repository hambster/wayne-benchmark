package main

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
	EnvNumThreads        = "NUM_THREADS"
	EnvNumThreadsDefault = 2
	ModeIsolation        = "isolation"
	ModeShared           = "shared"
	IterationDefault     = 10000
	NumRandomGen         = 256
)

func usage() {
	cliName := os.Args[0]
	fmt.Printf(`
	command:
		%s [%s|%s] [number-of-iteration]

	example:

		# populate 16 threads to stress CPU with 1000000 iteration without interacting with each other
		export NUM_THREADS=16
		%s isoloation 1000000

		# populate 16 threads to stress CPU with 1000000 iteration with interacting with each other
		export NUM_THREADS=16
		%s shared 1000000

		# populate 16 threads to stress CPU with 1000000 iteration with interacting with each other
		export NUM_THREADS=16
		taskset 0xFFF %s shared 1000000

`, cliName, ModeIsolation, ModeShared, cliName, cliName)
	os.Exit(1)
}

func checkMode(mode string) bool {
	if ModeIsolation != mode && ModeShared != mode {
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

func consumeCPU(randSrc *rand.Rand) string {
	sha512Handle := sha512.New512_256()
	for idx := 0; idx < NumRandomGen; idx++ {
		sha512Handle.Write(
			[]byte(fmt.Sprintf("%f", randSrc.Float64())),
		)
	}

	return fmt.Sprintf("%x", sha512Handle.Sum(nil))
}

func runBenchmark(mode string, numThreads int, numIteration int) {
	startTime := time.Now()
	wg := &sync.WaitGroup{}
	if ModeIsolation == mode {
		for idx := 0; idx < numThreads; idx++ {
			wg.Add(1)
			go func() {
				rndHandle := rand.New(rand.NewSource(time.Now().Unix()))
				for idxIter := 0; idxIter < numIteration; idxIter++ {
					consumeCPU(rndHandle)
				}
				wg.Done()
			}()
		}
	} else {
		sharedChan := make(chan string, numThreads)
		for idx := 0; idx < numThreads; idx++ {
			wg.Add(1)
			go func() {
				rndHandle := rand.New(rand.NewSource(time.Now().Unix()))
				for idxIter := 0; idxIter < numIteration; idxIter++ {
					if 0 != idxIter {
						<-sharedChan
					}

					tmp := consumeCPU(rndHandle)
					sharedChan <- tmp
				}
				wg.Done()
			}()
		}
	}

	wg.Wait()
	fmt.Printf("Time eclipsed: %s\n", time.Now().Sub(startTime))
}

func main() {
	if 3 != len(os.Args) || !checkMode(os.Args[1]) || !checkIteration(os.Args[2]) {
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

	runBenchmark(os.Args[1], numThreads, numIteration)
}
