package save

import (
	"encoding/json"
	"testing"
)

func TestASaveStillOpensWhenTheScreenStoresItsSettingsOnIt(t *testing.T) {
	b, err := Wrap("climb", map[string]int{"floor": 3})
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	m["screen"] = json.RawMessage(`{"flat":false,"danger":true,"minimap":false}`)
	b, err = json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]int
	kind, err := Open(b, &got)
	if err != nil {
		t.Fatal(err)
	}
	if kind != "climb" || got["floor"] != 3 {
		t.Fatalf("kind %q floor %d", kind, got["floor"])
	}
}
