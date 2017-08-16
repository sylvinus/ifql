package promql

import (
	"log"
)

func NewQuery(promql string, opts ...Option) (string, error) {
	f, err := Parse("", []byte(promql), opts...)
	if err != nil {
		return "", err
	}
	log.Printf("%#+v", f)

	return f.(string), nil
}
