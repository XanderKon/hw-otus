package main

import (
	"io"
	"net"
	"time"
)

type client struct {
	address string
	timeout time.Duration
	in      io.ReadCloser
	out     io.Writer
	conn    net.Conn
	TelnetClient
}

type TelnetClient interface {
	Connect() error
	io.Closer
	Send() error
	Receive() error
}

func (c *client) Connect() error {
	conn, err := net.DialTimeout("tcp", c.address, c.timeout)
	if err != nil {
		return err
	}

	c.conn = conn
	return nil
}

func (c *client) Send() error {
	_, err := io.Copy(c.conn, c.in)
	if err != nil {
		return err
	}

	return nil
}

func (c *client) Receive() error {
	_, err := io.Copy(c.out, c.conn)
	if err != nil {
		return err
	}

	return nil
}

func (c *client) Close() error {
	return c.conn.Close()
}

func NewTelnetClient(address string, timeout time.Duration, in io.ReadCloser, out io.Writer) TelnetClient {
	client := &client{
		address: address,
		timeout: timeout,
		in:      in,
		out:     out,
	}

	return client
}
