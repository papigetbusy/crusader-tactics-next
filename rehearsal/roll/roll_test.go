package roll

import (
	"reflect"
	"testing"
)

func TestSameSeedSameDice(t *testing.T) {
	a, b := Dice(123, 10, 20), Dice(123, 10, 20)
	if !reflect.DeepEqual(a, b) {
		t.Fatalf("same seed gave different dice: %v vs %v", a, b)
	}
}

func TestDifferentSeedDifferentDice(t *testing.T) {
	if reflect.DeepEqual(Dice(123, 10, 20), Dice(124, 10, 20)) {
		t.Fatal("different seeds gave identical dice")
	}
}

func TestDiceStayInRange(t *testing.T) {
	for _, v := range Dice(7, 1000, 6) {
		if v < 1 || v > 6 {
			t.Fatalf("die out of range: %d", v)
		}
	}
}
