package recent

import (
	"testing"

	C "github.com/libost/bandori-tg/constants"
)

func TestGetOngoingEventsIDHandlesShortEndAtSlices(t *testing.T) {
	oldRecent := Recent
	oldKeys := REventsKeys
	Recent = C.Recent{
		Events: map[string]C.EventData{
			"e1": {EndAt: []string{"", "10", "", ""}},
			"e2": {EndAt: []string{"", "20", "", ""}},
			"e3": {EndAt: []string{"", "30", "", ""}},
			"e4": {EndAt: []string{"", "40", "", ""}},
		},
	}
	REventsKeys = []string{"e1", "e2", "e3", "e4"}
	defer func() {
		Recent = oldRecent
		REventsKeys = oldKeys
	}()

	ids, err := GetOngoingEventsID()
	if err != nil {
		t.Fatalf("GetOngoingEventsID returned an error: %v", err)
	}
	if ids[1] != "e4" {
		t.Fatalf("expected latest EN event to be e4, got %q", ids[1])
	}
}
