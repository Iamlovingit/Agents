package agent

import "runtime"

func goArch() string {
	return runtime.GOARCH
}
