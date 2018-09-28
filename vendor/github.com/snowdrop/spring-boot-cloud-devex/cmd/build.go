package cmd

import (
	"archive/tar"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/snowdrop/spring-boot-cloud-devex/pkg/buildpack"
	"github.com/snowdrop/spring-boot-cloud-devex/pkg/buildpack/types"
	"github.com/snowdrop/spring-boot-cloud-devex/pkg/common/oc"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"path/filepath"
)

func init() {
	buildCmd := &cobra.Command{
		Use:     "build [flags]",
		Short:   "Build an image of the application",
		Long:    `Build an image of the application.`,
		Example: ` sb build`,
		Args:    cobra.RangeArgs(0, 1),
		Run: func(cmd *cobra.Command, args []string) {

			log.Info("Build command called")

			setup := Setup()
			log.Debugf("Namespace: %s", setup.Application.Namespace)

			// Create Build
			log.Info("Create Build resource")
			buildpack.CreateBuild(setup.RestConfig, setup.Application)

			//Create target imageStream
			images := []types.Image{
				*buildpack.CreateTypeImage(false, setup.Application.Name, "latest", "", false),
			}
			buildpack.CreateImageStreamTemplate(setup.RestConfig, setup.Application, images)

			// Generate tar and next start the build using it
			//r := generateTar("")

			// TODO - Add a watch to check if the build has been created

			// Start the build
			args = []string{"start-build", setup.Application.Name, "--from-dir=" + oc.Client.Pwd, "--follow"}
			log.Infof("Starr-build cmd : %s", args)
			oc.ExecCommand(oc.Command{Args: args})
		},
	}

	// Add a defined annotation in order to appear in the help menu
	buildCmd.Annotations = map[string]string{"command": "build"}

	rootCmd.AddCommand(buildCmd)
}

func generateTar(dirToskip string) (f *os.File) {
	// Create and add some files to the archive.
	current, _ := os.Getwd()
	dir := filepath.Join(current, "spring-boot")
	files := walkTree(dir, dirToskip)

	return createTar(current, files)
}

// exists returns whether the given file or directory exists or no
func isDir(path string) bool {
	info, _ := os.Stat(path)
	if info.IsDir() {
		return true
	} else {
		return false
	}
}

// walk through the dir to doscovery files recursively
func walkTree(dir string, dirToskip string) []string {
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", dir, err)
			return err
		}
		if info.IsDir() && info.Name() == dirToskip {
			fmt.Printf("skipping a dir without errors: %+v \n", info.Name())
			return filepath.SkipDir
		}
		//fmt.Printf("visited file: %q\n", path)
		files = append(files, path)
		return nil
	})

	if err != nil {
		fmt.Printf("error walking the path %q: %v\n", dir, err)
	}
	return files
}

func createTar(current string, files []string) (f *os.File) {
	// Open TAR file
	f, err := os.OpenFile(current+"/touch.tar", os.O_RDWR, os.ModePerm)
	if err != nil {
		log.Fatalln(err)
	}
	// Create Tar writer
	tw := tar.NewWriter(f)

	// Iterate through the files of the dir
	for _, file := range files {
		if !isDir(file) {
			// Read file's content
			content, err := ioutil.ReadFile(file)
			if err != nil {
				log.Fatal(err)
			}
			f, _ := os.Stat(file)
			fmt.Println("File name : ", f.Name())

			h := &tar.Header{
				Name: f.Name(),
				Mode: 0600,
				Size: int64(len(content)),
			}

			if err := tw.WriteHeader(h); err != nil {
				log.Fatal("Can't write header : ", h, " - ", err.Error())
			}

			if _, err := tw.Write([]byte(content)); err != nil {
				log.Fatal(err)
			}
		}
	}

	if err := tw.Close(); err != nil {
		log.Fatal(err)
	}

	return f
}
