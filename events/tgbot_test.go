package events

import (
	"testing"

	"github.com/PaulSonOfLars/gotgbot/v2"
)

func TestEventTimingParagraphIncludesStartAndEndLabels(t *testing.T) {
	paragraph := eventTimingParagraph("Start", "End", "en", 1700000000, 1700003600)

	if len(paragraph) != 5 {
		t.Fatalf("expected 5 rich-text items, got %d", len(paragraph))
	}

	startLabel, ok := paragraph[0].(gotgbot.RichTextString)
	if !ok || string(startLabel) != "Start" {
		t.Fatalf("expected start label 'Start', got %#v", paragraph[0])
	}

	if _, ok := paragraph[1].(gotgbot.RichTextDateTime); !ok {
		t.Fatalf("expected start datetime at index 1, got %#v", paragraph[1])
	}

	endLabel, ok := paragraph[3].(gotgbot.RichTextString)
	if !ok || string(endLabel) != "End" {
		t.Fatalf("expected end label 'End', got %#v", paragraph[3])
	}

	if _, ok := paragraph[4].(gotgbot.RichTextDateTime); !ok {
		t.Fatalf("expected end datetime at index 4, got %#v", paragraph[4])
	}
}
