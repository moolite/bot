package dicer

import (
	"testing"
)

func TestParse(t *testing.T) {
	dies := Parse("3d6")

	if len(dies) > 1 {
		t.Error("parsed more than one dice roll")
	}

	die := dies[0]
	if die.Keep != 0 {
		t.Error("keep should be 0")
	}
	if len(die.Results) != 3 {
		t.Error("results should be len 3")
	}
	if die.Total == 0 {
		t.Error("total should be greater than 0")
	}
	if die.Operator != Add {
		t.Error("operator should be ", Add)
	}
	if die.Number != 3 {
		t.Error("expected number of throws is 3, ", die.Number)
	}
	if die.Sides != 6 {
		t.Error("expected number of sides is 6, ", die.Sides)
	}
}

func TestParseMulti(t *testing.T) {
	dies := Parse("3d6 4d8 8d4")

	if len(dies) != 3 {
		t.Error("parser failed to parse 3 dies")
	}
}

func TestParseKeep(t *testing.T) {
	dies := Parse("4d6k3")

	if len(dies) != 1 {
		t.Error("parser failed to parse 1 die")
	}

	die := dies[0]
	if die.Keep != 3 {
		t.Error("failed to parse keep value: ", die.Keep)
	}
	if len(die.Results) != 3 {
		t.Error("kept dies are more or less than the requested number: ", die.Results)
	}
	if len(die.Removed) != 1 {
		t.Error("removed dies are more or less than requested: ", die.Removed)
	}
}

func TestParseOpMod(t *testing.T) {
	dies := Parse("2d6+4 2d6-4")

	if len(dies) != 2 {
		t.Error("parser failed to parse 1 die")
	}

	die := dies[0]
	if die.Operator != Add {
		t.Error("failed to parse operator: ", die.Operator)
	}
	if die.Mod != 4 {
		t.Error("modificator parse error ", die.Mod)
	}

	die = dies[1]
	if die.Operator != Subtract {
		t.Error("failed to parse operator: ", die.Operator)
	}
	if die.Mod != 4 {
		t.Error("modificator parse error ", die.Mod)
	}
}

func TestRoll(t *testing.T) {
	die := &Dice{
		Operator: Add,
		Number:   6,
		Sides:    1,
	}
	die.roll()

	if die.Total != 6 {
		t.Error("rolling one sided dice unexpected result: ", die.Total)
	}

	die.Total = 0
	die.Mod = 10
	die.Operator = Add
	die.roll()
	if die.Total != 16 {
		t.Error("modifier Add is not applied")
	}

	die.Total = 0
	die.Mod = 10
	die.Operator = Subtract
	die.roll()
	if die.Total != -4 {
		t.Error("modifier Subtract is not applied: ", die.Total)
	}
}
