//go:build !linux && !darwin && !freebsd && !openbsd && !netbsd

package agent

func diskStats(path string) (total int64, free int64) {
	return 0, 0
}
