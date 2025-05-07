package call_manager

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/webitel/engine/gen/fs"
	"github.com/webitel/engine/model"
	"go.uber.org/ratelimit"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

const (
	FS_CONNECTION_TIMEOUT = 2 * time.Second
)

var patternSps = regexp.MustCompile(`\D+`)
var patternVersion = regexp.MustCompile(`^.*?\s(\d+[\.\S]+[^\s]).*`)

type CallConnection struct {
	proxy       string
	name        string
	host        string
	port        int
	rateLimiter ratelimit.Limiter
	client      *grpc.ClientConn
	api         fs.ApiClient
}

func NewCallConnection(name, host, proxy string, port int) (CallClient, model.AppError) {
	var err error
	c := &CallConnection{
		proxy: proxy,
		name:  name,
		host:  host,
		port:  port,
	}

	c.client, err = grpc.Dial(fmt.Sprintf("%s:%d", c.host, c.port), grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(FS_CONNECTION_TIMEOUT))

	if err != nil {
		return nil, model.NewInternalError("grpc.create_connection.app_error", err.Error())
	}

	c.api = fs.NewApiClient(c.client)
	return c, nil
}

func (c *CallConnection) MakeOutboundCall(req *model.CallRequest) (string, model.AppError) {
	if req.Variables == nil {
		req.Variables = make(map[string]string)
	}

	req.Variables["sip_route_uri"] = c.proxy
	//DUMP(req)
	uuid, cause, err := c.NewCall(req)
	if err != nil {
		return "", err
	}

	if cause != "" {
		//FIXME
	}

	return uuid, nil
}

func (c *CallConnection) Ready() bool {
	switch c.client.GetState() {
	case connectivity.Idle, connectivity.Ready:
		return true
	}
	return false
}

func (c *CallConnection) Close() error {
	err := c.client.Close()
	if err != nil {
		return model.NewInternalError("grpc.close_connection.app_error", err.Error())
	}

	return nil
}

func (c *CallConnection) Name() string {
	return c.name
}

func (c *CallConnection) Host() string {
	return c.host
}

func (c *CallConnection) GetServerVersion() (string, model.AppError) {
	res, err := c.api.Execute(context.Background(), &fs.ExecuteRequest{
		Command: "version",
	})

	if err != nil {
		return "", model.NewInternalError("external.get_server_version.app_error", err.Error())
	}

	return patternVersion.ReplaceAllString(strings.TrimSpace(res.Data), "$1"), nil
}

func (c *CallConnection) SetConnectionSps(sps int) (int, model.AppError) {
	if sps > 0 {
		c.rateLimiter = ratelimit.New(sps)
	}
	return sps, nil
}

func (c *CallConnection) GetRemoteSps() (int, model.AppError) {
	res, err := c.api.Execute(context.Background(), &fs.ExecuteRequest{
		Command: "fsctl",
		Args:    "sps",
	})

	if err != nil {
		return 0, model.NewInternalError("external.get_sps.app_error", err.Error())
	}

	return parseSps(res.String()), nil
}

func (c *CallConnection) NewCallContext(ctx context.Context, settings *model.CallRequest) (string, string, model.AppError) {
	request := &fs.OriginateRequest{
		Endpoints:    settings.Endpoints,
		Destination:  settings.Destination,
		CallerNumber: settings.CallerNumber,
		CallerName:   settings.CallerName,
		Timeout:      int32(settings.Timeout),
		Context:      settings.Context,
		Dialplan:     settings.Dialplan,
		Variables:    settings.Variables,
	}

	if len(settings.Applications) > 0 {
		request.Extensions = []*fs.OriginateRequest_Extension{}

		for _, v := range settings.Applications {
			request.Extensions = append(request.Extensions, &fs.OriginateRequest_Extension{
				AppName: v.AppName,
				Args:    v.Args,
			})
		}
	}

	switch settings.Strategy {
	case model.CALL_STRATEGY_FAILOVER:
		request.Strategy = fs.OriginateRequest_FAILOVER
		break
	case model.CALL_STRATEGY_MULTIPLE:
		request.Strategy = fs.OriginateRequest_MULTIPLE
		break
	}

	if c.rateLimiter != nil {
		c.rateLimiter.Take()
	}

	response, err := c.api.Originate(ctx, request)

	if err != nil {
		return "", "", model.NewInternalError("external.new_call.app_error", err.Error())
	}

	if response.Error != nil {
		return "", response.Error.Message, model.NewInternalError("external.new_call.app_error", response.Error.String())
	}

	return response.Uuid, "", nil
}

func (c *CallConnection) NewCall(settings *model.CallRequest) (string, string, model.AppError) {
	DUMP(settings)
	return c.NewCallContext(context.Background(), settings)
}

func (c *CallConnection) HangupCall(id, cause string) model.AppError {
	res, err := c.api.Hangup(context.Background(), &fs.HangupRequest{
		Uuid:  id,
		Cause: cause,
	})

	if err != nil {
		return model.NewInternalError("external.hangup_call.app_error", err.Error())
	}

	if res.Error != nil {
		//todo
		if res.Error.Message == "No such channel!" {

			return NotFoundCall
		}

		return model.NewInternalError("external.hangup_call.app_error", res.Error.String())
	}
	return nil
}

