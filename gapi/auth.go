package gapi

import (
	"context"
	"fmt"

	"golang.org/x/oauth2"
)

const (
	CredentialsFileName = "credentials.json"
	TokenFileTemplate   = "tokens/token-%s.json"
)

func (api *GMailAPI) getTokenFromWeb(account EmailAddress, config *oauth2.Config) (token *oauth2.Token, err error) {
	var authURL string
	var authCode string

	// Force out-of-band flow for CLI applications
	config.RedirectURL = "urn:ietf:wg:oauth:2.0:oob"

	authURL = config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)

	ensureOutput()
	writer.Printf("Go to the following link in your browser: \n%v\n", authURL)
	writer.Printf("Enter the authorization code for %s: ", account)

	_, err = fmt.Scan(&authCode)
	if err != nil {
		goto end
	}

	token, err = config.Exchange(context.TODO(), authCode)

end:
	return token, err
}
