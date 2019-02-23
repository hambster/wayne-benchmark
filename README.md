# Wayne's System Benchmark

System Benchmark Tool to Generate CPU Utilization and I/O.

## Usage and Examples


	command:
		./wayne-benchmark [isolation|shared|spinlock] [number-of-iteration] [random-write-ratio] [path-to-write] [bytes-to-write]

	note:
		the maximum value of random-write-ratio is 100, which means you will definite write data in every iteration.

	example:

		# populate 16 threads to stress CPU with 1000000 iteration without interacting with each other
		# and generate 20% random write 1KB data in /tmp with
		export NUM_THREADS=16
		./wayne-benchmark isolation 1000000 20 /tmp 1024

		# populate 16 threads to stress CPU with 1000000 iteration with interacting with each other
		# and generate 30% random write 2KB in /mnt/resource
		export NUM_THREADS=16
		./wayne-benchmark shared 1000000 30 /mnt/resource 2048

		# populate 16 threads to stress CPU with 1000000 iteration with interacting with each other
		# and generate 90% random write 2KB in /tmp
		export NUM_THREADS=16
		taskset 0xFFF ./wayne-benchmark shared 1000000 90 /tmp 2048

		# populate 16 threads to stress CPU with 1000000 iteration with interacting with each other using spinlock
		# and generate 90% random write 2KB in /tmp
		export NUM_THREADS=16
		taskset 0xFFF ./wayne-benchmark spinlock 1000000 90 /tmp 2048
