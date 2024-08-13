package zipg

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

type MapJSON map[string]any

func (a MapJSON) Value() (driver.Value, error) {
	return json.Marshal(a)
}

func (a *MapJSON) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &a)
}
