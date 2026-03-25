package i18n

import "testing"

func TestFlattenJSON(t *testing.T) {
	t.Parallel()

	input := map[string]interface{}{
		"lang_name": "English",
		"TITLE": map[string]interface{}{
			"window_title": "LoD Config Editor",
		},
		"NON_STRING": map[string]interface{}{
			"value": 123,
		},
	}

	out := make(map[string]string)
	flattenJSON("", input, out)

	if out["lang_name"] != "English" {
		t.Fatalf("expected lang_name to be flattened")
	}
	if out["TITLE.window_title"] != "LoD Config Editor" {
		t.Fatalf("expected TITLE.window_title to be flattened")
	}
	if _, exists := out["NON_STRING.value"]; exists {
		t.Fatalf("non-string values should not be added to flattened output")
	}
}

func TestUnflattenJSON(t *testing.T) {
	t.Parallel()

	flat := map[string]string{
		"lang_name":          "English",
		"TITLE.window_title": "LoD Config Editor",
	}

	nested := unflattenJSON(flat)

	titleObj, ok := nested["TITLE"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected TITLE object")
	}
	if titleObj["window_title"] != "LoD Config Editor" {
		t.Fatalf("unexpected nested TITLE.window_title: %v", titleObj["window_title"])
	}
	if nested["lang_name"] != "English" {
		t.Fatalf("unexpected lang_name: %v", nested["lang_name"])
	}
}
