package i18n

import "strings"

func flattenJSON(prefix string, value interface{}, result map[string]string) {
	obj, ok := value.(map[string]interface{})
	if !ok {
		return
	}

	for k, v := range obj {
		newKey := k
		if prefix != "" {
			newKey = prefix + "." + k
		}

		switch val := v.(type) {
		case map[string]interface{}:
			flattenJSON(newKey, val, result)
		case string:
			result[newKey] = val
		}
	}
}

func unflattenJSON(flat map[string]string) map[string]interface{} {
	result := make(map[string]interface{})

	for k, v := range flat {
		keys := strings.Split(k, ".")
		current := result

		for i, key := range keys {
			if i == len(keys)-1 {
				current[key] = v
				continue
			}

			next, ok := current[key]
			if !ok {
				child := make(map[string]interface{})
				current[key] = child
				current = child
				continue
			}

			child, ok := next.(map[string]interface{})
			if !ok {
				break
			}
			current = child
		}
	}

	return result
}
