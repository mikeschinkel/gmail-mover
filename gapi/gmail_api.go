package gapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

// GMailAPI provides Gmail API operations for a specific app configuration
type GMailAPI struct {
	appConfigDir string
	fileStore    FileStorer
}

// NewGMailAPI creates a new GMailAPI instance with file store
func NewGMailAPI(appConfigDir string, fileStore FileStorer) *GMailAPI {
	return &GMailAPI{
		appConfigDir: appConfigDir,
		fileStore:    fileStore,
	}
}

// GetGmailService creates an authenticated Gmail service for the specified email
func (api *GMailAPI) GetGmailService(email string) (service *gmail.Service, err error) {
	var config *oauth2.Config
	var token *oauth2.Token
	var client *http.Client

	if !api.fileStore.Exists(CredentialsFileName) {
		err = api.setupCredentials()
		if err != nil {
			goto end
		}
	}

	config, err = api.loadCredentials()
	if err != nil {
		goto end
	}

	token, err = api.getToken(config, email)
	if err != nil {
		goto end
	}

	// Create a token source that automatically refreshes and saves tokens
	client = oauth2.NewClient(context.Background(), &savingTokenSource{
		base:          config.TokenSource(context.Background(), token),
		api:           api,
		tokenFilename: fmt.Sprintf(TokenFileTemplate, email),
	})
	service, err = gmail.NewService(context.Background(), option.WithHTTPClient(client))

end:
	return service, err
}

func (api *GMailAPI) setupCredentials() (err error) {
	var credentialsJSON string
	var credentialsRaw json.RawMessage

	fmt.Println("Gmail Mover requires OAuth2 credentials to access Gmail.")
	fmt.Println("Please follow these steps:")
	fmt.Println("1. Go to https://console.cloud.google.com/")
	fmt.Println("2. Create a new project or select an existing one")
	fmt.Println("3. Enable the Gmail API")
	fmt.Println("4. Create OAuth 2.0 Client ID credentials (Desktop Application)")
	fmt.Println("5. Download the credentials JSON file")
	fmt.Println("")
	fmt.Print("Paste the contents of your credentials JSON file here and press Enter: ")

	_, err = fmt.Scanln(&credentialsJSON)
	if err != nil {
		err = fmt.Errorf("failed to read credentials input: %w", err)
		goto end
	}

	// Validate the JSON by trying to parse it
	credentialsRaw = json.RawMessage(credentialsJSON)
	_, err = google.ConfigFromJSON(credentialsRaw, gmail.MailGoogleComScope)
	if err != nil {
		err = fmt.Errorf("invalid credentials JSON: %w", err)
		goto end
	}

	// Save the credentials
	err = api.fileStore.Save(CredentialsFileName, credentialsRaw)
	if err != nil {
		err = fmt.Errorf("failed to save credentials: %w", err)
		goto end
	}

	fmt.Println("Credentials saved successfully!")

end:
	return err
}

func (api *GMailAPI) loadCredentials() (config *oauth2.Config, err error) {
	var credentialsJSON json.RawMessage

	err = api.fileStore.Load(CredentialsFileName, &credentialsJSON)
	if err != nil {
		err = fmt.Errorf("failed to load credentials: %w", err)
		goto end
	}

	config, err = google.ConfigFromJSON(credentialsJSON, gmail.MailGoogleComScope)
	if err != nil {
		err = fmt.Errorf("failed to parse credentials: %w", err)
	}

end:
	return config, err
}

func (api *GMailAPI) getToken(config *oauth2.Config, email string) (token *oauth2.Token, err error) {
	var tokenFilename string

	tokenFilename = fmt.Sprintf(TokenFileTemplate, email)

	// Try to load existing token
	if api.fileStore.Exists(tokenFilename) {
		token = &oauth2.Token{}
		err = api.fileStore.Load(tokenFilename, token)
		if err == nil && token.Valid() {
			goto end
		}
		// If load failed or token invalid, continue to get new token
	}

	logger.Info("Requesting access token", "email_address", email)
	// Get new token via OAuth flow
	token, err = getTokenFromWeb(config)
	if err != nil {
		goto end
	}

	// Save token for future use
	err = api.fileStore.Save(tokenFilename, token)

end:
	return token, err
}

// savingTokenSource wraps an oauth2.TokenSource to automatically save refreshed tokens
type savingTokenSource struct {
	base          oauth2.TokenSource
	api           *GMailAPI
	tokenFilename string
}

// Token returns a token, automatically refreshing if needed and saving any updates
func (s *savingTokenSource) Token() (token *oauth2.Token, err error) {
	token, err = s.base.Token()
	if err != nil {
		goto end
	}

	// Save the token (which may have been refreshed)
	err = s.api.fileStore.Save(s.tokenFilename, token)
	if err != nil {
		logger.Warn("Failed to save refreshed token", "error", err, "filename", s.tokenFilename)
	}

end:
	return token, err
}
