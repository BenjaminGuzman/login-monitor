package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
)

func configFlags(configPath *string) {
	flag.StringVar(
		configPath,
		"config",
		"config.json",
		"Config file to use",
	)
}

// ensures permissions for the executable that started the process are set to 500 (r-x --- ---)
func checkPermissions() error {
	execPath, err := os.Executable()
	if err != nil {
		return err
	}
	stat, err := os.Stat(execPath)
	if err != nil {
		return err
	}

	// if permissions are not ok, change them
	if permissions := stat.Mode().Perm(); permissions&0500 != 0 {
		err := os.Chmod(execPath, 0500)
		if err != nil {
			return err
		}
	}

	// try to ensure owner is root
	if err = os.Lchown(execPath, 0, 0); err != nil {
		return err
	}

	return nil
}

func main() {
	if err := checkPermissions(); err != nil {
		fmt.Println("Error while checking permissions.", err)
	}

	var configFile string
	configFlags(&configFile)
	flag.Parse()

	configReader, err := os.Open(configFile)
	if err != nil {
		log.Fatalf("Error while reading config file '%s'. %s", configFile, err)
	}
	email := NewEmail(&GmailOAuth2Strategy{})
	err = json.NewDecoder(configReader).Decode(&email)
	if err != nil {
		log.Fatalf("Error while decoding config file '%s'. %s", configFile, err)
	}

	if email.IsPGPCandidate() {
		if _, err := email.SendPGPEmail(); err != nil {
			log.Fatalf("Error while sending PGP email. Config file: '%s'. %s", configFile, err)
		}
	} else {
		if _, err := email.SendEmail(); err != nil {
			log.Fatalf("Error while sending plain text email. Config file: '%s'. %s", configFile, err)
		}
	}
}
