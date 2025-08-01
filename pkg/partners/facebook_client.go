package partners

import (
	"fmt"

	"github.com/huandu/facebook/v2"
	"github.com/jibe0123/mysteryfactory/internal/models"
	"os"
)

type facebookClient struct {
	session *facebook.Session
}

func (c *facebookClient) Authenticate(ws *models.Workspace) error {
	appID := os.Getenv("FB_APP_ID")
	appSecret := os.Getenv("FB_APP_SECRET")
	token := ws.FacebookPageToken
	if appID == "" || appSecret == "" || token == "" {
		return fmt.Errorf("Facebook credentials missing")
	}
	app := facebook.New(appID, appSecret)
	session := app.Session(token)
	c.session = session
	return nil
}

func (c *facebookClient) Upload(video *models.Video) (string, error) {
	res, err := c.session.Post("/me/videos", facebook.Params{
		"file_url":    video.FileURL,
		"description": video.Description,
	})
	if err != nil {
		return "", err
	}
	id, _ := res.Get("id").(string)
	return id, nil
}

func (c *facebookClient) Publish(video *models.Video, ws *models.Workspace) error {
	_, err := c.session.Post(fmt.Sprintf("/%s/feed", ws.FacebookPageID), facebook.Params{
		"message": fmt.Sprintf("Nouvelle vidéo : https://facebook.com/%s/videos/%s", ws.FacebookPageID, video.FacebookID),
	})
	return err
}

func (c *facebookClient) FetchStats(video *models.Video) (*models.VideoStats, error) {
	return nil, fmt.Errorf("fetch stats not implemented")
}
