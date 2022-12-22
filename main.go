package gmail_reader

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

var srv *gmail.Service

const tmplt = "Here is the verification code for your recent login attempt"

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline /*AccessTypeOnline*/)
	fmt.Printf("Go to the following link in your browser then type the authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

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
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func InitReader() error {
	ctx := context.Background()
	b, err := os.ReadFile("credentials.json")
	if err != nil {
		err = fmt.Errorf("Unable to read client secret file: %v", err)
		return err
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, gmail.GmailReadonlyScope /* MailGoogleComScope */)
	if err != nil {
		err = fmt.Errorf("Unable to parse client secret file to config: %v", err)
		return err
	}
	client := getClient(config)

	srv, err = gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		err = fmt.Errorf("Unable to retrieve Gmail client: %v", err)
		return err
	}

	return nil
}

func GetCode() (string, error) {
	var err error
	var code string

	msgs, err := srv.Users.Messages.List("me").Do()
	if err != nil {
		return code, err
	}

	for _, v := range msgs.Messages {

		msg, err := srv.Users.Messages.Get("me", v.Id).Format("raw").Do()
		if err != nil {
			log.Printf("srv.Users.Messages.Get error: %+s", err)
			continue
		}

		if !strings.Contains(msg.Snippet, tmplt) {
			continue
		}

		parts := strings.Split(msg.Snippet, ":")
		code = strings.Split(parts[1], " ")[1]
		// log.Printf("verification code: %s\n", code)

		break
	}

	return code, err
}

/* func main() {
	_ = InitReader()
	_, _ = GetCode()
} */

// var msg *gmail.Message
// decodedData, _ := base64.URLEncoding.DecodeString(gmailMessageResposne.Raw)
// base64.URLEncoding(decodedData, decodedData)
// ioutile.WriteFile("message.eml", decodedData, os.ModePerm)

// Subject: The Appliance Repair Men - Login Verification Code
// From: notifications@theappliancerepairmen.com

/* 		msg, err := srv.Users.Messages.Get("me", v.Id).Do()
   		if err != nil {
   			log.Printf("err: %+s", err)
   			continue
   		}

   		for _, h := range msg.Payload.Headers {
   			if h.Name != "Subject" {
   				continue
   			}

   			if h.Value != "The Appliance Repair Men - Login Verification Code" {
   				continue
   			}

   			log.Println(h.Name, h.Value, v.Id, v.Snippet)

   			decodedData, err := base64.URLEncoding.DecodeString(msg.Payload.Body.Data)
   			if err != nil {
   				log.Printf("err: %+s", err)
   				continue
   			}

   			log.Printf("decodedData: %v\n", decodedData)
   			log.Println("---------------------------------------------------------------------------------------------------")
   		}
*/

/* Create a New Project.
Go to APIs and Services.
Go to Enable APIs and Services.
Enable Gmail API.
Configure Consent screen.
Enter Application name.
Go to Credentials.
Create an OAuth Client ID.
Get in from URL*/
