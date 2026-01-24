//go:build ignore

package main

import (
	"fmt"

	"github.com/binaryphile/fluentfp/slice"
	"github.com/binaryphile/fluentfp/tuple/pair"
)

// This example demonstrates advanced slice operations: MapTo, Fold, Unzip, and Zip.
func main() {
	posts := slice.From([]Post{
		{ID: 1, Title: "Introduction to Go"},
		{ID: 2, Title: "Functional Programming"},
		{ID: 3, Title: "Error Handling"},
	})

	// === Mapping to Different Types ===

	// titleFromPost extracts a Title from a Post.
	titleFromPost := func(p Post) Title { return Title(p.Title) }

	titles := slice.MapTo[Title](posts).To(titleFromPost)
	fmt.Println("last title:", titles[len(titles)-1]) // last title: Error Handling

	lengths := slice.MapTo[Title](posts).
		To(titleFromPost).
		ToInt(Title.Len)
	fmt.Println("last length:", lengths[len(lengths)-1]) // last length: 14

	// === Reducing ===

	// sumIDs accumulates post IDs into a running total.
	sumIDs := func(total int, p Post) int { return total + p.ID }
	totalID := slice.Fold(posts, 0, sumIDs)
	fmt.Println("sum of IDs:", totalID) // sum of IDs: 6

	// indexByID builds a map keyed by post ID.
	indexByID := func(m map[int]Post, p Post) map[int]Post {
		m[p.ID] = p
		return m
	}
	byID := slice.Fold(posts, make(map[int]Post), indexByID)
	fmt.Println("post with ID 2:", byID[2].Title) // post with ID 2: Functional Programming

	// === Multi-field Extraction ===

	ids, postTitles := slice.Unzip2(posts, Post.GetID, Post.GetTitle)
	fmt.Printf("unzip2: %d IDs, %d titles\n", len(ids), len(postTitles)) // unzip2: 3 IDs, 3 titles

	ids3, titles3, _ := slice.Unzip3(posts,
		Post.GetID,
		Post.GetTitle,
		Post.GetIDAsFloat64,
	)
	fmt.Printf("unzip3: %d IDs, %d titles\n", len(ids3), len(titles3)) // unzip3: 3 IDs, 3 titles

	// === Zipping ===

	ratings := []int{5, 4, 3}
	pairs := pair.Zip([]Post(posts), ratings)
	fmt.Printf("zipped %d post-rating pairs\n", len(pairs)) // zipped 3 post-rating pairs
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
