//go:build ignore

package main

import (
	"fmt"
	"github.com/binaryphile/fluentfp/option"
	"github.com/google/go-cmp/cmp"
	"os"
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
	// commands source and dest do the same thing each with a different database: print the users
	case "source":
		// The idea with OpenApp is to allow an argument for each dependency.
		// If the function receives a non-zero version of that argument,
		// it opens the dependency.
		// Otherwise, the resulting option for the dependency is not-ok.
		// This works most easily with a single struct argument, OpenAppArgs.
		// Any unsupplied field gets the zero-value.
		// OpenApp is implemented to not open dependencies with zero-value arguments.
		app := OpenApp(OpenAppArgs{
			// in this example, we've resorted to using the user data itself
			// as the configuration argument for the dependency,
			// but the argument would normally be something like a struct
			// with connection details (server name, DSN, etc.).
			SourceData: sourceData,
		})
		defer app.Close() // app.Close knows to only try to close opened dependencies

		// unwrap the client from the option, so we can use it
		client := app.SourceClientOption.MustGet() // we have to have this dependency, so panic if not there

		// now do something useful with it, list users in this case
		users := client.ListUsers()
		fmt.Print("Source users:\n", users)

	// dest is the same as above
	case "dest":
		app := OpenApp(OpenAppArgs{
			DestData: destData, // this is the only difference
		})
		defer app.Close()

		client := app.DestClientOption.MustGet()

		users := client.ListUsers()
		fmt.Print("Dest users:\n", users)

	// compare gives a diff of the users data by google's diff library
	case "compare":
		// this time we need both dependencies, so they both get arguments
		app := OpenApp(OpenAppArgs{
			SourceData: sourceData,
			DestData:   destData,
		})
		defer app.Close()

		// unwrap dependencies
		sourceClient := app.SourceClientOption.MustGet()
		destClient := app.DestClientOption.MustGet()

		// use them
		sourceUsers := sourceClient.ListUsers()
		destUsers := destClient.ListUsers()
		diff := cmp.Diff(sourceUsers, destUsers)
		switch diff {
		case "":
			fmt.Println("data sources are in sync")
		default:
			fmt.Print("data source are NOT in sync:\n", diff, "\n")
		}
	}
}

// === ADVANCED OPTION TYPE ===

// ClientOption embeds a basic option and adds a conditional Close method.
type ClientOption struct {
	option.Basic[Client] // we just need the basic option and to add a Close method, shown after the factories below.
}

// NewClientOption is a factory that just accepts the embedded option field as an argument.
// Having a simple factory like this that only accepts fields helps keep the consumer code readable.
// It can also be used as an argument to higher-order functions,
// whereas the literal form used to provide its return value cannot.
func NewClientOption(basic option.Basic[Client]) ClientOption {
	return ClientOption{
		Basic: basic,
	}
}

// OpenClientAsOption returns an ok ClientOption if users is provided (not empty).
// This function handles everything so that you don't need to write code around it in OpenApp,
// allowing the resulting ClientOption to be assigned straight to its field in the App struct.
// This keeps the OpenApp code economical which is our goal,
// because OpenApp is where complexity mushrooms as dependencies are added.
func OpenClientAsOption(users string) ClientOption {
	usersOption := option.IfProvided(users)                  // ok if not empty
	clientBasicOption := option.Map(usersOption, OpenClient) // option.Map accepts and returns basic options
	return NewClientOption(clientBasicOption)                // convert the basic option to the advanced option
}

// Close conditionally closes the Client.
// In our example, this method is the whole point of the advanced option.
// It means the consumer doesn't need conditional statements,
// it can simply call Close and get the proper behavior
// based on whether the dependency was opened or not.
// Having a method to do this rather than the consumer having to use option.Call directly
// means the resulting code conforms to existing Go developer expectations
// of how closing the dependency *should* look when they read the code.
func (o ClientOption) Close() {
	// option.Call applies the function argument to the contained value for its side effect
	// if the option is ok
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

// OpenApp conditionally opens each dependency if it was requested with a non-zero argument.
// It relies on OpenClientAsOption to fill out the field directly, using the provided argument.
// This is the code that benefits the most from advanced options,
// because this is where code otherwise mushrooms as dependencies are added.
func OpenApp(a OpenAppArgs) App {
	return App{
		SourceClientOption: OpenClientAsOption(a.SourceData),
		DestClientOption:   OpenClientAsOption(a.DestData),
	}
}

// OpenAppArgs offers a field for each dependency to obtain its configuration parameters as a signal to open it.
// Zero-values in fields indicate that a dependency is not requested,
// so the consumer doesn't have to provide undesired fields explicitly when calling OpenApp.
type OpenAppArgs struct {
	SourceData string
	DestData   string
}

// Close closes the dependencies.
// This is also where we benefit from advanced options,
// since we only have to call Close on them without making any decisions.
// The options will do the right thing without needing conditionals here.
// This is another place where complexity otherwise mushrooms as dependencies are added.
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
