package partners

import (
	"fmt"

	"github.com/jibe0123/mysteryfactory/internal/models"
)

// PartnerClient defines common methods for content partners.
type PartnerClient interface {
	Authenticate(workspace *models.Workspace) error
	Upload(video *models.Video) (string, error)
	Publish(video *models.Video, workspace *models.Workspace) error
}

// New returns a client implementation for the given platform.
func New(platform string) (PartnerClient, error) {
	switch platform {
	case string(models.PlatformYouTube):
		return &youtubeClient{}, nil
	case string(models.PlatformTikTok):
		return &tiktokClient{}, nil
	default:
		return nil, fmt.Errorf("unsupported platform: %s", platform)
	}
}

// --- Example partner clients ---

type youtubeClient struct{}

func (c *youtubeClient) Authenticate(ws *models.Workspace) error {
	// TODO: implement OAuth flow for YouTube
	return nil
}

func (c *youtubeClient) Upload(video *models.Video) (string, error) {
	// TODO: upload video file using YouTube Data API
	return "", nil
}

func (c *youtubeClient) Publish(video *models.Video, ws *models.Workspace) error {
	// TODO: publish the uploaded video using YouTube Data API
	return nil
}

type tiktokClient struct{}

func (c *tiktokClient) Authenticate(ws *models.Workspace) error {
	// TODO: implement OAuth flow for TikTok
	return nil
}

func (c *tiktokClient) Upload(video *models.Video) (string, error) {
	// TODO: upload video using TikTok API
	return "", nil
}

func (c *tiktokClient) Publish(video *models.Video, ws *models.Workspace) error {
	// TODO: publish the uploaded video on TikTok
	return nil
}
