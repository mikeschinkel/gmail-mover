package gapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

const (
	ConfigBaseDirName   = ".config"
	CredentialsFileName = "credentials.json"
	TokensDirName       = "tokens"
	TokenFileTemplate   = "%s_token.json"
)

// getCredentialsPath returns the path to credentials file for this API instance
func (api *GMailAPI) getCredentialsPath() (credentialsPath string, err error) {
	var homeDir, configDir string

	homeDir, err = os.UserHomeDir()
	if err != nil {
		goto end
	}

	configDir = filepath.Join(homeDir, ConfigBaseDirName, api.appConfigDir)
	err = os.MkdirAll(configDir, 0700) // Private directory
	if err != nil {
		goto end
	}

	credentialsPath = filepath.Join(configDir, CredentialsFileName)

end:
	return credentialsPath, err
}

// getTokenPath returns the path to token file for this API instance and email
func (api *GMailAPI) getTokenPath(email string) (tokenPath string, err error) {
	var homeDir, configDir, tokenDir string

	homeDir, err = os.UserHomeDir()
	if err != nil {
		goto end
	}

	configDir = filepath.Join(homeDir, ConfigBaseDirName, api.appConfigDir)
	tokenDir = filepath.Join(configDir, TokensDirName)
	err = os.MkdirAll(tokenDir, 0700) // Private directory
	if err != nil {
		goto end
	}

	tokenPath = filepath.Join(tokenDir, fmt.Sprintf(TokenFileTemplate, email))

end:
	return tokenPath, err
}

// GetGmailService creates an authenticated Gmail service for the specified email
func (api *GMailAPI) GetGmailService(email string) (service *gmail.Service, err error) {
	var config *oauth2.Config
	var token *oauth2.Token
	var client *http.Client

	config, err = api.loadCredentials()
	if err != nil {
		goto end
	}

	token, err = api.getToken(config, email)
	if err != nil {
		goto end
	}

	client = config.Client(context.Background(), token)
	service, err = gmail.NewService(context.Background(), option.WithHTTPClient(client))

end:
	return service, err
}

func (api *GMailAPI) loadCredentials() (config *oauth2.Config, err error) {
	var credentialsData []byte
	var credentialsPath string

	credentialsPath, err = api.getCredentialsPath()
	if err != nil {
		goto end
	}

	credentialsData, err = os.ReadFile(credentialsPath)
	if err != nil {
		err = fmt.Errorf("credentials not found at %s: %w\nPlease download OAuth2 credentials from Google Cloud Console and save as credentials.json in ~/.config/gmail-mover/", credentialsPath, err)
		goto end
	}

	config, err = google.ConfigFromJSON(credentialsData, gmail.GmailReadonlyScope, gmail.GmailModifyScope)

end:
	return config, err
}

func (api *GMailAPI) getToken(config *oauth2.Config, email string) (token *oauth2.Token, err error) {
	var tokenPath string

	tokenPath, err = api.getTokenPath(email)
	if err != nil {
		goto end
	}

	// Try to load existing token
	token, err = loadTokenFromFile(tokenPath)
	if err == nil && token.Valid() {
		goto end
	}

	logger.Info("Requesting access token", "email_address", email)
	// Get new token via OAuth flow
	token, err = getTokenFromWeb(config)
	if err != nil {
		goto end
	}

	// Save token for future use
	err = api.saveToken(email, token)

end:
	return token, err
}

func loadTokenFromFile(tokenPath string) (token *oauth2.Token, err error) {
	var f *os.File

	f, err = os.Open(tokenPath)
	if err != nil {
		goto end
	}
	defer mustCloseOrLog(f)

	token = &oauth2.Token{}
	err = json.NewDecoder(f).Decode(token)

end:
	return token, err
}

func getTokenFromWeb(config *oauth2.Config) (token *oauth2.Token, err error) {
	var authURL string
	var authCode string

	// Force out-of-band flow for CLI applications
	config.RedirectURL = "urn:ietf:wg:oauth:2.0:oob"

	authURL = config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser: \n%v\n", authURL)
	fmt.Print("Enter the authorization code: ")

	_, err = fmt.Scan(&authCode)
	if err != nil {
		goto end
	}

	token, err = config.Exchange(context.TODO(), authCode)

end:
	return token, err
}

func (api *GMailAPI) saveToken(email string, token *oauth2.Token) (err error) {
	var f *os.File
	var tokenPath string

	tokenPath, err = api.getTokenPath(email)
	if err != nil {
		goto end
	}

	err = os.MkdirAll(filepath.Dir(tokenPath), 0755)
	if err != nil {
		goto end
	}

	f, err = os.Create(tokenPath)
	if err != nil {
		goto end
	}
	defer mustCloseOrLog(f)

	err = json.NewEncoder(f).Encode(token)

end:
	return err
}
