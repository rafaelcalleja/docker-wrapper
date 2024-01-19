package main

import (
	"encoding/base64"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/getsops/sops/v3/decrypt"
	gof "github.com/jessevdk/go-flags"
)

type TokenDecryptModule struct {
	DefaultRunModule
}

// HandleRun implements the WrapperRunModule interface
func (m *TokenDecryptModule) HandleRun(flags *DockerFlags, runFlags *DockerRunCommandFlags) []string {
	token := os.Getenv("TOKEN")
	passphrase := os.Getenv("GPG_PASSPHRASE")

	if passphrase != "" {
		gpgWrapper, err := ioutil.TempFile("", "gpg-wrapper")
		if err != nil {
			log.Fatalf("Error creating wrapper:", err)
		}
		defer os.Remove(gpgWrapper.Name())

		scriptContent := `#!/bin/bash
gpg --batch --pinentry-mode loopback --passphrase $GPG_PASSPHRASE --yes $@`
		err = gpgWrapper.Chmod(0755)
		if err != nil {
			log.Fatalf("Error changing perms wrapper:", err)
		}

		_, err = gpgWrapper.WriteString(scriptContent)

		if err != nil {
			log.Fatalf("Error writing GPG wrapper:", err)
		}

		gpgWrapper.Close()

		os.Setenv("SOPS_GPG_EXEC", gpgWrapper.Name())
	}

	encryptedEnvVariable, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		log.Fatalf("Error invalid license:", err)
	}

	decryptedValue, err := decrypt.Data(encryptedEnvVariable, "dotenv")
	if err != nil {
		log.Fatalf("Error unknown license: %v", err)
	}

	envMap := make(map[string]string)

	for _, line := range strings.Split(string(decryptedValue), "\n") {
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			envMap[parts[0]] = parts[1]
		}
	}

	log.Printf("INFO: Using license token: %s\n", token)

	command := strings.Fields(envMap["token"])

	valueToFind := runFlags.Args.Image
	imageIndex := m.getIndexOf(valueToFind, newDockerArgs)

	var myDockerFlags DockerFlags

	var myDockerRunFlags DockerRunCommandFlags

	myOptsParser := gof.NewParser(&myDockerFlags, gof.PassDoubleDash|gof.IgnoreUnknown|gof.PassAfterNonOption)

	myOptsParser.AddCommand("run",
		"",
		"",
		&myDockerRunFlags)

	_, err = myOptsParser.ParseArgs(command[1:])

	if err != nil {
		log.Fatalf("Error expired license: %v", err)
	}

	newDockerArgs[imageIndex] = myDockerRunFlags.Args.Image
	newDockerArgs = append(newDockerArgs[:imageIndex+1], myDockerRunFlags.Args.CmdArgs...)

	return command[2:m.getIndexOf(myDockerRunFlags.Args.Image, command)]
}

func (m *TokenDecryptModule) getIndexOf(valueToFind string, sliceToFind []string) int {
	imageIndex := -1

	for i, v := range sliceToFind {
		if v == valueToFind {
			imageIndex = i
			break
		}
	}
	return imageIndex
}

// init calls RegisterRunModule
func init() {
	RegisterRunModule(&TokenDecryptModule{})
}
