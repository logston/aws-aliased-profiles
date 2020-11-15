package common

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

const (
	AccountsAliasedJsonFilename   = "aliased-accounts.json"
	AccountsAliasedConfigFilename = "aliased-accounts.tmpl"
	AWSConfigFilename             = "config"
	DefaultProfileTemplate        = `
[profile {{ .Alias }}]
role_arn = arn:aws:iam::{{ .Id }}:role/MyFavRoleToAssume
source_profile = default
`
	AWSConfigDelimiter = "### ----- AWS Aliased Profiles -----"
)

type Tag struct {
	Key   string
	Value string
}

type Account struct {
	// The unique identifier (ID) of the account.
	//
	// The regex pattern (http://wikipedia.org/wiki/regex) for an account ID string
	// requires exactly 12 digits.
	Id string

	// The date the account became a part of the organization.
	JoinedTimestamp time.Time

	// The status of the account in the organization.
	Status string

	// Alias associated with the account.
	Alias string

	Tags []*Tag
}

func ExitWithError(err error) {
	fmt.Fprintf(os.Stderr, "error: %v\n", err)
	os.Exit(1)
}

func WriteAccountList(al []*Account) {
	data, err := json.MarshalIndent(al, "", "    ")
	if err != nil {
		ExitWithError(err)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		ExitWithError(err)
	}

	path := strings.Join([]string{home, ".aws", AccountsAliasedJsonFilename}, string(os.PathSeparator))

	if err := ioutil.WriteFile(path, data, 0644); err != nil {
		ExitWithError(err)
	}
}

func ReadAccountList() (al []*Account) {
	home, err := os.UserHomeDir()
	if err != nil {
		ExitWithError(err)
	}

	path := strings.Join([]string{home, ".aws", AccountsAliasedJsonFilename}, string(os.PathSeparator))

	data, err := ioutil.ReadFile(path)
	if err != nil {
		ExitWithError(err)
	}

	if err := json.Unmarshal(data, &al); err != nil {
		ExitWithError(err)
	}

	return
}
