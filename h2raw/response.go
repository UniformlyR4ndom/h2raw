package h2raw

import (
	"strings"
)


type Response struct {
	Status string
	Proto string

	Headers map[string][]string
	Body *[]byte
}


func (r *Response) HasBody() bool {
	return r.Body != nil
}

func (r *Response) BodyLength() int {
	if r.HasBody() {
		return len(*r.Body)
	} else {
		return -1
	}
}


func (r *Response) GetHeaders(key string) ([]string, bool) {
	val, ok := r.Headers[key]
	return val, ok
}


func (r *Response) GetHeaderCombineCases(key, separator string) ([]string, bool) {
	key = strings.ToLower(key)
	values := []string{}
	for k, values := range(r.Headers) {
		if strings.ToLower(k) == key {
			values = append(values, values...)
		}
	}

	has := len(values) > 0
	return values, has
}
