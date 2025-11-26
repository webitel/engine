package app

import (
	"context"
	"errors"
	"time"

	"github.com/gofrs/uuid"
	memcache "github.com/jellydator/ttlcache/v3"
	"github.com/webitel/engine/model"
)

const (
	defaultMeetingTTL = 60 * time.Minute
	maxMeetingTTL     = 3 * time.Hour
)

type MeetingStorage interface {
	Create(ctx context.Context, domainId int64, meeting *model.Meeting) error
	Get(ctx context.Context, id string) (*model.Meeting, error)
}

type IDGenerator interface {
	Generate() string
}

type MeetingHandler struct {
	storage     MeetingStorage
	idGenerator IDGenerator
}

func NewMeetingHandler(storage MeetingStorage, generator IDGenerator) (*MeetingHandler, error) {
	if storage == nil {
		storage = NewMeetingMemoryStorage()
	}
	if generator == nil {
		generator = UUIDGenerator{}
	}
	return &MeetingHandler{storage: storage, idGenerator: generator}, nil
}

func (meetingHandler *MeetingHandler) CreateMeeting(ctx context.Context, domainId int64, request *model.Meeting) (*model.Meeting, error) {

	if domainId <= 0 {
		return nil, errors.New("domainId is required to appoint a meeting")
	}
	var (
		meeting = &model.Meeting{
			ID:        meetingHandler.idGenerator.Generate(),
			Title:     request.Title,
			TTL:       request.TTL,
			CreatedAt: request.CreatedAt,
			ExpiresAt: request.ExpiresAt,
			Metadata:  request.Metadata,
			DomainId:  domainId,
		}
	)
	// Normalize fields before saving
	if meeting.CreatedAt == 0 {
		meeting.CreatedAt = time.Now().Unix()
	}
	if meeting.TTL <= 0 {
		meeting.TTL = defaultMeetingTTL
	}
	meeting.ExpiresAt = time.UnixMilli(meeting.CreatedAt).Add(meeting.TTL * time.Second).UnixMilli()

	err := meetingHandler.storage.Create(ctx, domainId, meeting)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (meetingHandler *MeetingHandler) GetMeeting(ctx context.Context, id string) (*model.Meeting, error) {

	if len(id) == 0 {
		return nil, errors.New("meeting id is required to get a meeting")
	}

	meeting, err := meetingHandler.storage.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	return meeting, nil
}

func (meetingHandler *MeetingHandler) SearchMeetings(ctx context.Context, domainId int64, search *model.Searcher) ([]*model.Meeting, error) {
	// Not implemented yet
	return nil, nil
}

type MeetingMemoryStorage struct {
	cache *memcache.Cache[string, *model.Meeting]
}

func (m *MeetingMemoryStorage) Create(_ context.Context, _ int64, meeting *model.Meeting) error {
	if meeting.TTL <= 0 {
		return errors.New("meeting ttl is required")
	}
	if meeting.ID == "" {
		return errors.New("meeting id is required")
	}
	m.cache.Set(meeting.ID, meeting, meeting.TTL)
	return nil
}

func (m *MeetingMemoryStorage) Get(_ context.Context, id string) (*model.Meeting, error) {
	if id == "" {
		return nil, errors.New("meeting id is required")
	}
	meeting := m.cache.Get(id)
	if meeting == nil {
		return nil, errors.New("meeting not found")
	}
	return meeting.Value(), nil
}

func NewMeetingMemoryStorage() *MeetingMemoryStorage {
	cache := memcache.New[string, *model.Meeting]()
	return &MeetingMemoryStorage{
		cache: cache,
	}
}

type UUIDGenerator struct{}

func (g UUIDGenerator) Generate() string {
	id, _ := uuid.NewV4()
	return id.String()
}
