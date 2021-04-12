package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/logston/aws-aliased-profiles/common"
	"github.com/logston/aws-aliased-profiles/defaults"
	"github.com/logston/aws-aliased-profiles/fetch"
	"github.com/logston/aws-aliased-profiles/upsert"
)

var rootCmd = &cobra.Command{
	Use:   "aws-aliased-profiles [command]",
	Short: "quickly update your aws config with all your OU accounts' aliases",
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "init ~/.aws/aliased-profiles/config.tpml",
	Run: func(cmd *cobra.Command, args []string) {
		defaults.InitProfileTemplate()
	},
}

var fetchCmd = &cobra.Command{
	Use:   "fetch <profile> <accountRole>",
	Short: "fetch data from organizational unit",
	Long: `fetch data from AWS

Fetch data for each account using the <profile> profile to assume the
<accountRole> in each account in the organizational unit.

<profile> is the profile from which STS tokens for assuming roles can be generated.
This profile will also be used to list all the accounts in an organizational unit.

<accountRole> is the role name to assume in each account such that alias
information can be gathered.
`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		fetch.AliasToAccountMap(common.NewCtx(), args[0], args[1])
	},
}

var upsertCmd = &cobra.Command{
	Use:   "upsert",
	Short: "upsert ~/.aws/config with data from organizational unit",
	Run: func(cmd *cobra.Command, args []string) {
		upsert.AWSConfig()
	},
}

func Execute() {
	rootCmd.AddCommand(
		fetchCmd,
		upsertCmd,
		initCmd,
	)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
