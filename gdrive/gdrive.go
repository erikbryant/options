package gdrive

//
// Create 'options' project in GCP
// Create Oauth tokens
// Save Oauth credentials to credentils.json
//
// Go example
//   https://cloud.google.com/docs/authentication?_ga=2.56685738.-1275246831.1637708616
// GCP Console
//   https://console.cloud.google.com/home/dashboard?project=options-333023
//
// https://developers.google.com/drive/api/v3/quickstart/go
// go get -u google.golang.org/api/drive/v3
// go get -u golang.org/x/oauth2/google
// go run quickstart.go
//

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/browser"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	tok := &oauth2.Token{}

	err = json.NewDecoder(f).Decode(tok)

	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) error {
	fmt.Printf("Saving credential file to: %s\n", path)

	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("Unable to cache oauth token: %v", err)
	}

	defer f.Close()

	json.NewEncoder(f).Encode(token)

	return nil
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) (*oauth2.Token, error) {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)

	fmt.Printf("Opening authorization link in your browser: \n%v\n\n", authURL)
	browser.OpenURL(authURL)

	fmt.Println("Enter the authorization code:")

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		return nil, fmt.Errorf("Unable to read authorization code %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		return nil, fmt.Errorf("Unable to retrieve token from web %v", err)
	}

	return tok, nil
}

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) (*http.Client, error) {
	// tokFile stores the user's access and refresh tokens. If it does not exist
	// or has expired we will create it.
	tokFile := "token.json"

	tok, err := tokenFromFile(tokFile)
	if err != nil || tok.Expiry.Before(time.Now()) {
		tok, err = getTokenFromWeb(config)
		if err != nil {
			return nil, err
		}
		err = saveToken(tokFile, tok)
		if err != nil {
			return nil, err
		}
	}

	return config.Client(context.Background(), tok), nil
}

// service returns a Drive service that can be used to access Drive assets.
func service() (*drive.Service, error) {
	ctx := context.Background()

	b, err := ioutil.ReadFile("credentials.json")
	if err != nil {
		return nil, fmt.Errorf("Unable to read client secret file: %v", err)
	}

	// If you modify these scopes, delete the old token.json.
	config, err := google.ConfigFromJSON(b, drive.DriveScope)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse client secret file to config: %v", err)
	}

	client, err := getClient(config)
	if err != nil {
		return nil, err
	}

	srv, err := drive.NewService(ctx, option.WithHTTPClient(client))

	return srv, err
}

// CreateSheet uploads a CSV file as a Google Sheet in Google Drive.
func CreateSheet(name string, parentID string) (*drive.File, error) {
	srv, err := service()
	if err != nil {
		return nil, err
	}

	content, err := os.Open(name)
	defer content.Close()

	f := &drive.File{
		MimeType: "application/vnd.google-apps.spreadsheet",
		Name:     name,
		Parents:  []string{parentID},
	}

	file, err := srv.Files.Create(f).Media(content, googleapi.ContentType("text/csv")).Do()
	if err != nil {
		return nil, fmt.Errorf("Could not create file: " + err.Error())
	}

	return file, nil
}
