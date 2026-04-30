package httpserver

import (
	"strings"
	"testing"

	"github.com/iamlovingit/clawmanager-openclaw-image/internal/process"
)

func TestOpenClawWaitReadyDoesNotRequireGatewayWarmup(t *testing.T) {
	if !openClawWaitReady(process.Snapshot{Status: process.StatusRunning, GatewayWarmupReady: false}) {
		t.Fatal("running gateway should release wait page even while models warmup continues")
	}
}

func TestOpenClawWaitReadyDoesNotReleaseStartingGateway(t *testing.T) {
	if openClawWaitReady(process.Snapshot{Status: process.StatusStarting}) {
		t.Fatal("starting gateway should not release wait page before gateway readiness promotion")
	}
}

func TestOpenClawWaitReadyReleasesWhenWarmupStarted(t *testing.T) {
	if !openClawWaitReady(process.Snapshot{Status: process.StatusStarting, GatewayWarmupStarted: true}) {
		t.Fatal("started warmup should release wait page into bounded warmup wait")
	}
}

func TestOpenClawWaitPageStartsWarmupTimeoutAfterGatewayReady(t *testing.T) {
	page := openClawWaitPage("http://localhost:18789")
	if !strings.Contains(page, "let gatewayReadyAt = 0") {
		t.Fatal("wait page should track when the gateway first becomes ready")
	}
	if !strings.Contains(page, "gatewayReadyAt = Date.now()") {
		t.Fatal("wait page should start warmup timeout after gateway readiness")
	}
	if strings.Contains(page, "const startedAt = Date.now()") {
		t.Fatal("wait page should not start warmup timeout at page load")
	}
}
