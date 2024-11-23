package h2raw

import (
	"bytes"

	"golang.org/x/net/http2/hpack"
	"golang.org/x/net/http2"
)

type Request struct {
	Method string
	Scheme string
	Authority string
	Path string

	Headers map[string][]string
	Body *[]byte
}


func NewRequest(method, scheme, authority, path string) *Request {
	r := Request{
		Method: method,
		Scheme: scheme,
		Authority: authority,
		Path: path,
		Headers: make(map[string][]string),
		Body: nil,
	}

	return &r
}


func (r *Request) SetHeader(key string, values []string) {
	r.Headers[key] = values
}


func (r *Request) AddHeader(key, value string) {
	if values, ok := r.Headers[key]; ok {
		newValues := append(values, value)
		r.Headers[key] = newValues
	} else {
		r.Headers[key] = []string{value}
	}
}


func (r *Request) HasBody() bool {
	return r.Body != nil
} 


func (r *Request) BodyLengh() int {
	if r.HasBody() {
		return len(*r.Body)
	} else {
		return -1
	}
}


// TODO: add prio
func (r *Request) Serialize(streamId uint32) ([]byte, error) {
	var headerBuf bytes.Buffer
	encoder := hpack.NewEncoder(&headerBuf)
	
	if err := encoder.WriteField(hpack.HeaderField{Name: ":method", Value: r.Method}); err != nil {
		return []byte{}, err
	}

	if err := encoder.WriteField(hpack.HeaderField{Name: ":scheme", Value: r.Scheme}); err != nil {
		return []byte{}, err
	}

	if err := encoder.WriteField(hpack.HeaderField{Name: ":path", Value: r.Path}); err != nil {
		return []byte{}, err
	}

	if err := encoder.WriteField(hpack.HeaderField{Name: ":authority", Value: r.Authority}); err != nil {
		return []byte{}, err
	}

	for k, values := range(r.Headers) {
		for _, v := range(values) {
			if err := encoder.WriteField(hpack.HeaderField{Name: k, Value: v}); err != nil {
				return []byte{}, err
			}
		}
	}

	var requestBuf bytes.Buffer
	var readBuf bytes.Buffer
	framer := http2.NewFramer(&requestBuf, &readBuf)
	headersParam := http2.HeadersFrameParam{
		StreamID:      streamId,
		BlockFragment: headerBuf.Bytes(),
		EndStream:     r.HasBody(),
		EndHeaders:    true,
	}

	if err := framer.WriteHeaders(headersParam); err != nil {
		return []byte{}, err
	}

	if r.HasBody() {
		if err := framer.WriteData(streamId, true, *r.Body); err != nil {
			return []byte{}, err
		}
	}
	
	return requestBuf.Bytes(), nil
}
