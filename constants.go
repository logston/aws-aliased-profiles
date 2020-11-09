package main

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
