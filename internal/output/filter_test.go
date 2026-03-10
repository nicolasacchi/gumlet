package output

import (
	"encoding/json"
	"testing"
)

func TestApplyFilter_SimplePath(t *testing.T) {
	data := json.RawMessage(`[{"id": "1", "name": "test"}, {"id": "2", "name": "other"}]`)
	result, err := ApplyFilter(data, "#.name")
	if err != nil {
		t.Fatal(err)
	}
	want := `["test","other"]`
	if string(result) != want {
		t.Errorf("got %s, want %s", result, want)
	}
}

func TestApplyFilter_NestedPath(t *testing.T) {
	data := json.RawMessage(`{"source": {"id": "abc", "name": "test"}}`)
	result, err := ApplyFilter(data, "source.name")
	if err != nil {
		t.Fatal(err)
	}
	if string(result) != `"test"` {
		t.Errorf("got %s, want %q", result, "test")
	}
}

func TestApplyFilter_NonexistentPath(t *testing.T) {
	data := json.RawMessage(`{"id": "1"}`)
	result, err := ApplyFilter(data, "missing")
	if err != nil {
		t.Fatal(err)
	}
	if string(result) != "null" {
		t.Errorf("got %s, want null", result)
	}
}

func TestApplyFilter_EmptyPath(t *testing.T) {
	data := json.RawMessage(`{"id": "1"}`)
	result, err := ApplyFilter(data, "")
	if err != nil {
		t.Fatal(err)
	}
	if string(result) != string(data) {
		t.Errorf("got %s, want %s", result, data)
	}
}

func TestApplyFilter_ObjectExtraction(t *testing.T) {
	data := json.RawMessage(`[{"id": "1", "name": "a"}, {"id": "2", "name": "b"}]`)
	result, err := ApplyFilter(data, `#.{id:id,name:name}`)
	if err != nil {
		t.Fatal(err)
	}
	// Should return array of extracted objects
	var items []map[string]string
	if err := json.Unmarshal(result, &items); err != nil {
		t.Fatalf("failed to parse result: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("got %d items, want 2", len(items))
	}
}
