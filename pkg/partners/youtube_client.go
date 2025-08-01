package partners

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/youtube/v3"

	"github.com/jibe0123/mysteryfactory/internal/models"
)

func getYouTubeClient(ctx context.Context, credsPath, tokenPath string) (*http.Client, error) {
	b, err := os.ReadFile(credsPath)
	if err != nil {
		return nil, fmt.Errorf("unable to read credentials: %w", err)
	}
	config, err := google.ConfigFromJSON(b, youtube.YoutubeUploadScope, youtube.YoutubeForceSslScope)
	if err != nil {
		return nil, fmt.Errorf("unable to parse credentials: %w", err)
	}

	tokFile := filepath.Join(tokenPath, "youtube_token.json")
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok, err = getTokenFromWeb(ctx, config)
		if err != nil {
			return nil, err
		}
		if err := saveToken(tokFile, tok); err != nil {
			return nil, err
		}
	}
	return config.Client(ctx, tok), nil
}

func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	return tok, json.NewDecoder(f).Decode(tok)
}

func saveToken(path string, tok *oauth2.Token) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(tok)
}

func getTokenFromWeb(ctx context.Context, config *oauth2.Config) (*oauth2.Token, error) {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following URL to authorize:\n%v\n", authURL)
	var code string
	fmt.Print("Enter the code you received here: ")
	if _, err := fmt.Scan(&code); err != nil {
		return nil, err
	}
	return config.Exchange(ctx, code)
}

type youtubeClient struct {
	service *youtube.Service
}

func (c *youtubeClient) Authenticate(ws *models.Workspace) error {
	ctx := context.Background()
	client, err := getYouTubeClient(ctx, ws.CredentialsPath, ws.TokenDir)
	if err != nil {
		return err
	}
	srv, err := youtube.New(client)
	if err != nil {
		return fmt.Errorf("youtube service init: %w", err)
	}
	c.service = srv
	return nil
}

func (c *youtubeClient) Upload(video *models.Video) (string, error) {
	call := c.service.Videos.Insert([]string{"snippet", "status"}, &youtube.Video{
		Snippet: &youtube.VideoSnippet{
			Title:       video.Title,
			Description: video.Description,
			Tags:        video.GetTags(),
		},
		Status: &youtube.VideoStatus{PrivacyStatus: "private"},
	})
	file, err := os.Open(video.FilePath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	res, err := call.Media(file).Do()
	if err != nil {
		return "", err
	}
	return res.Id, nil
}

func (c *youtubeClient) Publish(video *models.Video, ws *models.Workspace) error {
	_, err := c.service.Videos.Update([]string{"status"}, &youtube.Video{
		Id:     video.YouTubeID,
		Status: &youtube.VideoStatus{PrivacyStatus: "public"},
	}).Do()
	return err
}

func (c *youtubeClient) FetchStats(video *models.Video) (*models.VideoStats, error) {
	// Stats retrieval not implemented in this example
	return nil, fmt.Errorf("fetch stats not implemented")
}
