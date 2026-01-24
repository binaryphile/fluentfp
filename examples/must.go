//go:build ignore

package main

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/binaryphile/fluentfp/lof"
	"github.com/binaryphile/fluentfp/must"
	"github.com/binaryphile/fluentfp/slice"
)

// This example demonstrates must — panic-on-error for invariants.
func main() {
	// === Environment Variables ===

	home := must.Getenv("HOME")
	fmt.Println("home:", home) // (varies by system)

	// === File Operations ===

	file := must.Get(os.Open(home))
	fmt.Println("opened:", file.Name()) // (varies by system)

	must.BeNil(file.Close())
	fmt.Println("closed file") // closed file

	// === HTTP Pipeline ===

	// urlFromID builds a JSONPlaceholder URL for the given post ID.
	urlFromID := func(id int) string {
		return fmt.Sprintf("http://jsonplaceholder.typicode.com/posts/%d", id)
	}

	// bodyFromResponse reads the response body as a string.
	bodyFromResponse := func(resp *http.Response) (_ string, err error) {
		defer resp.Body.Close()
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return
		}
		return string(bodyBytes), nil
	}

	var ids slice.MapperTo[*http.Response, int] = []int{1, 2}
	bodies := ids.
		ToString(urlFromID).
		To(must.Of(http.Get)).
		ToString(must.Of(bodyFromResponse))
	bodies.Each(lof.Println) // (2 JSON responses from external API)
}
