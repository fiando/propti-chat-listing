package services

import (
	"context"
	"errors"
	"time"

	"github.com/fiando/propti/backend/internal/models"
)

type fakeMediaService struct {
	presignedURLs map[string]string
	signedURLs    map[string]string
	publicURLs    map[string]string
	heads         map[string]*MediaObjectHead
	bytesByKey    map[string][]byte
	copies        []mediaCopyCall
	deletedKeys   []string
}

type mediaCopyCall struct {
	sourceKey      string
	destinationKey string
}

func (f *fakeMediaService) GetPresignedUploadURL(ctx context.Context, key, contentType string) (string, error) {
	if f.presignedURLs == nil {
		return "", nil
	}
	if url, ok := f.presignedURLs[key]; ok {
		return url, nil
	}
	return "", nil
}

func (f *fakeMediaService) GetSignedDownloadURL(ctx context.Context, key string) (string, error) {
	if f.signedURLs == nil {
		return "", nil
	}
	if url, ok := f.signedURLs[key]; ok {
		return url, nil
	}
	return "", nil
}

func (f *fakeMediaService) BuildPublicURL(key string) string {
	if f.publicURLs == nil {
		return ""
	}
	return f.publicURLs[key]
}

func (f *fakeMediaService) HeadObject(ctx context.Context, key string) (*MediaObjectHead, error) {
	if f.heads == nil {
		return nil, errors.New("head object not configured")
	}
	head, ok := f.heads[key]
	if !ok {
		return nil, errors.New("object not found")
	}
	copy := *head
	return &copy, nil
}

func (f *fakeMediaService) CopyObject(ctx context.Context, sourceKey, destinationKey string) error {
	f.copies = append(f.copies, mediaCopyCall{
		sourceKey:      sourceKey,
		destinationKey: destinationKey,
	})
	return nil
}

func (f *fakeMediaService) DeleteObject(ctx context.Context, key string) error {
	f.deletedKeys = append(f.deletedKeys, key)
	return nil
}

func (f *fakeMediaService) GetObjectBytes(ctx context.Context, key string) ([]byte, error) {
	if f.bytesByKey == nil {
		return nil, errors.New("object bytes not configured")
	}
	body, ok := f.bytesByKey[key]
	if !ok {
		return nil, errors.New("object not found")
	}
	return append([]byte(nil), body...), nil
}

type fakeUploadSessionStore struct {
	sessions     map[string]*models.UploadSession
	putCalls     []*models.UploadSession
	consumeCalls []string
}

func (f *fakeUploadSessionStore) Put(ctx context.Context, session *models.UploadSession) error {
	if f.sessions == nil {
		f.sessions = make(map[string]*models.UploadSession)
	}
	copy := *session
	f.sessions[session.SessionID] = &copy
	f.putCalls = append(f.putCalls, &copy)
	return nil
}

func (f *fakeUploadSessionStore) GetBySessionID(ctx context.Context, sessionID string) (*models.UploadSession, error) {
	if f.sessions == nil {
		return nil, nil
	}
	session := f.sessions[sessionID]
	if session == nil {
		return nil, nil
	}
	copy := *session
	return &copy, nil
}

func (f *fakeUploadSessionStore) Consume(ctx context.Context, sessionID, listingID string, consumedAt time.Time) error {
	f.consumeCalls = append(f.consumeCalls, sessionID)
	session := f.sessions[sessionID]
	if session == nil {
		return nil
	}
	session.ListingID = listingID
	session.ConsumedAt = &consumedAt
	return nil
}
