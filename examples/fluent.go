///go:build fluent

package main

import (
	"cmp"
	"encoding/json"
	"fmt"
	"github.com/binaryphile/fluentfp/fluent"
	"io"
	"net/http"
	"slices"
	"strings"
)

// main retrieves a list of posts from the JSONPlaceholder API and prints various details about them.
func main() {
	resp, err := http.Get("https://jsonplaceholder.typicode.com/posts")
	if err != nil {
		panic(err)
	}
	defer Close(resp.Body)

	// a fluent slice is derived from a regular slice, so it is usable in the same way wherever a regular slice is.
	// just supply the element type in the declaration (fluent.SliceOf[element type]).
	var posts fluent.SliceOf[Post]
	if err = json.NewDecoder(resp.Body).Decode(&posts); err != nil {
		panic(err)
	}

	// the Convert method returns the result of applying a supplied function to each element of the slice.
	// the return type is a new fluent slice with the same element type as the original.
	// the following validates posts and makes post titles friendly by ensuring the title is not empty:
	posts = posts.
		KeepIf(Post.IsValid).
		Convert(Post.ToFriendlyPost)
	// for comparison to above:
	//
	//     friendlyPosts := make([]Post, 0, len(posts)) // do you make this correctly every time?  I don't.
	//     for _, post := range posts {                 // which form of this do you need, single or double assignment?
	//         if post.IsValid() {                      // do you need the index?  I have to think about it every time.
	//             friendlyPosts = append(friendlyPosts, post)
	//         }
	//     }
	//     posts = friendlyPosts

	// Post.ToFriendlyPost takes the usual method receiver (a post in this case) as its first argument instead.
	// see https://go.dev/ref/spec#Method_expressions
	//
	// There are methods for mapping to builtin types as well, e.g.:
	//
	//   postStrings := posts.ConvertToString(Post.String)

	// print the first three posts
	fmt.Println("the first three posts:")

	posts.
		Take(3).
		ConvertToString(Post.String).
		Each(Println)
	// for comparison to above:
	//
	//     for i, post := range posts { // again, which form of this?
	//         if i == 3 {              // three lines just to break, but simpler than old-style for loop
	//            break
	//         }
	//         fmt.Println(post.String())
	//     }

	// print the longest post title
	fmt.Println("\nthe longest post title in words:")

	titles := posts.ConvertToString(Post.GetTitle)
	longestTitle := slices.MaxFunc(titles, CompareWordCounts)
	// for comparison to above:
	//
	//     titles := make([]string, len(posts))
	//     for i, post := range posts {
	//         titles[i] = post.Title
	//     }
	//     longestTitle := slices.MaxFunc(titles, CompareWordCounts)

	fmt.Println(longestTitle)
}

// Post struct definition
/////////////////////////

// Post represents a post from the JSONPlaceholder API.
// despite the fact that there are more fields available,
// only the ID and Title fields are used here.
type Post struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

// GetTitle returns the post's title.
func (p Post) GetTitle() string {
	return p.Title
}

// IsValid returns whether the title is not empty and the ID is positive.
func (p Post) IsValid() bool {
	return p.ID > 0
}

// String generates a friendly, string version of p suitable for printing to stdout.
// The output looks like:
//
//	Post ID: 1, Title: sunt aut facere repellat provident ...
func (p Post) String() string {
	return fmt.Sprint("Post ID: ", p.ID, ", Title: ", p.Title)
}

// ToFriendlyPost returns p with title "No title" if p.Title is blank.
func (p Post) ToFriendlyPost() Post {
	if !p.IsValid() {
		panic("not valid")
	}

	if p.Title == "" {
		p.Title = "No title"
	}

	return p
}

// functions
////////////

// Close closes closer.
// if it errors, the error is printed to stderr.
func Close(closer io.ReadCloser) {
	err := closer.Close()
	if err != nil {
		fmt.Println("error:", err)
	}
}

func CompareWordCounts(first, second string) int {
	return cmp.Compare(len(strings.Fields(first)), len(strings.Fields(second)))
}

func Println(s string) {
	fmt.Println(s)
}
