package email

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
	"io/ioutil"
	"os"
)

type GmailOAuth2Strategy struct {
	gmailService *gmail.Service
}

// Init initiates a new gmail.Service (required by other methods)
// 1st param: filepath to oauth2 config file (client id, client secret, endpoint, redirect url...)
// 2nd param: filepath to oauth2 token file (refresh token, access token...)
// Returns nothing
func (s *GmailOAuth2Strategy) Init(params ...interface{}) (interface{}, error) {
	configFilepath := fmt.Sprint(params[0])
	tokenFilepath := fmt.Sprint(params[1])

	// read config
	configFile, err := os.Open(configFilepath)
	if err != nil {
		return nil, fmt.Errorf("error reading Gmail Oauth 2 config: %w", err)
	}
	defer configFile.Close()

	b, err := ioutil.ReadAll(configFile)
	if err != nil {
		return nil, fmt.Errorf("error reading Gmail Oauth 2 config: %w", err)
	}
	config, err := google.ConfigFromJSON(b, gmail.GmailSendScope)
	if err != nil {
		return nil, fmt.Errorf("error reading Gmail Oauth 2 config: %w", err)
	}

	// read token
	tokenFile, err := os.Open(tokenFilepath)
	if err != nil {
		return nil, fmt.Errorf("error reading Gmail Oauth 2 token: %w", err)
	}
	defer tokenFile.Close()

	token := oauth2.Token{}
	err = json.NewDecoder(tokenFile).Decode(&token)
	if err != nil {
		return nil, fmt.Errorf("error while parsing Gmail Oauth 2 token: %w", err)
	}

	tokenSource := config.TokenSource(context.Background(), &token)

	service, err := gmail.NewService(context.Background(), option.WithTokenSource(tokenSource))
	if err != nil {
		return nil, fmt.Errorf("couldn't start Gmail Oauth 2 client: %w", err)
	}
	s.gmailService = service

	return nil, nil
}

// SendEmail sends the email with the gmail api. Returns nothing but an error, if any.
func (s *GmailOAuth2Strategy) SendEmail(payload []byte, sender string) (interface{}, error) {
	var msg gmail.Message
	msg.Raw = base64.StdEncoding.EncodeToString(payload)
	//str, _ := base64.StdEncoding.DecodeString(msg.Raw)
	//fmt.Println(string(str))
	call := s.gmailService.Users.Messages.Send(sender, &msg)
	if _, err := call.Do(); err != nil {
		return nil, fmt.Errorf("couldn't send email: %w", err)
	}
	return nil, nil
}

type GmailServiceAccountStrategy struct {
	gmailService *gmail.Service
}

// Init initiates a new gmail.Service (required by other methods)
// 1st param: path to credentials file
// Returns nothing
// THIS DOESN'T WORK YET. Issues associated: https://github.com/googleapis/google-api-go-client/issues/645
func (s *GmailServiceAccountStrategy) Init(params ...interface{}) (interface{}, error) {
	b, _ := os.ReadFile(fmt.Sprint(params[0]))
	config, err := google.JWTConfigFromJSON(b, gmail.GmailSendScope)
	service, err := gmail.NewService(
		context.Background(),
		option.WithHTTPClient(config.Client(context.Background())),
		//option.WithCredentialsFile(fmt.Print(params[0])), // this won't work either
		//option.WithScopes(gmail.GmailSendScope),
	)
	if err != nil {
		return nil, err
	}
	s.gmailService = service

	return nil, nil
}

func (s *GmailServiceAccountStrategy) SendEmail(payload []byte, sender string) (interface{}, error) {
	var msg gmail.Message
	msg.Raw = base64.StdEncoding.EncodeToString(payload)
	//str, _ := base64.StdEncoding.DecodeString(msg.Raw)
	//fmt.Println(string(str))
	call := s.gmailService.Users.Messages.Send(sender, &msg)
	if _, err := call.Do(); err != nil {
		return nil, err
	}
	return nil, nil
}
