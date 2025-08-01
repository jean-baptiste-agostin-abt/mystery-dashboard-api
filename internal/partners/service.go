package partners

import (
	"strconv"

	"github.com/jibe0123/mysteryfactory/internal/models"
	pkgpartners "github.com/jibe0123/mysteryfactory/pkg/partners"
)

// Service handles business logic around partner platforms.
type Service struct {
	factory func(platform string) (pkgpartners.Client, error)
}

// NewService creates a new Service.
func NewService(factory func(string) (pkgpartners.Client, error)) *Service {
	return &Service{factory: factory}
}

// PublishVideo uploads then publishes a video to the specified platform.
func (s *Service) PublishVideo(ws *models.Workspace, v *models.Video, platform models.Platform) (*models.VideoStats, error) {
	client, err := s.factory(string(platform))
	if err != nil {
		return nil, err
	}
	if err := client.Authenticate(ws); err != nil {
		return nil, err
	}
	id, err := client.Upload(v)
	if err != nil {
		return nil, err
	}
	switch platform {
	case models.PlatformYouTube:
		v.YouTubeID = id
	case models.PlatformTikTok:
		v.TikTokID = id
	case models.PlatformInstagram:
		v.InstagramID = id
	case models.PlatformFacebook:
		v.FacebookID = id
	case models.PlatformTwitter:
		if val, convErr := strconv.ParseInt(id, 10, 64); convErr == nil {
			v.TwitterMediaID = val
		}
	case models.PlatformSnapchat:
		v.SnapchatMediaID = id
	}
	if err := client.Publish(v, ws); err != nil {
		return nil, err
	}
	return client.FetchStats(v)
}

// SyncStats retrieves latest statistics from the platform.
func (s *Service) SyncStats(ws *models.Workspace, v *models.Video, platform models.Platform) (*models.VideoStats, error) {
	client, err := s.factory(string(platform))
	if err != nil {
		return nil, err
	}
	if err := client.Authenticate(ws); err != nil {
		return nil, err
	}
	return client.FetchStats(v)
}
