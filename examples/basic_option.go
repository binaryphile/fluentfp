package main

import (
	"encoding/json"
	"fmt"
	"github.com/binaryphile/fluentfp/option"
	"net/http"
	"strconv"
)

// This is a usage example for option.Basic.
// Options are containers that either hold a value (ok) or don't (not-ok).
// option.Basic offers a variety of methods familiar from functional programming, such as map (MapWith).
// While option.Basic is useful on its own, as the name implies,
// there is a more advanced use case.
// See advanced_option.go as well.

// main fetches posts from the post api as options.
func main() {
	// option.Basic is generic, so when you have to declare a variable,
	// include the type of value it holds.
	// We alias the concrete type here to make things a little more readable when it's used.
	type PostOption = option.Basic[Post]

	// getPosts returns the posts with the provided ids.
	getPosts := func(ids ...int) []Post {
		posts := make([]Post, len(ids))
		for i, id := range ids {
			resp, _ := http.Get("https://jsonplaceholder.typicode.com/posts/" + strconv.Itoa(id))

			// Decode the response into a post.
			// This may error on individual requests, in which case,
			// we want the zero value for the post to know that it's invalid.
			err := json.NewDecoder(resp.Body).Decode(&posts[i])
			if err != nil {
				posts[i] = Post{} // Decode doesn't guarantee &posts[i] is unchanged
			}
		}

		return posts
	}

	// get two posts
	posts := getPosts(1, 2)

	// Sometimes you have data from an external source that you need to marry with other data.
	// Options can provide a way to combine corresponding data where individual records may have errored,
	// but you still want to work with the data that is available.

	// We'll play a trick here and make a slice of PostOptions with three slots instead of two,
	// so the last slot is always not-ok.
	// If any requests errored, those will also be not-ok since zero-value posts don't pass IsValid.

	// create options from posts using post.IsValid() as the ok value
	postOptions := make([]PostOption, 3)
	for i, post := range posts {
		postOptions[i] = option.New(post, post.IsValid())
	}

	// We have additional data about the posts in the form of personal notes on them.
	// The next couple blocks print the posts along with the associated note.

	myNotesOnPosts := []string{
		"a compelling tale of woe",
		"a reggae music review",
		"commercials from the 70s and 80s",
	}

	// since the last post is not-ok, it won't be printed
	for i, postOption := range postOptions {
		if post, ok := postOption.Get(); ok { // Get gets the option's value using Go's comma-ok idiom
			fmt.Print("\n", post.String(), "\n")
			fmt.Println("My notes:", myNotesOnPosts[i])
		}
	}

	// the next examples show how to create options

	fortyTwoOption := option.Of(42)        // Of creates an ok option from a value
	notOkIntOption := option.IfProvided(0) // IfProvided creates an ok option if the value is not the zero value for the type

	// there are pre-declared option types for the built-in types
	notOkStringOption := option.String{}   // the zero value is not-ok because the ok field's zero value is false
	notOkStringOption = option.NotOkString // but there are more readable package variables to create not-oks

	// Sometimes in Go you encounter a pointer being used as a pseudo-option where nil means not-ok.
	// The OfPointee method creates a formal option of the pointed-to value.
	postsOption := option.OfPointee(&posts) // this gives the same result as option.Of(posts)
	pseudoOption := postsOption.ToPointer() // the ToPointer method gets the pointer pseudo-option back

	// New dynamically creates an option when you have a value and an ok bool
	theAnswer := 42 // to the question of Life, the Universe and Everything
	ok := true
	okIntOption := option.New(theAnswer, ok)

	fortyTwo := okIntOption.MustGet()   // MustGet gets the value from an option you know is ok or else panics
	sixtyEight := notOkIntOption.Or(68) // Or gets the value from an ok option or else the supplied alternative
	zero := notOkIntOption.OrZero()     // OrZero gets the zero value of the type as the alternative

	// OrZero is generic and works for strings and bools, but there are more readable versions for them
	empty := option.NotOkString.OrEmpty()
	False := option.NotOkBool.OrFalse()

	// if the alternative value for Or requires computation, there is OrCall
	zero = notOkIntOption.OrCall(func() int { return 0 })

	// IsOk checks if an option is ok
	if okIntOption.IsOk() {
		fmt.Println("Int option is ok")
	}

	// Convert returns an option of either not-ok if the original is also not-ok,
	// or ok with the result of applying the function to the value.
	// The value type must be the same as the original.
	okDoubledIntOption := okIntOption.Convert(func(i int) int { return 2 * i })

	// there are To[Type]With methods for the built-in types
	okStringOption := okIntOption.ToStringWith(strconv.Itoa)

	// But no method for map to a named type.
	// Use option.Map instead, it is the generic map function.
	IntToPost := func(i int) Post {
		return Post{
			ID:    i,
			Title: "Post #" + strconv.Itoa(i),
		}
	}
	okPostOption := option.Map(okIntOption, IntToPost)

	intIs42 := func(i int) bool {
		return i == 42
	}
	// filter is implemented with two complementary methods
	stillOkIntOption := okIntOption.KeepOkIf(intIs42)
	nowNotOkIntOption := okIntOption.ToNotOkIf(intIs42)

	// Congratulations, you're officially ready to use options!
	// See advanced_option.go for examples of options with behavior from their value types.

	// ignore the following -- to keep Go happy
	eat[*[]Post](pseudoOption)
	eat[PostOption](okPostOption)
	eat[bool](False)
	eat[int](fortyTwo, sixtyEight, zero)
	eat[option.Int](fortyTwoOption, notOkIntOption, stillOkIntOption, nowNotOkIntOption, okDoubledIntOption)
	eat[option.String](notOkStringOption, okStringOption)
	eat[string](empty)
}

// Post type definition
///////////////////////

// Post represents a post from the JSONPlaceholder API.
type Post struct {
	ID    int
	Title string
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

func eat[T any](_ ...T) {
	return
}