func (c *CallConnection) ConfirmPushCall(id string) model.AppError {
	res, err := c.api.ConfirmPush(context.Background(), &fs.ConfirmPushRequest{
		Id: id,
	})

	if err != nil {
		return model.NewInternalError("external.push_call.app_error", err.Error())
	}

	if res.Error != nil {
		//todo
		if res.Error.Message == "No such channel!" {

			return NotFoundCall
		}

		return model.NewInternalError("external.push_call.app_error", res.Error.String())
	}
	return nil
}

func (c *CallConnection) SetCallVariables(id string, variables map[string]string) model.AppError {

	res, err := c.api.SetVariables(context.Background(), &fs.SetVariablesRequest{
		Uuid:      id,
		Variables: variables,
	})

	if err != nil {
		return model.NewInternalError("external.set_call_variables.app_error", err.Error())
	}

	if res.Error != nil {
		return model.NewInternalError("external.set_call_variables.app_error", res.Error.String())
	}

	return nil
}

func (c *CallConnection) Hold(id string) model.AppError {
	_, err := c.api.Hold(context.Background(), &fs.HoldRequest{
		Id: []string{id},
	})

	if err != nil {
		return model.NewInternalError("external.hold_call.app_error", err.Error())
	}

	return nil
}

func (c *CallConnection) UnHold(id string) model.AppError {
	_, err := c.api.UnHold(context.Background(), &fs.UnHoldRequest{
		Id: []string{id},
	})

	if err != nil {
		return model.NewInternalError("external.un_hold_call.app_error", err.Error())
	}

	return nil
}

func (c *CallConnection) BridgeCall(legAId, legBId string, vars map[string]string) (string, model.AppError) {
	response, err := c.api.BridgeCall(context.Background(), &fs.BridgeCallRequest{
		LegAId:    legAId,
		LegBId:    legBId,
		Variables: vars,
	})
	if err != nil {
		return "", model.NewInternalError("external.bridge_call.app_error", err.Error())
	}

	if response.Error != nil {
		return "", model.NewInternalError("external.bridge_call.app_error", response.Error.String())
	}

	return response.Uuid, nil
}

func (c *CallConnection) DTMF(id string, ch rune) model.AppError {
	_, err := c.api.Execute(context.Background(), &fs.ExecuteRequest{
		Command: "uuid_recv_dtmf",
		Args:    fmt.Sprintf("%s %c", id, ch),
	})

	if err != nil {
		return model.NewInternalError("external.dtmf.app_error", err.Error())
	}
	return nil
}

func (c *CallConnection) SetEavesdropState(id string, state string) model.AppError {
	_, err := c.api.SetEavesdropState(context.Background(), &fs.SetEavesdropStateRequest{
		Id:    id,
		State: state,
	})

	if err != nil {
		return model.NewInternalError("external.eavesdrop.app_error", err.Error())
	}
	return nil
}

func (c *CallConnection) BlindTransfer(id, destination string) model.AppError {
	_, err := c.api.Execute(context.Background(), &fs.ExecuteRequest{
		Command: "uuid_transfer",
		Args:    fmt.Sprintf("%s %s", id, destination),
	})

	if err != nil {
		return model.NewInternalError("external.blind_transfer.app_error", err.Error())
	}
	return nil
}

func (c *CallConnection) BlindTransferExt(id, destination string, vars map[string]string) model.AppError {
	res, err := c.api.BlindTransfer(context.Background(), &fs.BlindTransferRequest{
		Id:          id,
		Destination: destination,
		Variables:   vars,
		Dialplan:    "",
		Context:     "",
	})

	if err != nil {
		return model.NewInternalError("external.blind_transfer_ext.app_error", err.Error())
	}

	if res != nil && res.Error != nil {
		return model.NewBadRequestError("external.blind_transfer_ext.valid", res.Error.Message)
	}

	return nil
}

// uuid_audio 8e345bfc-47b9-46c1-bdf0-3b874a8539c8 start read mute -1
// add eavesdrop mute other channel write
func (c *CallConnection) Mute(id string, val bool) model.AppError {
	var mute = 0
	if val {
		mute = -1
	}
	_, err := c.api.Execute(context.Background(), &fs.ExecuteRequest{
		Command: "uuid_audio",
		Args:    fmt.Sprintf("%s start read mute %d", id, mute),
	})

	if err != nil {
		return model.NewInternalError("external.mute.app_error", err.Error())
	}
	return nil
}

func (c *CallConnection) close() {
	c.client.Close()
}

func parseSps(str string) int {
	i, _ := strconv.Atoi(patternSps.ReplaceAllString(str, ""))
	return i
}

func (c *CallConnection) Execute(app string, args string) model.AppError {
	_, err := c.api.Execute(context.Background(), &fs.ExecuteRequest{
		Command: app,
		Args:    args,
	})

	if err != nil {
		return model.NewInternalError("external.blind_transfer.app_error", err.Error())
	}
	return nil
}
