package partners

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/jibe0123/mysteryfactory/internal/models"
)

type snapchatClient struct {
	httpClient  *http.Client
	accessToken string
	profileId   string
}

func (c *snapchatClient) Authenticate(ws *models.Workspace) error {
	if ws.SnapchatAccessToken == "" || ws.SnapchatProfileID == "" {
		return fmt.Errorf("Snapchat credentials missing in workspace")
	}
	c.accessToken = ws.SnapchatAccessToken
	c.profileId = ws.SnapchatProfileID
	c.httpClient = http.DefaultClient
	return nil
}

func (c *snapchatClient) Upload(video *models.Video) (string, error) {
	buf := &bytes.Buffer{}
	writer := multipart.NewWriter(buf)
	file, err := os.Open(video.FilePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	part, err := writer.CreateFormFile("media", filepath.Base(video.FilePath))
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(part, file); err != nil {
		return "", err
	}
	writer.Close()

	url := fmt.Sprintf("https://businessapi.snapchat.com/v1/public_profiles/%s/media", c.profileId)
	req, err := http.NewRequestWithContext(context.Background(), "POST", url, buf)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("upload failed: %s", body)
	}

	var out struct {
		MediaID string `json:"media_id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	return out.MediaID, nil
}

func (c *snapchatClient) Publish(video *models.Video, ws *models.Workspace) error {
	payload := map[string]interface{}{
		"media_id": video.SnapchatMediaID,
		"caption":  video.Description,
	}
	body, _ := json.Marshal(payload)

	url := fmt.Sprintf("https://businessapi.snapchat.com/v1/public_profiles/%s/stories", c.profileId)
	req, err := http.NewRequestWithContext(context.Background(), "POST", url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("publish failed: %s", respBody)
	}
	return nil
}

func (c *snapchatClient) FetchStats(video *models.Video) (*models.VideoStats, error) {
	return nil, fmt.Errorf("fetch stats not implemented")
}
