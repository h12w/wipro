package wipro

import (
	"encoding/json"
)

func (b BNF) JSON() string {
	return toJSON(b)
}

func toJSON(v interface{}) string {
	buf, _ := json.MarshalIndent(v, "", "    ")
	return string(buf)
}
