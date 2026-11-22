package gcp
import (
	"testing"
)
func TestGetRegions_NonEmpty(t *testing.T) {
	regions := GetRegions()
	if len(regions) == 0 {
		t.Error("expected non-empty region list")
	}
}
func TestGetRegions_KeyRegionsPresent(t *testing.T) {
	regions := GetRegions()
	regionIDs := make(map[string]bool)
	for _, r := range regions {
		regionIDs[r.ID] = true
	}
	keyRegions := []string{
		"us-central1",
		"us-east1",
		"us-west1",
		"europe-west1",
		"europe-west3",
		"asia-east1",
		"asia-northeast1",
		"australia-southeast1",
		"southamerica-east1",
	}
	for _, id := range keyRegions {
		if !regionIDs[id] {
			t.Errorf("expected key region %s to be present", id)
		}
	}
}
func TestGetRegions_NoDuplicateIDs(t *testing.T) {
	regions := GetRegions()
	seen := make(map[string]bool)
	for _, r := range regions {
		if seen[r.ID] {
			t.Errorf("duplicate region ID: %s", r.ID)
		}
		seen[r.ID] = true
	}
}
func TestGetRegions_AllFieldsPopulated(t *testing.T) {
	regions := GetRegions()
	for _, r := range regions {
		if r.ID == "" {
			t.Error("found region with empty ID")
		}
		if r.Description == "" {
			t.Errorf("region %s has empty description", r.ID)
		}
	}
}
func TestGetRegions_HasContinentCoverage(t *testing.T) {
	regions := GetRegions()
	hasNorthAmerica := false
	hasSouthAmerica := false
	hasEurope := false
	hasAsia := false
	hasAustralia := false
	hasMiddleEast := false
	hasAfrica := false
	for _, r := range regions {
		switch {
		case len(r.ID) >= 2 && r.ID[:2] == "us" || len(r.ID) >= 12 && r.ID[:12] == "northamerica":
			hasNorthAmerica = true
		case len(r.ID) >= 12 && r.ID[:12] == "southamerica":
			hasSouthAmerica = true
		case len(r.ID) >= 6 && r.ID[:6] == "europe":
			hasEurope = true
		case len(r.ID) >= 4 && r.ID[:4] == "asia":
			hasAsia = true
		case len(r.ID) >= 9 && r.ID[:9] == "australia":
			hasAustralia = true
		case len(r.ID) >= 2 && r.ID[:2] == "me":
			hasMiddleEast = true
		case len(r.ID) >= 6 && r.ID[:6] == "africa":
			hasAfrica = true
		}
	}
	if !hasNorthAmerica {
		t.Error("expected North America regions")
	}
	if !hasSouthAmerica {
		t.Error("expected South America regions")
	}
	if !hasEurope {
		t.Error("expected Europe regions")
	}
	if !hasAsia {
		t.Error("expected Asia regions")
	}
	if !hasAustralia {
		t.Error("expected Australia regions")
	}
	if !hasMiddleEast {
		t.Error("expected Middle East regions")
	}
	if !hasAfrica {
		t.Error("expected Africa regions")
	}
}