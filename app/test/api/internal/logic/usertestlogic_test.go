package logic

import (
	"MuXiFresh-Be-2.0/app/test/api/internal/types"
	"testing"
)

func TestExamTruncatesOverlongChoice(t *testing.T) {
	choice := make([]types.ChoiceItem, 100)
	for i := range choice {
		choice[i] = types.ChoiceItem{Number: int64(i + 1), Data: "A"}
	}

	_, c := Exam(choice)

	if len(c) != 85 {
		t.Fatalf("expected choice truncated to 85, got %d", len(c))
	}
}

func TestExamKeepsNormalChoice(t *testing.T) {
	choice := make([]types.ChoiceItem, 5)
	for i := range choice {
		choice[i] = types.ChoiceItem{Number: int64(i + 1), Data: "A"}
	}

	_, c := Exam(choice)

	if len(c) != 5 {
		t.Fatalf("expected 5 choices unchanged, got %d", len(c))
	}
}
