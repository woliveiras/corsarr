package hostreadiness

import "testing"

func TestEvaluateAcceptsSupportedMac(t *testing.T) {
	status := Evaluate(Facts{
		Platform: "darwin", Architecture: "arm64", OSVersion: "26.5.2",
		MemoryBytes: 24 * 1024 * 1024 * 1024, FreeDiskBytes: 20 * 1024 * 1024 * 1024,
	})
	if !status.Ready || len(status.Issues) != 0 {
		t.Fatalf("unexpected readiness %#v", status)
	}
}

func TestEvaluateReportsEveryBlockingRequirement(t *testing.T) {
	status := Evaluate(Facts{
		Platform: "darwin", Architecture: "386", OSVersion: "13.6",
		MemoryBytes: 2 * 1024 * 1024 * 1024, FreeDiskBytes: 3 * 1024 * 1024 * 1024,
	})
	if status.Ready || len(status.Issues) != 4 {
		t.Fatalf("unexpected readiness %#v", status)
	}
}

func TestEvaluateRejectsUnmeasurableMacFacts(t *testing.T) {
	status := Evaluate(Facts{Platform: "darwin", Architecture: "arm64"})
	if status.Ready || len(status.Issues) != 3 {
		t.Fatalf("unexpected readiness %#v", status)
	}
}
