package main

import (
	"encoding/base64"
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

	encryptedEnvVariable, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		log.Fatalf("Error invalid license:", err)
	}

	decryptedValue, err := decrypt.Data(encryptedEnvVariable, "env")
	if err != nil {
		log.Fatalf("Error unknown license: %v", err)
	}

	log.Printf("INFO: Using license token: %s\n", token)

	command := strings.Fields(string(decryptedValue))

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
