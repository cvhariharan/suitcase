package main

import (
	"embed"
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"os"
	"strings"

	"filippo.io/age"
	"github.com/cvhariharan/suitcase/utils"
)

//go:embed templates/*
var fs embed.FS

type stringArr []string

func (s *stringArr) String() string {
	return fmt.Sprint(*s)
}

func (s *stringArr) Set(val string) error {
	*s = append(*s, val)
	return nil
}

var usage = `
suitcase (-p PUBLICKEY)... -i INPUT

Options:
	-p, --public-key        Public keys of the recipients. Can be repeated. Should begin with age1
	-i, --input             Input file to encrypt

The output will be an executable in the working directory with the contents of the file
embedded in it. No need for any client-side software to decrypt the file.
`

func main() {
	var pubKey stringArr
	var inputFile string

	flag.Usage = func() {
		fmt.Println(usage)
	}

	if len(os.Args) == 1 {
		flag.Usage()
		os.Exit(1)
	}

	flag.Var(&pubKey, "public-key", "Public key to encrypt the file")
	flag.Var(&pubKey, "p", "Public key to encrypt the file")
	flag.StringVar(&inputFile, "input", "", "Input file name")
	flag.StringVar(&inputFile, "i", "", "Input file name")
	flag.Parse()

	if len(pubKey) == 0 || inputFile == "" {
		log.Fatal("Public key and input file name cannot be empty")
	}

	var recipients []age.Recipient
	for _, v := range pubKey {
		recipient, err := age.ParseX25519Recipient(v)
		checkError(err)
		recipients = append(recipients, recipient)
	}

	encryptedFileName := inputFile + ".enc"
	f, err := os.Create(encryptedFileName)
	checkError(err)
	defer f.Close()

	encrypted, err := age.Encrypt(f, recipients...)
	checkError(err)

	in, err := os.Open(inputFile)
	checkError(err)
	defer in.Close()

	if _, err = io.Copy(encrypted, in); err != nil {
		log.Fatal(err)
	}
	encrypted.Close()

	fmap := template.FuncMap{
		"OutputName": func(s string) string {
			return strings.Replace(s, ".enc", "", -1)
		},
	}
	templateData, err := fs.ReadFile("templates/main.tmpl")
	checkError(err)

	t := template.Must(template.New("main.tmpl").Funcs(fmap).Parse(string(templateData)))

	out, err := os.Create("main.go")
	checkError(err)

	err = t.Execute(out, struct {
		FileName string
	}{f.Name()})
	checkError(err)

	err = utils.Build()
	checkError(err)

	checkError(os.Remove(encryptedFileName))
}

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
