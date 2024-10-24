package main

import (
	"fmt"
	"github.com/binaryphile/fluentfp/option"
	"github.com/google/go-cmp/cmp"
	"os"
)

// Advanced options expose methods corresponding to the value type's methods that do the right thing based on ok status.
// Advanced option usage as described here requires more code and has more limited use cases than basic options.
// The benefit, however, is that it can also drastically lower the complexity of code that manipulates
// many option instances in one place (i.e. one function), when the stored value types are needed for their methods.
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
// The databases are external dependencies that live for the duration of the tool's run.
// One way to approach long-lived dependencies is to collect them in a single place for lifecycle management.
// The App struct is this collection of dependencies needed by the application before it can run.
// Each dependency has a field in the struct.
// When the tool is run, we determine the user's command and instantiate an App by using the factory OpenApp.
// The arguments to OpenApp determine which of the dependencies are actually opened.
// For example, since the "source" command doesn't talk to the dest database, only the source database is opened.
// To handle the existence of some of the fields in App but not others, each field is an option type whose value
// is the now-open dependency, or not-ok if a dependency wasn't requested.
// The command code then obtains the dependency from the corresponding option in the App instance
// by calling app.[Dependency]Option.MustGet.
//
// In this example, we only have two dependencies,
// which makes this approach look overly verbose for the benefit it provides.
// Real applications may have many dependencies, however,
// which can quickly blow up the cyclomatic complexity of a factory like OpenApp.
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
	// commands source and dest do the same thing with a different database: print the users
	case "source":
		// The idea with OpenApp is to provide an argument for each dependency.
		// If the dependency gets a non-zero version of that argument,
		// it opens the dependency.
		// Otherwise, the resulting option for the dependency is not-ok.
		// This works most easily with a single struct argument, OpenAppArgs.
		// Any unsupplied field gets the zero-value.
		// OpenApp is implemented to not open dependencies with zero-value arguments.
		app := OpenApp(OpenAppArgs{
			// in this example, we've resorted to using the data itself as the configuration for the dependency,
			// but this would usually be something like a struct with connection details (server name, DSN, etc.)
			SourceData: sourceData,
		})
		defer app.Close() // app knows how to close all enabled dependencies

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

	// compare gives a diff of the reported users using google's diff library
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
// The return value (if there is one) then must also come back as an option of the return type,
// so each method's signature is usually different from the underlying method's.
// The way to implement this is with a named type that embeds a basic option
// and has methods for each of the value type's methods.
// In this example, we only need the Close method implementation.
type ClientOption struct {
	option.Basic[Client] // we just need the basic option and to add some methods, shown after the factories below.
}

// NewClientOption is a factory that simply accepts the embedded option field as an argument.
// Having a simple factory like this that just accepts fields helps keep the consumer code sensible.
// It can be used by higher-order functions, whereas the literal form cannot, by itself.
func NewClientOption(basic option.Basic[Client]) ClientOption {
	return ClientOption{
		Basic: basic,
	}
}

// OpenClientAsOption returns an ok ClientOption if users is provided (not empty).
// This function handles everything so that you don't need to write code around it in OpenApp,
// because the resulting ClientOption just goes straight into a field in the App struct.
// This keeps the OpenApp code exceedingly concise, which is our goal,
// because that's where complexity mushrooms.
func OpenClientAsOption(users string) ClientOption {
	usersOption := option.IfProvided(users)                  // ok if not empty
	clientBasicOption := option.Map(usersOption, OpenClient) // option.Map is a function that works on basic options (no fluent map)
	return NewClientOption(clientBasicOption)                // convert the basic option to the advanced option
}

// Close conditionally closes the Client.
// In our example, this method is the whole point of the advanced option.
// It means the consumer doesn't need conditional statements,
// it can simply call Close and get the proper behavior
// based on whether the dependency was opened or not.
// Having a method to do this rather than the consumer having to use option.Call directly
// means the resulting code conforms to existing Go developer expectations
// of how closing the dependency should look when they read the code.
func (o ClientOption) Close() {
	o.Call(Client.Close) // Call applies the function to the contained value for its side effect if the option is ok
}

// Client is a "client" to our fake database.
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
// Because not all may be instantiated at once, they are stored as options.
type App struct {
	SourceClientOption ClientOption
	DestClientOption   ClientOption
}

// OpenApp conditionally opens each dependency if it was requested with non-zero argument.
// It relies on the "OpenAsOption" factory for the clients to test the argument and do the right thing,
// resulting in very concise code here, which is the lion's share of the benefit of options,
// because this is where code mushrooms as dependencies get added.
func OpenApp(c OpenAppArgs) App {
	return App{
		SourceClientOption: OpenClientAsOption(c.SourceData),
		DestClientOption:   OpenClientAsOption(c.DestData),
	}
}

// OpenAppArgs offers a field to configure each dependency.
// Zero-values in fields indicate that a dependency is not requested,
// so the consumer doesn't have to provide all fields explicitly when calling OpenApp.
type OpenAppArgs struct {
	SourceData string
	DestData   string
}

// Close closes the dependencies.
// This is also where we benefit from the advanced options,
// since we only have to call Close on them.
// They will do the right thing without needing conditionals here.
// This is another place where complexity normally mushrooms as dependencies are added.
func (a App) Close() {
	a.SourceClientOption.Close()
	a.DestClientOption.Close()
}
