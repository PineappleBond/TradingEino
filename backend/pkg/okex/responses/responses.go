package responses

import (
	"encoding/json"
	"strconv"
)

// JSONInt is a custom type to handle both string and number JSON values
type JSONInt int

func (j *JSONInt) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as string first
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		// It's a string, parse it as int
		i, err := strconv.Atoi(s)
		if err != nil {
			return err
		}
		*j = JSONInt(i)
		return nil
	}

	// Try to unmarshal as number
	var i int
	if err := json.Unmarshal(data, &i); err == nil {
		*j = JSONInt(i)
		return nil
	}

	return nil
}

// Int converts JSONInt to int
func (j JSONInt) Int() int {
	return int(j)
}

type (
	Basic struct {
		Code JSONInt `json:"code"`
		Msg  string  `json:"msg,omitempty"`
	}
)
