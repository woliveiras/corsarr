package quality

import "testing"

func TestManagedPresetCatalogHasPinnedARRMappingsAndSummaries(t *testing.T) {
	if PresetCatalogVersion == "" {
		t.Fatal("preset catalog must be versioned")
	}
	for _, preset := range Presets() {
		if preset.Description == "" || preset.Summary == "" {
			t.Fatalf("preset is not understandable in onboarding: %#v", preset)
		}
		if preset.ID == PresetUnmanaged {
			if preset.RadarrTrashID != "" || preset.SonarrTrashID != "" {
				t.Fatalf("unmanaged preset claims an upstream mapping: %#v", preset)
			}
			continue
		}
		if preset.RadarrTrashID == "" || preset.SonarrTrashID == "" {
			t.Fatalf("managed preset is missing an Arr mapping: %#v", preset)
		}
	}
}
