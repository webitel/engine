package grpc_api

import (
	"context"
	"errors"
	pb "github.com/webitel/engine/gen/engine"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/pkg/wbt/auth_manager"
	"time"
)

var (
	_ pb.MeetingServiceServer = &MeetingApi{}
)

type MeetingHandler interface {
	CreateMeeting(ctx context.Context, domainId int64, request *model.Meeting) (*model.Meeting, error)
	GetMeeting(ctx context.Context, id string) (*model.Meeting, error)
	SearchMeetings(ctx context.Context, domainId int64, search *model.Searcher) ([]*model.Meeting, error)
}

type Authorizer interface {
	GetSessionFromCtx(ctx context.Context) (*auth_manager.Session, model.AppError)
}

type MeetingApi struct {
	auth    Authorizer
	handler MeetingHandler
	pb.UnimplementedMeetingServiceServer
}

func NewMeetingApi(app Authorizer, handler MeetingHandler) *MeetingApi {
	return &MeetingApi{auth: app, handler: handler}
}

func (svc *MeetingApi) CreateMeeting(ctx context.Context, req *pb.CreateMeetingRequest) (*pb.Meeting, error) {
	session, extErr := svc.auth.GetSessionFromCtx(ctx)
	if extErr != nil {
		return nil, extErr
	}
	if !session.HasLicense("CALL_CENTER") {
		return nil, errors.New("call center license is required")
	}
	meeting := &model.Meeting{
		TTL:       time.Duration(req.Ttl) * time.Second,
		CreatedAt: time.Now().Unix(),
		Metadata:  req.Metadata,
	}

	resultingMeeting, err := svc.handler.CreateMeeting(ctx, session.DomainId, meeting)
	if err != nil {
		return nil, model.NewInternalError("meeting.api.create_meeting.app_error", err.Error())
	}

	return &pb.Meeting{
		Id:        resultingMeeting.ID,
		Title:     resultingMeeting.Title,
		Ttl:       int32(resultingMeeting.TTL.Seconds()),
		CreatedAt: resultingMeeting.CreatedAt,
		ExpiresAt: resultingMeeting.ExpiresAt,
		Metadata:  resultingMeeting.Metadata,
	}, nil
}

func (svc *MeetingApi) GetMeeting(ctx context.Context, req *pb.GetMeetingRequest) (*pb.Meeting, error) {
	session, extErr := svc.auth.GetSessionFromCtx(ctx)
	if extErr != nil {
		return nil, extErr
	}
	if !session.HasLicense("CALL_CENTER") {
		return nil, errors.New("call center license is required")
	}

	meeting, err := svc.handler.GetMeeting(ctx, req.MeetingId)
	if err != nil {
		return nil, model.NewInternalError("meeting.api.get_meeting.app_error", err.Error())
	}
	if meeting.DomainId != session.DomainId {
		return nil, model.NewNotFoundError("meeting.api.get_meeting.not_found", "not found")
	}

	return &pb.Meeting{
		Id:        meeting.ID,
		Title:     meeting.Title,
		Ttl:       int32(meeting.TTL.Seconds()),
		CreatedAt: meeting.CreatedAt,
		ExpiresAt: meeting.ExpiresAt,
		Metadata:  meeting.Metadata,
	}, nil
}

func (svc *MeetingApi) GetMeetingNA(ctx context.Context, req *pb.GetMeetingRequest) (*pb.Meeting, error) {
	meeting, err := svc.handler.GetMeeting(ctx, req.MeetingId)
	if err != nil {
		return nil, model.NewInternalError("meeting.api.get_meeting.app_error", err.Error())
	}

	return &pb.Meeting{
		Id:        meeting.ID,
		Title:     meeting.Title,
		Ttl:       int32(meeting.TTL.Seconds()),
		CreatedAt: meeting.CreatedAt,
		ExpiresAt: meeting.ExpiresAt,
		Metadata:  meeting.Metadata,
	}, nil
}
