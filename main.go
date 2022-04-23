package main

import (
	"encoding/json"
	"flag"
	"fmt"
	log "github.com/sirupsen/logrus"
	configmodule "login-monitor/config"
	emailmodule "login-monitor/email"
	"os"
	"strings"
)

func configFlags(configPath, logLevel, strategy, gmailOAuth2Config, gmailOAuth2Token, goSMTPConfig *string) {
	flag.StringVar(
		configPath,
		"config",
		"config.json",
		"Config file to use",
	)
	flag.StringVar(
		logLevel,
		"log-level",
		"error",
		"Log level to use throughout the application. "+
			"Valid values are: trace, debug, info, warn, error, fatal, panic",
	)
	flag.StringVar(
		strategy,
		"strategy",
		"gmail-oauth2",
		"Strategy to use. Valid values are: gmail-oauth2, go-smtp",
	)

	// gmail-oauth2 config
	flag.StringVar(
		gmailOAuth2Config,
		"gmail-oauth2-config",
		"credentials.json",
		"Credentials file for the gmail-oauth2 strategy",
	)
	flag.StringVar(
		gmailOAuth2Token,
		"gmail-oauth2-token",
		"token.json",
		"Token file for the gmail-oauth2 strategy",
	)

	// go-smtp config
	flag.StringVar(
		goSMTPConfig,
		"go-smtp-config",
		"go-smtp-config.json",
		"Config file for go-smtp strategy",
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

func setLogLevel(logLevel string) {
	switch strings.TrimSpace(strings.ToLower(logLevel)) {
	case "trace":
		log.SetLevel(log.TraceLevel)
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "fatal":
		log.SetLevel(log.FatalLevel)
	case "panic":
		log.SetLevel(log.PanicLevel)
	default:
		fmt.Printf("%s is not recognized as a log level. Setting log level to error", logLevel)
		log.SetLevel(log.ErrorLevel)
	}
}

// if str is empty, def is returned. If str is not empty, str is returned
func stringDefault(str, def string) string {
	if str == "" {
		return def
	} else {
		return str
	}
}

func main() {
	if err := checkPermissions(); err != nil {
		fmt.Println("Error while checking permissions.", err)
	}

	var configFile, logLevel, strategy string      // general configuration
	var gmailOAuth2Config, gmailOAuth2Token string // gmail-oauth2 strategy config
	var goSMTPConfig string                        // go-smtp strategy config
	configFlags(&configFile, &logLevel, &strategy, &gmailOAuth2Config, &gmailOAuth2Token, &goSMTPConfig)
	flag.Parse()

	setLogLevel(logLevel)

	configReader, err := os.Open(configFile)
	if err != nil {
		log.Fatalf("Error while reading config file '%s'. %s", configFile, err)
	}
	defer configReader.Close()

	var email *emailmodule.Email

selectStrategy:
	switch strings.TrimSpace(strings.ToLower(strategy)) {
	case "go-smtp":
		// read go-smtp config
		smtpConfigF, err := os.Open(goSMTPConfig)
		if err != nil {
			log.Fatalf("Couldn't read file %s: %s", goSMTPConfig, err)
		}
		defer smtpConfigF.Close()
		smtpConfig := configmodule.GoSMTPConfig{}
		err = json.NewDecoder(smtpConfigF).Decode(&smtpConfig)
		if err != nil {
			log.Fatalf("Error while parsing JSON config %s: %s", goSMTPConfig, err)
		}

		email = emailmodule.NewEmail(&emailmodule.GoSMTPStrategy{})
		_, err = email.InitStrategy(
			smtpConfig.Identity,
			smtpConfig.Username,
			smtpConfig.Password,
			stringDefault(smtpConfig.Host, "127.0.0.1"),
			stringDefault(smtpConfig.Port, "25"),
		)
		if err != nil {
			log.Fatalf(
				"Error while initiating go-smtp strategy. Config file: '%s'. %s",
				goSMTPConfig,
				err,
			)
		}
	case "gmail-oauth2":
		email = emailmodule.NewEmail(&emailmodule.GmailOAuth2Strategy{})
		_, err = email.InitStrategy(gmailOAuth2Config, gmailOAuth2Token)
		if err != nil {
			log.Fatalf(
				"Error while initiating gmail-oauth2 strategy. Config file: '%s', token file: '%s'. %s",
				gmailOAuth2Config,
				gmailOAuth2Token,
				err,
			)
		}
	default:
		log.Warnf("%s is not recognized as a valid strategy. Using default gmail-oauth2 strategy", strategy)
		strategy = "gmail-oauth2"
		goto selectStrategy
	}

	config := configmodule.EmailConfig{}
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
