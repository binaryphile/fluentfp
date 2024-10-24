package main

import (
	"fmt"
	"github.com/binaryphile/fluentfp/option"
	"github.com/google/go-cmp/cmp"
	"os"
)

// Advanced option usage as described here requires more code and has more limited use cases.
// However, it can also drastically lower the complexity of code having
// many optional types when those types are needed for their behavior (methods),
// as opposed to their field values.
//
// As an example, let's design a simple tool intended to "synchronize" two databases.
// Well, we haven't gotten as far as actually synchronizing them yet,
// but we do have commands to show users in the source and destination databases,
// as well as to see if the databases differ: subcommands "source", "dest" and "compare".
// Each subcommand of the tool needs to connect to the appropriate database(s);
// the first two commands need the requisite source or destination database,
// while the third requires both.
//
// We won't actually connect to real databases, but we'll simulate the process with a Client struct
// that has a Close method and whose factory is OpenClient.
// Each "database" is seeded with a JSON string representing a users table, but it's just a string.
//
// The databases are external dependencies for the tool that live for the duration of the run.
func main() {
	sourceData := `[{
	"id": 1,
	"email": "user1@example.com"
}]`

	destData := `[{
	"id": 2,
	"email": "user2@example.com"
}]`

	command := os.Args[1]

	switch command {
	case "source":
		app := OpenApp(OpenAppArgs{
			SourceData: sourceData,
		})
		defer app.Close()

		client := app.SourceClientOption.MustGet()

		users := client.ListUsers()
		fmt.Print("Source users:\n", users)

	case "dest":
		app := OpenApp(OpenAppArgs{
			DestData: destData,
		})
		defer app.Close()

		client := app.DestClientOption.MustGet()

		users := client.ListUsers()
		fmt.Print("Dest users:\n", users)

	case "compare":
		app := OpenApp(OpenAppArgs{
			SourceData: sourceData,
			DestData:   destData,
		})
		defer app.Close()

		sourceClient := app.SourceClientOption.MustGet()
		destClient := app.DestClientOption.MustGet()

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

type ClientOption struct {
	option.Basic[Client]
}

func NewClientOption(basic option.Basic[Client]) ClientOption {
	return ClientOption{
		Basic: basic,
	}
}

func OpenClientAsOption(users string) ClientOption {
	usersOption := option.IfProvided(users)
	clientBasicOption := option.Map(usersOption, OpenClient)
	return NewClientOption(clientBasicOption)
}

func (o ClientOption) Close() {
	o.Call(Client.Close)
}

type Client struct {
	users string
}

func OpenClient(users string) Client {
	return Client{
		users: users,
	}
}

func (c Client) Close() {}

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

func OpenApp(c OpenAppArgs) App {
	return App{
		SourceClientOption: OpenClientAsOption(c.SourceData),
		DestClientOption:   OpenClientAsOption(c.DestData),
	}
}

type OpenAppArgs struct {
	SourceData string
	DestData   string
}

func (a App) Close() {
	a.SourceClientOption.Close()
	a.DestClientOption.Close()
}
