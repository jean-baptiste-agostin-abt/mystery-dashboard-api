package partners

import (
	"fmt"

	"github.com/HiWay-Media/tiktok-go-sdk/tiktok"
	"github.com/jibe0123/mysteryfactory/internal/models"
	"os"
)

type tiktokClient struct {
	sdk tiktok.ITiktok
}

func (c *tiktokClient) Authenticate(ws *models.Workspace) error {
	appID := os.Getenv("TIKTOK_APP_ID")
	secret := os.Getenv("TIKTOK_APP_SECRET")
	client, err := tiktok.NewTikTok(appID, secret, false)
	if err != nil {
		return fmt.Errorf("failed to create TikTok client: %w", err)
	}
	authURL := client.CodeAuthUrl()
	fmt.Printf("Visit this URL to authorize: %s\n", authURL)
	token := ws.OAuthCode
	client.SetAccessToken(token)
	c.sdk = client
	return nil
}

func (c *tiktokClient) Upload(video *models.Video) (string, error) {
	resp, err := c.sdk.PostVideoInit(video.Title, video.Description, video.FileURL, "PUBLIC", false, false, false)
	if err != nil {
		return "", err
	}
	return resp.Data.PubblishId, nil
}

func (c *tiktokClient) Publish(video *models.Video, ws *models.Workspace) error {
	_, err := c.sdk.PublishVideo(video.TikTokID)
	return err
}

func (c *tiktokClient) FetchStats(video *models.Video) (*models.VideoStats, error) {
	return nil, fmt.Errorf("fetch stats not implemented")
}
