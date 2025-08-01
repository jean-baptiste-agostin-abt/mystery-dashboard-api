package partners

import (
	"context"
	"fmt"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/jibe0123/mysteryfactory/internal/models"
)

type twitterClient struct {
	client *twitter.Client
}

func (c *twitterClient) Authenticate(ws *models.Workspace) error {
	cfg := oauth1.NewConfig(ws.TwitterConsumerKey, ws.TwitterConsumerSecret)
	token := oauth1.NewToken(ws.TwitterAccessToken, ws.TwitterAccessSecret)
	httpClient := cfg.Client(context.Background(), token)
	c.client = twitter.NewClient(httpClient)
	return nil
}

func (c *twitterClient) Upload(video *models.Video) (string, error) {
	// Media upload not implemented
	return "", nil
}

func (c *twitterClient) Publish(video *models.Video, ws *models.Workspace) error {
	_, _, err := c.client.Statuses.Update(video.Description, nil)
	return err
}

func (c *twitterClient) FetchStats(video *models.Video) (*models.VideoStats, error) {
	return nil, fmt.Errorf("fetch stats not implemented")
}
