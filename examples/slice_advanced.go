//go:build ignore

package main

import (
	"fmt"

	"github.com/binaryphile/fluentfp/slice"
	"github.com/binaryphile/fluentfp/tuple/pair"
)

// This example demonstrates advanced slice operations: MapTo, Fold, Unzip, and Zip.
func main() {
	// Sample data — inline for self-contained example
	posts := slice.From([]Post{
		{ID: 1, Title: "Introduction to Go"},
		{ID: 2, Title: "Functional Programming"},
		{ID: 3, Title: "Error Handling"},
	})

	// === Mapping to Different Types ===

	// MapTo[R] creates a MapperTo that can map to arbitrary type R
	// titleFromPost extracts a Title from a Post.
	titleFromPost := func(p Post) Title { return Title(p.Title) }

	titles := slice.MapTo[Title](posts).To(titleFromPost)
	fmt.Println("titles as Title type:", titles)

	// Chain with other operations
	lengths := slice.MapTo[Title](posts).
		To(titleFromPost).
		ToInt(Title.Len)
	fmt.Println("title lengths:", lengths)

	// === Reducing ===

	// Fold reduces a slice to a single value.
	// This pattern is essential for event sourcing state reconstruction.

	// sumIDs accumulates post IDs into a running total.
	sumIDs := func(total int, p Post) int { return total + p.ID }
	totalID := slice.Fold(posts, 0, sumIDs)
	fmt.Println("sum of IDs:", totalID)

	// indexByID builds a map keyed by post ID.
	indexByID := func(m map[int]Post, p Post) map[int]Post {
		m[p.ID] = p
		return m
	}
	byID := slice.Fold(posts, make(map[int]Post), indexByID)
	fmt.Println("post with ID 2:", byID[2].Title)

	// === Multi-field Extraction ===

	// Unzip extracts multiple fields in a single pass (avoids N iterations)
	ids, postTitles := slice.Unzip2(posts, Post.GetID, Post.GetTitle)
	fmt.Println("extracted IDs:", ids)
	fmt.Println("extracted titles:", postTitles)

	// Unzip3 and Unzip4 extract more fields in one pass
	// (shown with same fields for brevity — real use would have distinct fields)
	ids3, titles3, _ := slice.Unzip3(posts,
		Post.GetID,
		Post.GetTitle,
		Post.GetIDAsFloat64,
	)
	fmt.Printf("unzip3: %d IDs, %d titles\n", len(ids3), len(titles3))

	// === Zipping ===

	// pair.Zip combines two slices into pairs
	// pair.ZipWith combines and transforms in one step
	// See tuple/pair package for details.

	ratings := []int{5, 4, 3}
	pairs := pair.Zip([]Post(posts), ratings) // explicit conversion from Mapper to []Post
	fmt.Printf("zipped %d post-rating pairs\n", len(pairs))
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

// GetIDAsFloat64 returns the post's ID as a float64.
func (p Post) GetIDAsFloat64() float64 { return float64(p.ID) }

// Title is a distinct type for post titles.
type Title string

// Len returns the length of the title.
func (t Title) Len() int { return len(t) }
