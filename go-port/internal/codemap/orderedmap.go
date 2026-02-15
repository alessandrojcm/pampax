package codemap

import (
	"bytes"
	"encoding/json"
)

type OrderedMap struct {
	keys   []string
	values map[string]any
}

func NewOrderedMap() *OrderedMap {
	return &OrderedMap{
		keys:   make([]string, 0),
		values: make(map[string]any),
	}
}

func (o *OrderedMap) Set(key string, value any) {
	if o.values == nil {
		o.values = make(map[string]any)
	}

	if _, exists := o.values[key]; !exists {
		o.keys = append(o.keys, key)
	}

	o.values[key] = value
}

func (o *OrderedMap) Get(key string) (any, bool) {
	if o == nil || o.values == nil {
		return nil, false
	}

	v, ok := o.values[key]
	return v, ok
}

func (o *OrderedMap) Keys() []string {
	if o == nil {
		return []string{}
	}

	out := make([]string, len(o.keys))
	copy(out, o.keys)
	return out
}

func (o *OrderedMap) MarshalJSON() ([]byte, error) {
	if o == nil || len(o.keys) == 0 {
		return []byte("{}"), nil
	}

	var buf bytes.Buffer
	buf.WriteByte('{')

	for i, key := range o.keys {
		if i > 0 {
			buf.WriteByte(',')
		}

		keyBytes, err := json.Marshal(key)
		if err != nil {
			return nil, err
		}

		valueBytes, err := json.Marshal(o.values[key])
		if err != nil {
			return nil, err
		}

		buf.Write(keyBytes)
		buf.WriteByte(':')
		buf.Write(valueBytes)
	}

	buf.WriteByte('}')
	return buf.Bytes(), nil
}
