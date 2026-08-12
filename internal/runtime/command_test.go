package runtime

import (
	"reflect"
	"testing"
)

func TestEnvironmentWithOverridesReplacesInheritedValue(t *testing.T) {
	got := environmentWithOverrides(
		[]string{"PATH=/usr/bin", "RADARR_API_KEY=stale", "FLAG"},
		map[string]string{"RADARR_API_KEY": "current"},
	)
	want := []string{"PATH=/usr/bin", "FLAG", "RADARR_API_KEY=current"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected one authoritative environment value %v, got %v", want, got)
	}
}
