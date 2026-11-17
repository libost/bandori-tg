package cards

import (
	"strings"
	"testing"
)

func TestCleanSVGXMLExpandsClassStyles(t *testing.T) {
	input := `<svg viewBox="0 0 10 10"><defs><style>.accent { fill:#ff345a; } .light { fill:white; }</style></defs><circle class="accent" cx="5" cy="5" r="5"/><path class="light" d="M0 0h1v1z"/></svg>`

	cleaned, err := cleanSVGXML(input)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(cleaned, `fill="#ff345a"`) {
		t.Fatalf("class fill was not expanded: %s", cleaned)
	}
	if !strings.Contains(cleaned, `fill="white"`) {
		t.Fatalf("second class fill was not expanded: %s", cleaned)
	}
	if strings.Contains(cleaned, "<style") || strings.Contains(cleaned, `class="`) {
		t.Fatalf("CSS-only markup was not removed: %s", cleaned)
	}
}

func TestCleanSVGXMLExpandsInlineStyles(t *testing.T) {
	input := `<svg viewBox="0 0 1 1"><path style="fill:#123456;stroke:none" d="M0 0h1v1z"/></svg>`

	cleaned, err := cleanSVGXML(input)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(cleaned, `fill="#123456"`) || !strings.Contains(cleaned, `stroke="none"`) {
		t.Fatalf("inline styles were not expanded: %s", cleaned)
	}
}

func TestCleanSVGXMLDisablesZeroWidthStrokes(t *testing.T) {
	input := `<svg viewBox="0 0 1 1"><path style="fill:#123456;stroke:#000000;stroke-width:0" d="M0 0h1v1z"/></svg>`

	cleaned, err := cleanSVGXML(input)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(cleaned, `stroke="none"`) {
		t.Fatalf("zero-width stroke was not disabled: %s", cleaned)
	}
}
