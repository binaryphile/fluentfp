//go:build ignore

package main

import (
	"fmt"
	"github.com/binaryphile/fluentfp/must"
	"os"
)

func main() {
	// get the contents of the environment variable $HOME,
	// panic if it is empty or not set
	home := must.Getenv("HOME")
	fmt.Println("got", home, "for $HOME")

	// panic if os.File returns an error
	file := must.Get(os.Open(home + "/.profile"))
	fmt.Println("opened file")

	// panic if there is an error on close
	err := file.Close()
	must.BeNil(err)
	fmt.Println("closed file")

	// show a panic
	must.BeNil(fmt.Errorf("this will panic"))
}
