//go:build ignore

package main

import (
	"fmt"
	"github.com/binaryphile/fluentfp/fluent"
	"github.com/binaryphile/fluentfp/hof"
	"github.com/binaryphile/fluentfp/must"
	"io"
	"net/http"
	"os"
)

func main() {
	// get the contents of the environment variable $HOME,
	// panic if it is empty or not set
	home := must.Getenv("HOME")
	fmt.Println("got", home, "for $HOME")

	// panic if os.Open returns an error
	file := must.Get(os.Open(home + "/.profile")) // TODO: need universal example
	fmt.Println("opened file")

	// panic if there is an error on close
	err := file.Close()
	must.BeNil(err) // you could call this on file.Close directly, but assigning to err is more readable
	fmt.Println("closed file")

	// consume fallible functions like http.Get and stringFromResponseBody using must.Of
	urlFromID := func(id int) string {
		return fmt.Sprintf("http://jsonplaceholder.typicode.com/posts/%d", id)
	}

	stringFromResponseBody := func(resp *http.Response) (_ string, err error) {
		defer resp.Body.Close()

		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return
		}

		return string(bodyBytes), nil
	}

	// print some posts
	var ids fluent.SliceToNamed[int, *http.Response] = []int{1, 2}
	ids.
		ToStrings(urlFromID).
		ToNamed(must.Of(http.Get)).
		ToString(must.Of(stringFromResponseBody)).
		Each(hof.Println)

	// show a panic
	must.BeNil(fmt.Errorf("this will panic"))
}
