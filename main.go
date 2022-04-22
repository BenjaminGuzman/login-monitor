package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
)

func configFlags(configPath, gmailOAuth2Config, gmailOAuth2Token *string) {
	flag.StringVar(
		configPath,
		"config",
		"config.json",
		"Config file to use",
	)
	flag.StringVar(
		gmailOAuth2Config,
		"gmail-oauth2-config",
		"credentials.json",
		"Credentials file for the Gmail OAuth2 strategy",
	)
	flag.StringVar(
		gmailOAuth2Token,
		"gmail-oauth2-token",
		"token.json",
		"Token file for the Gmail Oauth2 strategy",
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

	var configFile, gmailOAuth2Config, gmailOAuth2Token string
	configFlags(&configFile, &gmailOAuth2Config, &gmailOAuth2Token)
	flag.Parse()

	configReader, err := os.Open(configFile)
	defer configReader.Close()
	if err != nil {
		log.Fatalf("Error while reading config file '%s'. %s", configFile, err)
	}
	email := NewEmail(&GmailOAuth2Strategy{}) // TODO handle other strategies
	_, err = email.InitStrategy(gmailOAuth2Config, gmailOAuth2Token)
	if err != nil {
		log.Fatalf(
			"Error while initiating Gmail Oauth 2 strategy. Config file: '%s', token file: '%s'. %s",
			gmailOAuth2Config,
			gmailOAuth2Token,
			err,
		)
	}

	config := EmailConfig{}
	err = json.NewDecoder(configReader).Decode(&config)
	email.InitFromConfig(&config)
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
