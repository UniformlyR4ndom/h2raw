package h2raw

import (
	"crypto/tls"
	"io"
	"net"

	"golang.org/x/net/http2"
)


type H2Conn struct {
	Conn io.ReadWriteCloser
	Framer *http2.Framer
	PrefaceSent bool
	SettingsSent bool
}


func NewH2ConnPlain(endpoint string, allowIllegal bool) (*H2Conn, error) {
	c, err := net.Dial("tcp", endpoint)
	if err != nil {
		return nil, err
	}

	framer := http2.NewFramer(c, c)
	framer.AllowIllegalWrites = allowIllegal
	framer.AllowIllegalReads = allowIllegal
	h2c := H2Conn {
		Conn: c,
		Framer: framer,
	}

	return &h2c, nil
}


func NewH2ConnTLS(endpoint string, tlsConf *tls.Config, allowIllegal bool) (*H2Conn, error) {
	c, err := tls.Dial("tcp", endpoint, tlsConf)
	if err != nil {
		return nil, err
	}

	framer := http2.NewFramer(c, c)
	framer.AllowIllegalWrites = allowIllegal
	framer.AllowIllegalReads = allowIllegal
	h2c := H2Conn {
		Conn: c,
		Framer: framer,
	}

	return &h2c, nil
}


func (c *H2Conn) Close() error {
	return c.Conn.Close()
}


// Send HTTP2 preface and a simple SETTINGS frame
func (c *H2Conn) DefaultInit() error {
	if _, err := c.Conn.Write([]byte(http2.ClientPreface)); err != nil {
		return err
	}
	c.PrefaceSent = true

	defaultSettings := http2.Setting {
		ID:  http2.SettingMaxConcurrentStreams,
		Val: 100,
	}
	if err := c.Framer.WriteSettings(defaultSettings); err != nil {
		return err
	}
	c.SettingsSent = true

	return nil
}


func (c *H2Conn) IsIntialized() bool {
	return c.PrefaceSent && c.SettingsSent
}
