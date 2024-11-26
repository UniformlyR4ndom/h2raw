package h2raw

import (
    "bytes"
    "crypto/tls"
    "fmt"
    "log"
    "net"
    "strconv"

    "golang.org/x/net/http2"
    "golang.org/x/net/http2/hpack"
)

type ClientType uint8

const (
    ClientTypePlain = iota
    ClientTypeTLS
)


type Client struct {
    TLSConfig *tls.Config
    clientType ClientType
    AllowIllegal bool
    Verbose bool
}


func NewPlainClient() (*Client, error) {
    c := Client {
        clientType: ClientTypePlain,
    }

    return &c, nil
}


func NewTLSClient(tlsConfig *tls.Config) (*Client, error) {
    c := Client {
        TLSConfig:  tlsConfig,
        clientType: ClientTypeTLS,
    }

    return &c, nil
}


func NewTLSInsecureClient(endpoint string) (*Client, error) {
    tlsConfig := tls.Config {
        InsecureSkipVerify: true,
        MinVersion: tls.VersionTLS10,
        MaxVersion: tls.VersionTLS13,
        NextProtos: []string{"h2"},
    }

    c := Client {
        TLSConfig:  &tlsConfig,
        clientType: ClientTypeTLS,
    }

    return &c, nil
}


func (c *Client) NewConnection(endpoint string) (*H2Conn, error) {
    ep, err := validateEndpoint(endpoint)
    if err != nil {
        return nil, err
    }

    if c.clientType == ClientTypePlain {
        return NewH2ConnPlain(ep, c.AllowIllegal)
    } else if c.clientType == ClientTypeTLS {
        return NewH2ConnTLS(ep, c.TLSConfig, c.AllowIllegal)
    } else {
        return nil, fmt.Errorf("Unknown client type %v", c.clientType)
    }
}


func (c *Client) MakeRequest(h2conn *H2Conn, req *Request, streamId uint32) (*Response, error) {
    if !h2conn.IsIntialized() {
        if err := h2conn.DefaultInit(); err != nil {
            return nil, err
        }
    }

    headerBlockBuf := bytes.Buffer{}
    encoder := hpack.NewEncoder(&headerBlockBuf)
    if err := encoder.WriteField(hpack.HeaderField{Name: ":method", Value: req.Method}); err != nil {
        return nil, err
    }

    if err := encoder.WriteField(hpack.HeaderField{Name: ":scheme", Value: req.Scheme}); err != nil {
        return nil, err
    }

    if err := encoder.WriteField(hpack.HeaderField{Name: ":authority", Value: req.Authority}); err != nil {
        return nil, err
    }

    if err := encoder.WriteField(hpack.HeaderField{Name: ":path", Value: req.Path}); err != nil {
        return nil, err
    }

    for k, values := range(req.Headers) {
        for _, v := range(values) {
            if err := encoder.WriteField(hpack.HeaderField{Name: k, Value: v}); err != nil {
                return nil, err
            }
        }
    }

    headers := http2.HeadersFrameParam{
        StreamID:      streamId,
        BlockFragment: headerBlockBuf.Bytes(),
        EndStream:     !req.HasBody(),
        EndHeaders:    true,
    }

    framer := h2conn.Framer
    if err := framer.WriteHeaders(headers); err != nil {
        return nil, err
    }

    if req.HasBody() {
        if err := framer.WriteData(streamId, true, *req.Body); err != nil {
            return nil, err
        }
    }

    return c.readResponse(h2conn)
}


func (c *Client) readResponse(h2conn *H2Conn) (*Response, error) {
    framer := h2conn.Framer
    hasBody := false
    body := []byte{}
    headers := make(map[string][]string)
    loop:
    for {
        frame, err := framer.ReadFrame()
        if err != nil {
            c.logVerbose(fmt.Sprintf("readResponse error: %v\n", err))
            return nil, err
        }

        c.logVerbose(fmt.Sprintf("Received frame: %v", frame))
        switch f := frame.(type) {
        case *http2.HeadersFrame:
            newHeaders, err := parseHeaders(f.HeaderBlockFragment())
            if err != nil {
                return nil, err
            }

            addHeaders(headers, newHeaders)
            if f.StreamEnded() {
                break loop
            }

        case *http2.DataFrame:
            tmp := f.Data()
            body = append(body, tmp...)
            hasBody = true

            if f.StreamEnded() {
                break loop
            }

        case *http2.GoAwayFrame:
            return nil, fmt.Errorf("Received GoAway frame")

        case *http2.RSTStreamFrame:
            return nil, fmt.Errorf("Received RSTStreamFrame frame")

        default:
            c.logVerbose(fmt.Sprintf("Ignoring frame: %T\n", frame))
        }
    }

    status := ""
    if s, ok := headers[":status"]; ok {
        status = s[0]
    }

    var b *[]byte
    if hasBody {
        b = &body
    }

    resp := Response {
        Status: status,
        Headers: headers,
        Body: b,
    }

    return &resp, nil
}


func parseHeaders(blockFragment []byte) (map[string][]string, error) {
    headers := make(map[string][]string)
    decoder := hpack.NewDecoder(4096, nil)

    hvals, err := decoder.DecodeFull(blockFragment)
    if err != nil {
        return headers, err
    }

    for _, h := range(hvals) {
        if vals, ok := headers[h.Name]; ok {
            newVals := append(vals, h.Value)
            headers[h.Name] = newVals
        } else {
            headers[h.Name] = []string{h.Value}
        }
    }

    return headers, nil
}


func addHeaders(headers, additionalHeaders map[string][]string) {
    for k, v := range(additionalHeaders) {
        hvals, ok := headers[k]
        if ok {
            headers[k] = append(hvals, v...)
        } else {
            headers[k] = v
        }
    }
}


func validateEndpoint(epStr string) (string, error) {
    host, port, err := net.SplitHostPort(epStr)
    if err != nil {
        return "", err
    }

    nPort, err := strconv.Atoi(port)
    if err != nil {
        return "", err
    }

    return fmt.Sprintf("%v:%v", host, nPort), nil
}


func (c *Client) logVerbose(msg string) {
    if c.Verbose {
        log.Printf(msg)
    }
}
