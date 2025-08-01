package partners

import (
	"testing"

	"github.com/jibe0123/mysteryfactory/internal/models"
	pkgpartners "github.com/jibe0123/mysteryfactory/pkg/partners"
)

type mockClient struct {
	calls []string
}

func (m *mockClient) Authenticate(*models.Workspace) error {
	m.calls = append(m.calls, "auth")
	return nil
}
func (m *mockClient) Upload(*models.Video) (string, error) {
	m.calls = append(m.calls, "upload")
	return "42", nil
}
func (m *mockClient) Publish(*models.Video, *models.Workspace) error {
	m.calls = append(m.calls, "publish")
	return nil
}
func (m *mockClient) FetchStats(*models.Video) (*models.VideoStats, error) {
	m.calls = append(m.calls, "stats")
	return &models.VideoStats{Views: 1}, nil
}

func TestServicePublishVideo(t *testing.T) {
	mc := &mockClient{}
	svc := NewService(func(string) (pkgpartners.Client, error) { return mc, nil })
	ws := &models.Workspace{}
	v := &models.Video{}
	stats, err := svc.PublishVideo(ws, v, models.PlatformYouTube)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats == nil || stats.Views != 1 {
		t.Fatalf("unexpected stats: %#v", stats)
	}
	expected := []string{"auth", "upload", "publish", "stats"}
	for i, call := range expected {
		if mc.calls[i] != call {
			t.Fatalf("expected call %s at index %d, got %s", call, i, mc.calls[i])
		}
	}
	if v.YouTubeID != "42" {
		t.Fatalf("video id not set")
	}
}

func TestServiceSyncStats(t *testing.T) {
	mc := &mockClient{}
	svc := NewService(func(string) (pkgpartners.Client, error) { return mc, nil })
	ws := &models.Workspace{}
	v := &models.Video{}
	stats, err := svc.SyncStats(ws, v, models.PlatformTikTok)
	if err != nil || stats.Views != 1 {
		t.Fatalf("unexpected result: %v %v", err, stats)
	}
	expected := []string{"auth", "stats"}
	for i, call := range expected {
		if mc.calls[i] != call {
			t.Fatalf("expected %s, got %s", call, mc.calls[i])
		}
	}
}
