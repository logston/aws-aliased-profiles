package main

import (
	"os"
	"strings"
)

//
// Insert named profiles into `~/.aws/config` file
//   - in between # ------ aws-named-profiles ------ start and end tags
//   - Alert if file does not exist
//   - Should handle config being in other dirs than ~/.aws

// Command should be able to redownload alias listing
// Command should be able to regenerate map

// profile_template = ```
// [profile {{ alias }}]
// role_arn = arn:aws:iam::{{ accountId }}:role/MyFavRoleToAssume
// source_profile = default
// ```

func GetProfileTemplate() {
	home, err := os.UserHomeDir()
	if err != nil {
		ExitWithError(err)
	}

	_ = strings.Join([]string{home, ".aws", AccountsAliasedConfigFilename}, string(os.PathSeparator))
}
