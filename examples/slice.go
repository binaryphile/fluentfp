//go:build ignore

package main

import (
	"cmp"
	"encoding/json"
	"fmt"
	"github.com/binaryphile/fluentfp/lof"
	"github.com/binaryphile/fluentfp/slice"
	"net/http"
	"slices"
	"strconv"
	"strings"
)

// This example assumes you are familiar with the standard map and filter functions from functional programming.
// If not, you may want to read up on them now:
// https://en.wikipedia.org/wiki/Map_(higher-order_function)
// https://en.wikipedia.org/wiki/Filter_(higher-order_function)
//
// Whenever this example refers to "map", I'm referring to the functional programming concept,
// not the Go built-in type.
//
// Also, method expressions are very useful with fluent slices.
// See https://go.dev/ref/spec#Method_expressions for details.

// main retrieves a list of posts from the JSONPlaceholder API and prints various details about them.
func main() {
	// get some posts from the REST endpoint
	resp, _ := http.Get("https://jsonplaceholder.typicode.com/posts")

	// A fluent slice is a named type with an underlying type of slice,
	// so it usable in all the same ways as slice.
	// Its definition is `type Mapper[T any] []T`.
	// The slice is also automatically converted between a regular slice and fluent slice
	// as you assign it to a variable of one type or the other.
	// That includes supplying them as arguments to function calls.
	// A function signature that expects a regular slice can receive a fluent slice
	// and implicitly convert it to regular, and vice versa.

	// From the Go language specification:

	// > A value x of type V is assignable to a variable of type T ("x is assignable to T")
	// > if [...] V and T have identical underlying types [...] and at least one of V or T is not a named type.

	// [Ed: slices are defined to have themselves as their underlying type]

	// https://go.dev/ref/spec#Types
	// https://go.dev/ref/spec#Underlying_types
	// https://go.dev/ref/spec#Assignability

	// create a fluent slice and decode the response into it
	var posts slice.Mapper[Post]              // supply the element type in the declaration
	json.NewDecoder(resp.Body).Decode(&posts) // Decode takes an `any` argument that happens to work with fluent slices

	// There are three names for the map-related methods:
	//
	//  - Convert for returning a slice of the same type as the original
	//  - To[Type]s for returning slices of basic built-in types, such as string or int (caveat: not all built-ins are covered)
	//  - To for returning a slice of a named type or a built-in not covered by To[Type]s
	//
	// Shown here is Convert since we're making posts from posts.

	// Frequently, a data source that is managed by others requires input validation and/or normalization.
	// You can do this easily with fluent slices.
	// Here we filter out invalid posts and normalize the titles of the rest.
	posts = posts.
		KeepIf(Post.IsValid). // KeepIf is a filter implementation
		Convert(Post.ToFriendlyPost)
	// Post.ToFriendlyPost takes the usual method receiver (a post in this case)
	// as its first (and only) regular argument instead.
	// See https://go.dev/ref/spec#Method_expressions.

	// for comparison to above:
	//
	// friendlyPosts := make([]Post, 0, len(posts)) // do you make this correctly every time?  I don't.
	// for _, post := range posts {                 // which form of this do you need, single or double assignment?
	//     if post.IsValid() {                      // do you need the index?  I have to think about it every time.
	//         friendlyPosts = append(friendlyPosts, post)
	//     }
	// }
	// posts = friendlyPosts  // now friendlyPosts is just hanging around, an unnecessary artifact

	// print the first three posts
	fmt.Println("the first three posts:")

	posts.
		TakeFirst(3).          // TakeFirst returns a slice of the first n elements
		ToString(Post.String). // ToString is map to the string built-in type
		Each(lof.Println)      // Each applies its argument, a function, to each element for its side effects
	// for comparison to above:
	//
	// for i, post := range posts { // again, which form of this? notice it's different this time.
	//     if i == 3 {              // three lines of code just to break, but simpler than a C-style for loop
	//        break
	//     }
	//     fmt.Println(post.String())
	// }

	// print the longest post title in words
	fmt.Println("\nthe longest post title in words:")

	titles := posts.ToString(Post.GetTitle)
	// for comparison to above:
	//
	// titles := make([]string, len(posts))
	// for i, post := range posts {
	//     titles[i] = post.Title
	// }

	// fluent slices don't have a method like MaxFunc,
	// so use the slices package
	longestTitle := slices.MaxFunc(titles, CompareWordCounts) // fluent slices can be regular slice arguments
	fmt.Println(longestTitle)

	// we'll use this function in our next example
	titleFromPost := func(post Post) Title {
		return Title(post.Title)
	}

	// A Mapper[T] only has one type parameter, the type of the elements, so it can't map to an arbitrary type.
	// For this reason, there's an additional type, MapperTo[R, T], where R is the return type of the method To.
	// The MapsTo function creates a MapperTo from a slice.
	first3Titles := slice.MapTo[Title](posts).
		TakeFirst(3).
		To(titleFromPost)

	fmt.Println("\ntitle lengths in characters of the first three posts:")
	first3Titles.
		ToInt(Title.Len).
		ToString(strconv.Itoa).
		Each(lof.Println) // type issues prevent using fmt.Println directly
}

// Title type definition

type Title string

func (t Title) Len() int {
	return len(t)
}

// Post type definition
///////////////////////

// Post represents a post from the JSONPlaceholder API.
type Post struct {
	ID    int
	Title string
}

// GetTitle returns the post's title.
func (p Post) GetTitle() string {
	return p.Title
}

// IsValid returns whether the post id is positive.
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

// ToFriendlyPost returns p with title "No title" if its title is blank.
func (p Post) ToFriendlyPost() Post {
	if p.Title == "" {
		p.Title = "No title"
	}

	return p
}

// Functions
////////////

// CompareWordCounts compares the number of words in two strings.
func CompareWordCounts(first, second string) int {
	return cmp.Compare(len(strings.Fields(first)), len(strings.Fields(second)))
}
