package prod

import "encoding/json"

// jsonUnmarshal is a tiny shim so callers don't have to import encoding/json
// when the only thing they want is "decode bytes into struct".
func jsonUnmarshal(raw []byte, v any) error {
	return json.Unmarshal(raw, v)
}
