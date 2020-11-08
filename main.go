package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	ParseArgsAndDispatch()
}

func ParseArgsAndDispatch() {
	fetchMapCmd := flag.NewFlagSet("fetch", flag.ExitOnError)

	// masterProfile is the profile used to generate STS tokens for
	// assuming other roles. This profile will also be used to list all the
	// accounts in an organizational unit.
	var masterProfile = fetchMapCmd.String(
		"masterProfile",
		"default",
		"Profile from which STS tokens for assuming roles can be generated.",
	)

	// accountRole is the role name to assume in
	// each account such that alias information can be gathered.
	var accountRole = fetchMapCmd.String(
		"accountRole",
		"",
		"Name of role to assume in each account gather alias information.",
	)

	if len(os.Args) < 2 {
		fmt.Println("Must specify a subcommand to execute. fetch or upsert")
		os.Exit(1)
	}

	switch os.Args[1] {

	case "fetch":
		fetchMapCmd.Parse(os.Args[2:])
		if *accountRole == "" {
			fmt.Println("Must specify a role to assume in each account")
			os.Exit(1)
		}
		FetchAliasToAccountMap(*masterProfile, *accountRole)

	default:
		fmt.Println("Expected subcommands")
		os.Exit(1)
	}
}
