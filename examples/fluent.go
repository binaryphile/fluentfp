package main

import (
	"cmp"
	"encoding/json"
	"fmt"
	"github.com/binaryphile/fluentfp/fluent"
	"github.com/binaryphile/fluentfp/hof"
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
// Also, method expressions are very useful with fluent slices.
// See https://go.dev/ref/spec#Method_expressions for details.

// main retrieves a list of posts from the JSONPlaceholder API and prints various details about them.
// We'll skip checking errors and closing the response body for this example, even though you should in practice.
func main() {
	// get some posts from the REST endpoint
	resp, _ := http.Get("https://jsonplaceholder.typicode.com/posts")

	// A fluent slice is derived from a regular slice, so it usable in most places the slice type is.
	// Just supply the element type in the declaration, i.e. fluent.SliceOf[element type].
	var posts fluent.SliceOf[Post]
	json.NewDecoder(resp.Body).Decode(&posts)

	// There are three names for (functional programming) map-related methods:
	//  - To[Type]sWith for returning builtin types such as string or int
	//  - Convert for returning a slice of the same type as the original
	//  - MapWith for returning a slice of a specified type, usually a defined struct type
	// Shown here is the Convert method since we're making Posts from Posts.

	// Frequently a data source that is managed by others requires input validation and/or normalization.
	// You can do this easily with fluent slices.
	// Here we filter out invalid posts and normalize the titles of the rest.
	posts = posts.
		KeepIf(Post.IsValid). // KeepIf is a filter implementation
		Convert(Post.ToFriendlyPost)
	// Post.ToFriendlyPost takes the usual method receiver (a post in this case) as its first regular argument instead.
	// See https://go.dev/ref/spec#Method_expressions.

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
		TakeFirst(3).               // TakeFirst returns a slice of the first n elements
		ToStringsWith(Post.String). // ToStringsWith is the same as map but to the named builtin type, string in this case
		Each(hof.Println)           // Each applies the named function to each element for its side effects
	// for comparison to above:
	//
	//     for i, post := range posts { // again, which form of this?
	//         if i == 3 {              // three lines of code just to break, but simpler than C-style for loop
	//            break
	//         }
	//         fmt.Println(post.String())
	//     }

	// print the longest post title
	fmt.Println("\nthe longest post title in words:")

	titles := posts.ToStringsWith(Post.GetTitle)
	// for comparison to above:
	//
	//     titles := make([]string, len(posts))
	//     for i, post := range posts {
	//         titles[i] = post.Title
	//     }

	longestTitle := slices.MaxFunc(titles, CompareWordCounts) // fluent slices don't have Max or MaxFunc methods
	fmt.Println(longestTitle)

	// The MapWith method requires a type to map to, so SliceOf[T] doesn't have enough type parameters to support it.
	// For this reason, there's an additional type, MappableSliceOf.
	// Rather than go through creating a new type to demonstrate, we'll just specify string as the target type,
	// but it could be any type of your own.  Imagine your own type as the return type in this example.

	// first, type-convert our existing slice to a MappableSliceOf
	mappablePosts := fluent.MappableSliceOf[Post, string](posts)

	// now use Map to convert to strings
	fmt.Println("\nthe title lengths of the first three posts:")
	first3Titles := mappablePosts.
		TakeFirst(3).
		MapWith(Post.GetTitle) // imagine your return type here
	first3Titles.
		ToIntsWith(hof.StringLen).
		ToStringsWith(strconv.Itoa).
		Each(hof.Println)
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

// IsValid returns whether the post ID is positive.
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
