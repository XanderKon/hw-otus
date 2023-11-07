package main

import (
	"bufio"
	"fmt"
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
	scanner := bufio.NewScanner(c.in)
	for scanner.Scan() {
		_, err := c.conn.Write([]byte(scanner.Text() + "\n"))
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *client) Receive() error {
	scanner := bufio.NewScanner(c.conn)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
		_, err := c.out.Write([]byte(scanner.Text() + "\n"))
		if err != nil {
			return err
		}
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
