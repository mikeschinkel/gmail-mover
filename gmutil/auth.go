package gmutil

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

func GetGmailService(email string) (service *gmail.Service, err error) {
	var config *oauth2.Config
	var token *oauth2.Token
	var client *http.Client

	config, err = loadCredentials()
	if err != nil {
		goto end
	}

	token, err = getToken(config, email)
	if err != nil {
		goto end
	}

	client = config.Client(context.Background(), token)
	service, err = gmail.NewService(context.Background(), option.WithHTTPClient(client))

end:
	return service, err
}

func loadCredentials() (config *oauth2.Config, err error) {
	var credentialsData []byte

	credentialsData, err = os.ReadFile("credentials.json")
	if err != nil {
		goto end
	}

	config, err = google.ConfigFromJSON(credentialsData, gmail.GmailReadonlyScope, gmail.GmailModifyScope)

end:
	return config, err
}

func getToken(config *oauth2.Config, email string) (token *oauth2.Token, err error) {
	var tokenPath string

	tokenPath = filepath.Join("tokens", fmt.Sprintf("%s_token.json", email))

	// Try to load existing token
	token, err = loadTokenFromFile(tokenPath)
	if err == nil && token.Valid() {
		goto end
	}

	// Get new token via OAuth flow
	token, err = getTokenFromWeb(config)
	if err != nil {
		goto end
	}

	// Save token for future use
	err = saveToken(tokenPath, token)

end:
	return token, err
}

func loadTokenFromFile(tokenPath string) (token *oauth2.Token, err error) {
	var f *os.File

	f, err = os.Open(tokenPath)
	if err != nil {
		goto end
	}
	defer f.Close()

	token = &oauth2.Token{}
	err = json.NewDecoder(f).Decode(token)

end:
	return token, err
}

func getTokenFromWeb(config *oauth2.Config) (token *oauth2.Token, err error) {
	var authURL string
	var authCode string

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

func saveToken(tokenPath string, token *oauth2.Token) (err error) {
	var f *os.File

	err = os.MkdirAll(filepath.Dir(tokenPath), 0755)
	if err != nil {
		goto end
	}

	f, err = os.Create(tokenPath)
	if err != nil {
		goto end
	}
	defer f.Close()

	err = json.NewEncoder(f).Encode(token)

end:
	return err
}
