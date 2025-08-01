package partners

import (
	"fmt"
	"github.com/jibe0123/mysteryfactory/internal/models"
)

// Client defines platform client capabilities.
type Client interface {
	Authenticate(*models.Workspace) error
	Upload(*models.Video) (string, error)
	Publish(*models.Video, *models.Workspace) error
	// FetchStats retrieves statistics for the provided video from the platform.
	FetchStats(*models.Video) (*models.VideoStats, error)
}

// Factory creates a new client for the specified platform.
func New(platform string) (Client, error) {
	switch models.Platform(platform) {
	case models.PlatformYouTube:
		return &youtubeClient{}, nil
	case models.PlatformTikTok:
		return &tiktokClient{}, nil
	case models.PlatformInstagram:
		return &instagramClient{}, nil
	case models.PlatformFacebook:
		return &facebookClient{}, nil
	case models.PlatformTwitter:
		return &twitterClient{}, nil
	case models.PlatformSnapchat:
		return &snapchatClient{}, nil
	default:
		return nil, fmt.Errorf("unsupported platform: %s", platform)
	}
}
