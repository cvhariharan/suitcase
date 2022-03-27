package main

import (
	"embed"
	"flag"
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

func main() {
	log.SetFlags(log.Lshortfile)
	pubKey := flag.String("publicKey", "", "Private key to decrypt the file")
	inputFile := flag.String("input", "", "Input file name")
	flag.Parse()

	if *pubKey == "" || *inputFile == "" {
		log.Fatal("Public key and input file name cannot be empty")
	}

	recipient, err := age.ParseX25519Recipient(*pubKey)
	if err != nil {
		log.Fatal(err)
	}

	encryptedFileName := *inputFile + ".enc"
	f, err := os.Create(encryptedFileName)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	encrypted, err := age.Encrypt(f, recipient)
	if err != nil {
		log.Fatal(err)
	}

	in, err := os.Open(*inputFile)
	if err != nil {
		log.Fatal(err)
	}
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
	if err != nil {
		log.Fatal(err)
	}

	t := template.Must(template.New("main.tmpl").Funcs(fmap).Parse(string(templateData)))

	out, err := os.Create("main.go")
	if err != nil {
		log.Fatal(err)
	}

	err = t.Execute(out, struct {
		FileName string
	}{f.Name()})
	if err != nil {
		log.Fatal(err)
	}

	err = utils.Build()
	if err != nil {
		log.Fatal(err)
	}

	checkError(os.Remove(encryptedFileName))
}

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
