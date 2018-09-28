package buildpack_test

import (
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/shurcooL/httpfs/vfsutil"
	"github.com/snowdrop/spring-boot-cloud-devex/pkg/buildpack"
)

var (
	templateFiles []string
	project       = "java"
)

func TestVfsBuildpack(t *testing.T) {
	tExpectedFiles := []string{
		"java/imagestream",
		"java/route",
		"java/service",
	}

	tFiles := walkTree()

	for i := range tExpectedFiles {
		if tExpectedFiles[i] != tFiles[i] {
			t.Errorf("Template was incorrect, got: '%s', want: '%s'.", tFiles[i], tExpectedFiles[i])
		}
	}
}

func walkTree() []string {
	var fs http.FileSystem = buildpack.Assets

	vfsutil.Walk(fs, project, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("can't stat file %s: %v\n", path, err)
			return nil
		}

		if fi.IsDir() {
			return nil
		}

		fmt.Println(path)
		templateFiles = append(templateFiles, path)
		return nil
	})
	return templateFiles
}
