package main

import (
	"fmt"
	"github.com/binaryphile/fluentfp/must"
	"os"
)

func main() {
	// get the contents of the environment variable $HOME,
	// panic if it is empty or not set
	homeEnv := must.Getenv("HOME")
	fmt.Println("got", homeEnv, "for $HOME")

	// panic if os.File returns an error
	file := must.Get(os.Open(homeEnv + "/.bashrc"))
	fmt.Println("opened file")

	// panic if there is an error on close
	err := file.Close()
	must.BeNil(err)
	fmt.Println("closed file")
}
