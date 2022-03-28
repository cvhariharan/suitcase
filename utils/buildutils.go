package utils

import (
	"errors"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/hashicorp/go-version"
	"github.com/rs/xid"
)

const MIN_VERSION = "1.16"

// GoVersion returns the installed version of Go
func GoVersion() (*version.Version, error) {
	re, err := regexp.Compile(`go\d.\d{2}`)
	if err != nil {
		return nil, err
	}

	versionOut, err := exec.Command("go", "version").Output()
	if err != nil {
		return nil, err
	}

	goVersion := strings.Replace(string(re.Find(versionOut)), "go", "", -1)
	return version.NewSemver(goVersion)
}

// VersionCheck checks if the provided Go version is greator than
// the required minimum version
func IsVersionOk(current string) bool {
	baseVersion, err := version.NewSemver(MIN_VERSION)
	if err != nil {
		log.Println(err)
		return false
	}

	currVersion, err := version.NewSemver(current)
	if err != nil {
		log.Println(err)
		return false
	}

	return currVersion.GreaterThanOrEqual(baseVersion)
}

// CurrentVersionCheck checks if the intalled Go version is greator than
// the required minimum version
func IsCurrentVersionOk() bool {
	currVersion, err := GoVersion()
	if err != nil {
		log.Println(err)
		return false
	}
	return IsVersionOk(currVersion.String())
}

func CreateBuildDir() (string, error) {
	gid := xid.New().String()

	if _, err := os.Stat(gid); err == nil {
		os.RemoveAll(gid)
	}

	err := os.Mkdir(gid, 0755)
	return gid, err
}

func Build() error {
	if !IsCurrentVersionOk() {
		return errors.New("go versions is less than " + MIN_VERSION)
	}

	gid := xid.New().String()
	err := CreateMod(gid)
	if err != nil {
		return err
	}

	stop := make(chan int)
	go showSpinner(stop)

	err = exec.Command("go", "mod", "tidy").Run()
	if err != nil {
		return err
	}

	err = exec.Command("go", "build", "-o", "secure-case").Run()
	if err != nil {
		return err
	}

	stop <- 1

	exec.Command("rm", "go.mod", "go.sum", "main.go").Run()
	return nil
}

func showSpinner(stop chan int) {
	s := spinner.New(spinner.CharSets[36], 100*time.Millisecond)
	s.Prefix = "Building... "
	s.Start()

	for {
		select {
		case <-stop:
			s.Stop()
			break
		default:
			time.Sleep(1 * time.Second)
		}
	}
}
