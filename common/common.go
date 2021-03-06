package common

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

const (
	DirName                = "aliased-profiles"
	StateFilename          = "state.json"
	ConfigFilename         = "config.tmpl"
	AWSConfigFilename      = "config"
	DefaultProfileTemplate = `
{{- define "profileBody" }}
cli_pager=
source_profile = default
{{- if .HasTagKeyValue "environment" "staging" }}
role_arn = arn:aws:iam::{{ .Id }}:role/Staging
{{ else }}
role_arn = arn:aws:iam::{{ .Id }}:role/Production
{{ end -}}
{{ end -}}

[profile {{ .Id }}]
{{- template "profileBody" . -}}

{{- if .Alias }}
[profile {{ .Alias }}]
{{- template "profileBody" . -}}
{{ end -}}
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

func (a *Account) HasTagKeyValue(key, value string) bool {
	for _, t := range a.Tags {
		if t.Key == key && t.Value == value {
			return true
		}
	}

	return false
}

func NewCtx() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		sCh := make(chan os.Signal, 1)
		signal.Notify(sCh, syscall.SIGINT, syscall.SIGTERM)
		<-sCh
		cancel()
	}()
	return ctx
}

func CheckContext(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}

func ExitWithError(err error) {
	if err != nil {
		panic(err)
	}
}

func WriteAccountList(al []*Account) {
	data, err := json.MarshalIndent(al, "", "    ")
	if err != nil {
		ExitWithError(err)
	}

	path := GetAPPath(StateFilename)

	if err := ioutil.WriteFile(path, data, 0644); err != nil {
		ExitWithError(err)
	}
}

func ReadAccountList() (al []*Account) {
	path := GetAPPath(StateFilename)

	data, err := ioutil.ReadFile(path)
	if err != nil {
		ExitWithError(err)
	}

	if err := json.Unmarshal(data, &al); err != nil {
		ExitWithError(err)
	}

	return
}

func GetAWSPath(files ...string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		ExitWithError(err)
	}

	parts := []string{home, ".aws"}
	for _, file := range files {
		parts = append(parts, file)
	}

	return strings.Join(parts, string(os.PathSeparator))
}

func GetAPPath(files ...string) string {
	parts := []string{DirName}
	for _, file := range files {
		parts = append(parts, file)
	}

	return GetAWSPath(parts...)
}
