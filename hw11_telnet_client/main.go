package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var timeout time.Duration

func main() {
	flag.DurationVar(&timeout, "timeout", time.Second*10, "--timeout=10s")
	flag.Parse()

	args := flag.Args()

	if len(args) == 0 {
		fmt.Println("Not enough arguments!")
		fmt.Println("Usage: go-telnet [--timeout=10s] host port")
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGQUIT, syscall.SIGINT)
	defer stop()

	address := net.JoinHostPort(args[0], args[1])

	writer := bufio.NewWriter(os.Stdout)
	reader := io.NopCloser(bufio.NewReader(os.Stdin))

	client := NewTelnetClient(address, timeout, reader, writer)
	err := client.Connect()
	if err != nil {
		fmt.Println("Error client connection!")
		return
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := client.Receive()
		if err != nil {
			fmt.Println("...EOF")
			os.Exit(0)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		err := client.Send()
		if err != nil {
			fmt.Println("...Connection was closed by peer")
			os.Exit(0)
		}
	}()

	go func() {
		<-ctx.Done()
		fmt.Println("...Received SIGINT, closing connection")
		client.Close()
		os.Exit(0)
	}()

	wg.Wait()

	defer client.Close()
}
