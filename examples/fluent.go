//go:build fluent

package main

import (
	"cmp"
	"encoding/json"
	"fmt"
	"github.com/binaryphile/fluentfp/fluent"
	"net/http"
	"slices"
	"strings"
)

// main retrieves a list of posts from the JSONPlaceholder API and prints various details about them.
func main() {
	// get some posts from the REST endpoint with some standard boilerplate error handling etc.
	resp, err := http.Get("https://jsonplaceholder.typicode.com/posts")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// A fluent slice is derived from a regular slice, so it is usable in the same way wherever a regular slice is.
	// Just supply the element type in the declaration (fluent.SliceOf[element type]).
	var posts fluent.SliceOf[Post] // instead of []Post
	if err = json.NewDecoder(resp.Body).Decode(&posts); err != nil {
		panic(err)
	}

	// There are three names for Map-related methods:
	//  - To[BuiltinType] for returning a slice of string or int, for example
	//  - Convert for returning a slice of the same type as the original
	//  - Map for returning a slice of a specified type, usually a struct type
	// Shown here is the Convert method since we're making Posts from Posts.
	// The following validates posts and makes titles friendly by ensuring they are not empty:
	posts = posts.
		KeepIf(Post.IsValid). // KeepIf is a filter implementation
		Convert(Post.ToFriendlyPost)
	// Post.ToFriendlyPost takes the usual method receiver (a post in this case) as its first argument instead.
	// see https://go.dev/ref/spec#Method_expressions

	// for comparison to above:
	//
	//     friendlyPosts := make([]Post, 0, len(posts)) // do you make this correctly every time?  I don't.
	//     for _, post := range posts {                 // which form of this do you need, single or double assignment?
	//         if post.IsValid() {                      // do you need the index?  I have to think about it every time.
	//             friendlyPosts = append(friendlyPosts, post)
	//         }
	//     }
	//     posts = friendlyPosts

	// print the first three posts
	fmt.Println("the first three posts:")

	posts.
		TakeFirst(3).              // TakeFirst returns a slice of the first n elements
		ToStringWith(Post.String). // ToStringWith is Map to the named builtin type, string in this case
		Each(Println)              // Each applies the named function to each element for its side effects
	// for comparison to above:
	//
	//     for i, post := range posts { // again, which form of this?
	//         if i == 3 {              // three lines of code just to break, but simpler than C-style for loop
	//            break
	//         }
	//         fmt.Println(post.String())
	//     }

	// print the longest post title (if multiple, the first one found)
	fmt.Println("\nthe longest post title in words:")

	titles := posts.ToStringWith(Post.GetTitle)
	// for comparison to above:
	//
	//     titles := make([]string, len(posts))
	//     for i, post := range posts {
	//         titles[i] = post.Title
	//     }

	longestTitle := slices.MaxFunc(titles, CompareWordCounts) // fluent slices don't have Max or MaxFunc methods
	fmt.Println(longestTitle)

	// The Map method requires a type to map to, so SliceOf[T] doesn't have enough type parameters to support Map.
	// For this reason, there's an additional type, SliceWithMap.
	// Rather than go through creating a new type to demonstrate, we'll just specify string as the target type,
	// but it could be any type of your own.  Imagine your own type as the return type in this example.

	// first, type-convert our existing slice to a SliceWithMap
	mappablePosts := fluent.SliceWithMap[Post, string](posts)

	// now use Map to convert to strings
	fmt.Println("\nthe first three posts again!  encore!")
	mappablePosts.
		TakeFirst(3).
		Map(Post.String). // imagine to your type here instead
		Each(Println)

	// Because of Go's type system, there are a couple other
}

// Post type definition
///////////////////////

// Post represents a post from the JSONPlaceholder API.
type Post struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

// GetTitle returns the post's title.
func (p Post) GetTitle() string {
	return p.Title
}

// IsValid returns whether the ID is positive.
func (p Post) IsValid() bool {
	return p.ID > 0
}

// String generates a friendly, string version of p suitable for printing to stdout.
// The output looks like:
//
//	Post ID: 1, Title: sunt aut facere repellat provident
func (p Post) String() string {
	return fmt.Sprint("Post ID: ", p.ID, ", Title: ", p.Title)
}

// ToFriendlyPost returns p with title "No title" if p.Title is blank.
func (p Post) ToFriendlyPost() Post {
	if p.Title == "" {
		p.Title = "No title"
	}

	return p
}

// Functions
////////////

func CompareWordCounts(first, second string) int {
	return cmp.Compare(len(strings.Fields(first)), len(strings.Fields(second)))
}

// Println prints a string to stdout.
// Since fmt.Println has variadic `any` args, it can't be used directly with Each.
// This function wraps fmt.Println to make it usable with Each
// by changing the signature from variadic `any` arguments to a single argument of string.
func Println(s string) {
	fmt.Println(s)
}
