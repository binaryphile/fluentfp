package main

import (
	"fmt"
	"github.com/binaryphile/fluentfp/option"
	"github.com/google/go-cmp/cmp"
	"os"
)

// Advanced options expose methods corresponding to the value type's methods that do the right thing based on ok status.
// Advanced option usage as described here requires more code and has more limited use cases than basic options.
// The benefit, however, is that these options can significantly lower the complexity of code that manipulates
// many option instances in one place (i.e. in one central function),
// at least when the stored value types are needed for their methods.
// That's because the difference between a basic and advanced option
// is the presence of option-aware versions of those methods.
//
// As an example, let's design a simple cli tool intended to synchronize the users in two databases.
// Well, we haven't gotten as far as actually synchronizing them yet,
// but we do have subcommands to show the users from each of the source and destination databases,
// as well as to see if the databases differ: commands "source", "dest" and "compare".
// Each command needs to connect to the appropriate database(s);
// the first two need the requisite source or destination database,
// while the third requires both.
//
// We won't actually connect to real databases, but we'll simulate the process with a Client struct
// that looks like a database connection in that it has a deferrable Close method
// and a factory function named OpenClient.
// Each "database" has a single user in a JSON string that represents a users table,
// but it's just a hardwired string in the Client struct.
//
// The databases are external dependencies to this tool that live for the duration of the tool's run.
// One way to approach long-lived dependencies is to collect them in a single place for lifecycle management.
// The App struct is this collection of dependencies needed by the application before it can run.
// Each dependency has a field in the struct.
// When the tool is run, we determine the user's command and instantiate an App by using the factory function OpenApp.
// The arguments to OpenApp determine which of the dependencies are actually opened.
// For example, since the "source" command doesn't talk to the dest database, only the source database is opened.
// To handle the existence of some of the fields in App but not others, each field is an option type whose value
// is the now-open dependency, or not-ok if a dependency wasn't requested.
// The command code then obtains the dependency from the corresponding option in the App instance
// by calling app.[Dependency]Option.MustGet.
//
// In this example, we only have two dependencies, which makes this approach appear verbose.
// Real applications may have many dependencies, however,
// which can quickly blow up the complexity of an omnibus-style factory like OpenApp.
// To illustrate the usefulness of this advanced option approach,
// we'll create an OpenApp that now can be very simple because
// we've offloaded the conditionality (i.e. the if-statements) to these advanced options.

// main allows the user to specify "source", "dest" or "compare" to inspect the two databases.
func main() {
	// the data in the two databases
	sourceData := `[{
	"id": 1,
	"email": "user1@example.com"
}]`

	destData := `[{
	"id": 2,
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

// Below is where it gets interesting.  We'll see the definition of the advanced option ClientOption
// and the implementation of OpenApp, which is the primary beneficiary of ClientOption's ergonomics.

// ClientOption is an advanced option faking our database dependencies.
// An advanced option is a named type that offers conditional access to the methods of the value type.
// It embeds a basic option so that the basic option methods are available as well.
// The idea is to offer the same behavior as the value type,
// but to only call the contained value's method if the option is ok.
// The return value (if there is one) then must come back as an option of the original method's return type.
// The way to implement this is with a struct that embeds a basic option
// and has methods for the value type's methods of interest.
// In this example, we only need the Close method.
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

// dependency instantiation

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

// Ultimately, this pattern can be followed with little variation for each type of dependency,
// and the resulting code pushes the complexity into a consistently-structured set of methods and functions
// which are themselves kept simple by being couched as functional operations.
// It's how they are assembled that makes them powerful.
//
// Notice that each new type of dependency usually only needs one argument in OpenAppArgs and one line of code in
// the OpenApp function, so long as you also provide:
// - the client implementation of the dependency
// - an advanced option wrapping a basic option of the client
// - a factory for the advanced option that accepts just the basic option
// - a factory that acts on the OpenAppArgs argument and returns an advanced option
// - a field in the App struct that gets filled in with a single call to the second factory (usually,
//  	although dependency trees can require more work)
// - a call to the dependency's Close method in app's Close method
//
// Also notice how none of these structs/functions are more than three (readable) lines as implemented in this example,
// and don't contain any branches.
// That's about as straightforward as it can get.
//
// While the benefits of advanced options *are* valuable in this and other use cases,
// their application is more limited than that of basic options and requires more explanation,
// so use them with temperance if you work on teams.
