package utils

import (
	"io/ioutil"
	"log"
	"os"
	"os/exec"

	"golang.org/x/mod/modfile"
)

func CreateMod(path string) error {
	exec.Command("rm", "go.mod", "go.sum").Run()
	goModCmd := exec.Command("go", "mod", "init", path)
	return goModCmd.Run()
}

func GetModPath() string {
	f, err := os.Open("go.mod")
	if err != nil {
		log.Println(err)
		return ""
	}
	modData, err := ioutil.ReadAll(f)
	if err != nil {
		log.Println(err)
		return ""
	}
	return modfile.ModulePath(modData)
}
