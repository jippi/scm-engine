package gitlab

import (
	"encoding/json"
	"io"
)

type ContextLabels []ContextLabel

func (l ContextLabels) MarshalGQL(writer io.Writer) {
	data, err := json.Marshal(l)
	if err != nil {
		panic(err)
	}

	writer.Write(data)
}

func (l *ContextLabels) UnmarshalGQL(v interface{}) error {
	return nil
}
