//go:build option

package main

import (
	"encoding/json"
	"fmt"
	"github.com/binaryphile/fluentfp/fluent"
	"io"
	"log"
	"net/http"
)

// main fetches posts from the post api as options
func main() {

}

// Post represents a post from the JSONPlaceholder API.
// Only the ID and Title fields are used.
type Post struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

// ToFriendlyPost returns a validated version of the post.
// If the title is empty, it is set to "No title".
func (p Post) ToValidatedPost() Post {
	if p.Title == "" {
		p.Title = "No title"
	}

	return p
}

// main retrieves a list of posts from the JSONPlaceholder API and prints the first three to stdout.
func main() {
	resp, err := http.Get("https://jsonplaceholder.typicode.com/posts")
	if err != nil {
		log.Fatal(err)
	}
	defer Close(resp.Body)

	// a fluent slice is derived from a regular slice, so it is usable wherever a regular slice is.
	// just use the fluent.SliceOf[T] type to declare it.
	var posts fluent.SliceOf[Post]
	if err = json.NewDecoder(resp.Body).Decode(&posts); err != nil {
		log.Fatal(err)
	}

	// the Map method returns the result of applying a supplied function to each element of the slice.
	// the return type is a new fluent slice with the same element type as the original.
	// the following validates the posts in the slice:
	posts = posts.Convert(Post.ToValidatedPost)

	// Post.ToValidatedPost is a method expression.
	// a method expression is a function with the same signature as the method it is named for,
	// but in the form Type.Method and with a first argument that is the receiver (in this case, a Post instance).
	// that means, for the element type of a fluent slice, any no-argument method such as ToValidatedPost()
	// is also a single-argument function.
	// Any Type.MethodNameHere can be used as an argument to Map.
	// https://go.dev/ref/spec#Method_expressions

	// Now print the first three posts.
	posts.TakeFirst(3).Each(PrintPost)
}

// Close closes closer.
// If it errors, the error is printed to stderr.
func Close(closer io.ReadCloser) {
	err := closer.Close()
	if err != nil {
		fmt.Println("error:", err)
	}
}

// PrintPost prints a friendly version of post to stdout.
// It could easily also have been a method on Post.
// The output looks like:
//
//	Post ID: 1, Title: sunt aut facere repellat provident ...
func PrintPost(post Post) {
	fmt.Println("Post ID:", post.ID, ", Title:", post.Title)
}
