package main

import (
	"encoding/json"
	"fmt"
	"github.com/binaryphile/fluentfp/fluent"
	"github.com/binaryphile/fluentfp/option"
	"net/http"
	"strconv"
)

// main fetches posts from the post api as options.
func main() {
	// option.Basic offers a variety of option-related methods.
	// It is generic, so you have to instantiate it with the type of value it holds.
	// We alias option.Basic[Post] here to make things a little more readable when it's used.
	type PostOption = option.Basic[Post]

	// getPosts returns the posts with the provided ids.
	getPosts := func(ids ...int) []Post {
		posts := make([]Post, len(ids))
		for i, id := range ids {
			resp, _ := http.Get("https://jsonplaceholder.typicode.com/posts/" + strconv.Itoa(id))

			// decode the response into a post
			json.NewDecoder(resp.Body).Decode(&posts[i])
		}

		return posts
	}

	posts := getPosts(1, 2, 3, 4)

	postOptions := make([]PostOption, len(posts))
	for i, post := range posts {
		postOptions[i] = post.ToOption()
	}

	return postOptions

	// print how many are not-ok, i.e. not valid
	lenNotOkPosts := 0
	for _, postOption := range postOptions {
		if !postOption.IsOk() {
			lenNotOkPosts += 1
		}
	}
	if lenNotOkPosts > 0 {
		fmt.Println(lenNotOkPosts, "posts were invalid.")
	}

	// print my notes on the posts that are ok
	// each note below corresponds to the post id by its index
	myNotesOnPosts := []string{
		"a compelling tale of woe",
		"some recipes from a stuntman",
		"a reggae music review",
		"commercials from the 70s and 80s",
	}
	for i, postOption := range postOptions {
		if post, ok := postOption.Get(); ok {
			fmt.Println(post)
			fmt.Println("my notes:", myNotesOnPosts[i])
		}
	}

	// for comparison, the same code with fluent slices
	var posts fluent.MappableSliceOf[Post, PostOption]

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

// ToOption returns an ok option for valid posts, otherwise a not-ok option.
func (p Post) ToOption() option.Basic[Post] {
	return option.NewBasic(p, p.IsValid())
}
