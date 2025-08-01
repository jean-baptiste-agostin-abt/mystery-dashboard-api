package partners

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/jibe0123/mysteryfactory/internal/models"
)

type instagramClient struct {
	httpClient  *http.Client
	userID      string
	accessToken string
}

func (c *instagramClient) Authenticate(ws *models.Workspace) error {
	if ws.InstagramUserID == "" || ws.InstagramAccessToken == "" {
		return fmt.Errorf("Instagram credentials missing in workspace")
	}
	c.userID = ws.InstagramUserID
	c.accessToken = ws.InstagramAccessToken
	c.httpClient = http.DefaultClient
	return nil
}

func (c *instagramClient) Upload(video *models.Video) (string, error) {
	url := fmt.Sprintf("https://graph.facebook.com/v17.0/%s/media?image_url=%s&caption=%s&access_token=%s",
		c.userID, video.FileURL, video.Description, c.accessToken)
	resp, err := c.httpClient.Post(url, "application/json", nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var out struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	return out.ID, nil
}

func (c *instagramClient) Publish(video *models.Video, ws *models.Workspace) error {
	url := fmt.Sprintf("https://graph.facebook.com/v17.0/%s/media_publish?creation_id=%s&access_token=%s",
		c.userID, video.InstagramID, c.accessToken)
	resp, err := c.httpClient.Post(url, "application/json", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var out struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return err
	}
	return nil
}

func (c *instagramClient) FetchStats(video *models.Video) (*models.VideoStats, error) {
	return nil, fmt.Errorf("fetch stats not implemented")
}
