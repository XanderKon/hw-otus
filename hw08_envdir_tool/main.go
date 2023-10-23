package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
)

var ErrNotEnoughArguments = errors.New("not enough arguments")

func main() {
	flag.Parse()
	args := flag.Args()

	if len(args) < 2 {
		fmt.Println(ErrNotEnoughArguments.Error())
		return
	}

	env, err := ReadDir(args[0])
	if err != nil {
		fmt.Println(err)
		return
	}

	cmd := RunCmd(args[1:], env)
	os.Exit(cmd)
}
