package upsert

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"text/template"

	"github.com/logston/aws-aliased-profiles/common"
)

func AWSConfig() {
	t := GetProfileTemplate()

	al := common.ReadAccountList()

	profiles := GetProfileBuffer(t, al)

	config := ReadAWSConfig()

	config = InsertProfiles(config, profiles)

	WriteAWSConfig(config)
}

func GetProfileTemplate() *template.Template {
	home, err := os.UserHomeDir()
	if err != nil {
		common.ExitWithError(err)
	}

	path := strings.Join([]string{home, ".aws", common.AccountsAliasedConfigFilename}, string(os.PathSeparator))

	t, err := template.New("aliased-accounts.tmpl").ParseFiles(path)
	if err != nil {
		fmt.Printf("Looks like there is no template at '%s'\nHere's an example of what to place there...\n", path)
		fmt.Println(common.DefaultProfileTemplate)
		os.Exit(1)
	}

	return t
}

func GetProfileBuffer(t *template.Template, al []*common.Account) string {
	var b bytes.Buffer

	for _, a := range al {
		err := t.Execute(&b, a)
		if err != nil {
			common.ExitWithError(err)
		}
		_, err = b.WriteString("\n")
		if err != nil {
			common.ExitWithError(err)
		}
	}

	return b.String()
}

func ReadAWSConfig() string {
	home, err := os.UserHomeDir()
	if err != nil {
		common.ExitWithError(err)
	}

	path := strings.Join([]string{home, ".aws", common.AWSConfigFilename}, string(os.PathSeparator))

	buf, err := ioutil.ReadFile(path)
	if err != nil {
		common.ExitWithError(err)
	}

	return string(buf)
}

func WriteAWSConfig(config string) {
	home, err := os.UserHomeDir()
	if err != nil {
		common.ExitWithError(err)
	}

	path := strings.Join([]string{home, ".aws", common.AWSConfigFilename}, string(os.PathSeparator))

	err = ioutil.WriteFile(path, []byte(config), 0644)
	if err != nil {
		common.ExitWithError(err)
	}
}

func InsertProfiles(config, profiles string) string {
	// If the delimiter is not in the config file, append it.
	if !strings.Contains(config, common.AWSConfigDelimiter) {
		config = strings.Join(
			[]string{config, common.AWSConfigDelimiter, common.AWSConfigDelimiter},
			"\n",
		)
	}

	parts := strings.Split(config, common.AWSConfigDelimiter)
	if len(parts) != 3 {
		common.ExitWithError(fmt.Errorf("Unexpected number of parts after configuration split"))
	}

	// Add well defined amounts of padding around profiles section
	parts = []string{
		strings.Trim(parts[0], " \n") + "\n\n",
		"\n" + strings.Trim(profiles, " \n") + "\n",
		"\n\n" + strings.Trim(parts[2], " \n"),
	}
	config = strings.Join(parts, common.AWSConfigDelimiter)

	return config
}
