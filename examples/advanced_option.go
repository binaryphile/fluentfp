//go:build ignore

package main

import (
	"fmt"
	"os"

	"github.com/binaryphile/fluentfp/lof"
	"github.com/binaryphile/fluentfp/option"
	"github.com/google/go-cmp/cmp"
)

// === ADVANCED OPTIONS ===
//
// Advanced options embed a basic option and add methods that conditionally
// call the wrapped value's methods based on ok status. ("Advanced" complements
// "basic" - the standard option.Basic[T] type.)
//
// This example demonstrates the pattern with lifecycle management (Close),
// but it applies to any type where you want conditional method calls.
//
// === PATTERN ===
//
// Problem: An App struct holds optional dependencies. Some commands need one
// dependency, others need both. How do you Close() without conditionals?
//
// Solution: Wrap each dependency in an advanced option that has a Close method.
// The option's Close calls the underlying Close only if ok.
//
// Result: App.Close() just calls each option's Close - no if-statements needed.
// OpenApp() just calls factory functions - no if-statements needed.
// The conditionality lives in the option type, not scattered through your code.

// === USAGE EXAMPLE ===

// main allows the user to specify "source", "dest" or "compare" to inspect the two databases.
func main() {
	// the data in the two databases
	sourceData := `[{
	"email": "user1@example.com"
}]`

	destData := `[{
	"email": "user2@example.com"
}]`

	// get the user's command
	command := os.Args[1]

	switch command {
	case "source":
		app := OpenApp(OpenAppArgs{SourceData: sourceData})
		defer app.Close() // closes only opened deps

		client := app.SourceClientOption.MustGet()
		users := client.ListUsers()
		fmt.Print("Source users:\n", users)

	case "dest":
		app := OpenApp(OpenAppArgs{DestData: destData})
		defer app.Close()

		client := app.DestClientOption.MustGet()
		users := client.ListUsers()
		fmt.Print("Dest users:\n", users)

	case "compare":
		app := OpenApp(OpenAppArgs{SourceData: sourceData, DestData: destData})
		defer app.Close()

		sourceClient := app.SourceClientOption.MustGet()
		destClient := app.DestClientOption.MustGet()
		sourceUsers := sourceClient.ListUsers()
		destUsers := destClient.ListUsers()

		result := cmp.Diff(sourceUsers, destUsers)
		if diff, hasDiff := lof.IfNotEmpty(result); hasDiff {
			fmt.Print("data sources are NOT in sync:\n", diff, "\n")
		} else {
			fmt.Println("data sources are in sync")
		}
	}
}

// === ADVANCED OPTION TYPE ===

// ClientOption embeds a basic option and adds a conditional Close method.
type ClientOption struct {
	option.Basic[Client]
}

// NewClientOption wraps a basic option in a ClientOption.
func NewClientOption(basic option.Basic[Client]) ClientOption {
	return ClientOption{
		Basic: basic,
	}
}

// OpenClientAsOption returns ok ClientOption if users is non-empty.
func OpenClientAsOption(users string) ClientOption {
	// This chains options so OpenApp needs no conditionals:
	usersOption := option.IfNotZero(users)                   // "" → not-ok
	clientBasicOption := option.Map(usersOption, OpenClient) // not-ok passes through
	return NewClientOption(clientBasicOption)
}

// Close closes the Client if ok.
func (o ClientOption) Close() {
	o.Call(Client.Close)
}

// === DOMAIN TYPES ===

// Client is a "client" to our fake database.
// It is the value type contained by its complementary advanced option.
// It returns the configured users string as if it were the result of a query.
type Client struct {
	users string
}

// OpenClient returns a Client with the users field populated.
// It intentionally follows the open/close protocol of stateful dependencies,
// even though our fake database doesn't have a stateful connection.
func OpenClient(users string) Client {
	return Client{
		users: users,
	}
}

// Close closes...nothing.
// It's here for the open/close protocol to be followed.
func (c Client) Close() {}

// ListUsers returns the users string as if it were the result of a query.
func (c Client) ListUsers() string {
	return c.users
}

// === APP STRUCT & FACTORY ===

// App is a collection of external dependencies.
// Because not all may be opened at once, they are stored as options.
type App struct {
	SourceClientOption ClientOption
	DestClientOption   ClientOption
}

// OpenApp opens dependencies for non-zero arguments.
func OpenApp(a OpenAppArgs) App {
	return App{
		SourceClientOption: OpenClientAsOption(a.SourceData),
		DestClientOption:   OpenClientAsOption(a.DestData),
	}
}

// OpenAppArgs holds configuration for each dependency. Zero fields are not opened.
type OpenAppArgs struct {
	SourceData string
	DestData   string
}

// Close closes opened dependencies. No conditionals needed - options handle it.
func (a App) Close() {
	a.SourceClientOption.Close()
	a.DestClientOption.Close()
}

// === WHEN TO USE ===
//
// Use advanced options when:
// - Many dependencies with lifecycle methods (Open/Close)
// - Factory functions that conditionally open resources
// - You want to eliminate conditional logic in Close() methods
//
// Skip when: single dependency, or types without methods to call conditionally.
//
// Each dependency needs: a client type, an advanced option wrapping it,
// a factory returning the advanced option, a field in App, and a Close call.
// None of these require conditionals - that's the pattern's value.
