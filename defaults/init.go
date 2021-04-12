package defaults

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/logston/aws-aliased-profiles/common"
)

func InitProfileTemplate() {
	err := os.Mkdir(common.GetAPPath(), 0755)
	common.ExitWithError(err)

	path := common.GetAPPath(common.ConfigFilename)

	err = ioutil.WriteFile(path, []byte(common.DefaultProfileTemplate), 0644)
	if err != nil {
		common.ExitWithError(err)
	}

	fmt.Printf("New template placed at %s\n", path)
}
