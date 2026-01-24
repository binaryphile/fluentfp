//go:build ignore

package main

import (
	"fmt"

	"github.com/binaryphile/fluentfp/lof"
	"github.com/binaryphile/fluentfp/slice"
)

// This example demonstrates slice.Mapper — a fluent slice type for filtering, mapping, and iteration.
func main() {
	// === Creating Fluent Slices ===

	posts := slice.From([]Post{
		{ID: 1, Title: "Introduction to Go"},
		{ID: 0, Title: ""},                        // invalid: ID is 0
		{ID: 2, Title: "Functional Programming"},
		{ID: 3, Title: "Error Handling Patterns"},
	})

	fmt.Println("all posts:", len(posts)) // all posts: 4

	// === Filtering ===

	validPosts := posts.KeepIf(Post.IsValid)
	fmt.Println("valid posts:", len(validPosts)) // valid posts: 3

	// isShortTitle reports whether the post title is 20 characters or fewer.
	isShortTitle := func(p Post) bool {
		return len(p.Title) <= 20
	}
	longTitles := posts.KeepIf(Post.IsValid).RemoveIf(isShortTitle)
	fmt.Println("posts with long titles:", len(longTitles)) // posts with long titles: 2

	// === Mapping ===

	normalized := validPosts.Convert(Post.Normalize)
	fmt.Println("first normalized:", normalized[0].Title) // first normalized: Introduction to Go

	titles := validPosts.ToString(Post.GetTitle)
	fmt.Println("last title:", titles[len(titles)-1]) // last title: Error Handling Patterns

	ids := validPosts.ToInt(Post.GetID)
	fmt.Println("last id:", ids[len(ids)-1]) // last id: 3

	// === Utilities ===

	first2 := validPosts.TakeFirst(2)
	fmt.Println("first 2 posts:", len(first2)) // first 2 posts: 2

	count := validPosts.Len()
	fmt.Println("count:", count) // count: 3

	fmt.Println("\nall valid posts:") // (3 posts follow)
	validPosts.ToString(Post.String).Each(lof.Println)

	// === Comparison to Loop ===

	// The fluent approach above:
	//   validPosts.ToString(Post.GetTitle)
	//
	// Equivalent loop:
	//   titles := make([]string, 0, len(validPosts))
	//   for _, p := range validPosts {
	//       titles = append(titles, p.GetTitle())
	//   }
	//
	// Loops are clearer when you need break/continue or channel consumption.
}

// Post represents a blog post.
type Post struct {
	ID    int
	Title string
}

// GetID returns the post's ID.
func (p Post) GetID() int { return p.ID }

// GetTitle returns the post's title.
func (p Post) GetTitle() string { return p.Title }

// IsValid reports whether the post has a positive ID.
func (p Post) IsValid() bool { return p.ID > 0 }

// Normalize returns the post with a default title if empty.
func (p Post) Normalize() Post {
	if p.Title == "" {
		p.Title = "Untitled"
	}
	return p
}

// String returns a display representation of the post.
func (p Post) String() string {
	return fmt.Sprintf("Post %d: %s", p.ID, p.Title)
}
