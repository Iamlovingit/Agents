//go:build linux

package bootstrap

import (
	"log"
	"syscall"
)

const openFileLimitTarget uint64 = 65535

func raiseOpenFileLimit() {
	var current syscall.Rlimit
	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &current); err != nil {
		log.Printf("bootstrap: read nofile limit failed: %v", err)
		return
	}
	if current.Cur >= openFileLimitTarget {
		return
	}

	desired := current
	desired.Cur = openFileLimitTarget
	if desired.Max < desired.Cur {
		desired.Max = desired.Cur
	}
	if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &desired); err == nil {
		log.Printf("bootstrap: raised nofile limit from %d/%d to %d/%d", current.Cur, current.Max, desired.Cur, desired.Max)
		return
	}

	if current.Max > current.Cur {
		fallback := current
		fallback.Cur = current.Max
		if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &fallback); err == nil {
			log.Printf("bootstrap: raised nofile limit from %d/%d to %d/%d", current.Cur, current.Max, fallback.Cur, fallback.Max)
			return
		}
	}

	log.Printf("bootstrap: leaving nofile limit at %d/%d", current.Cur, current.Max)
}
