package grpc_api

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
	"google.golang.org/grpc/metadata"

	"github.com/webitel/engine/gen/engine"
	"github.com/webitel/engine/model"
	"github.com/webitel/engine/pkg/wbt/auth_manager"
)

const exportMembersPageSize = 5000

func (api *member) ExportMembers(in *engine.ExportMembersRequest, stream engine.MemberService_ExportMembersServer) error {
	ctx := stream.Context()

	session, err := api.app.GetSessionFromCtx(ctx)
	if err != nil {
		return err
	}

	permission := session.GetPermission(model.PERMISSION_SCOPE_CC_QUEUE)
	if !permission.CanRead() {
		return api.app.MakePermissionError(session, permission, auth_manager.PERMISSION_ACCESS_READ)
	}

	if session.UseRBAC(auth_manager.PERMISSION_ACCESS_READ, permission) {
		var perm bool
		if perm, err = api.app.QueueCheckAccess(ctx, session.Domain(0), int64(in.GetQueueId()), session.GetAclRoles(),
			auth_manager.PERMISSION_ACCESS_READ); err != nil {
			return err
		} else if !perm {
			return api.app.MakeResourcePermissionError(session, int64(in.GetQueueId()), permission, auth_manager.PERMISSION_ACCESS_READ)
		}
	}

	format := in.GetFormat()
	if format != "csv" && format != "xlsx" {
		return model.NewBadRequestError("grpc.member.export_members.format", "format must be 'csv' or 'xlsx'")
	}

	fields := in.GetFields()
	if len(fields) == 0 {
		fields = model.Member{}.DefaultFields()
	}

	filename := "members_" + time.Now().Format("2006-01-02_15-04-05") + "." + format
	if sendErr := stream.SendHeader(metadata.Pairs("filename", filename, "format", format)); sendErr != nil {
		return sendErr
	}

	domainId := session.Domain(0)

	if format == "csv" {
		return api.exportMembersCSV(ctx, domainId, in, fields, stream)
	}
	return api.exportMembersXLSX(ctx, domainId, in, fields, stream)
}

func buildExportMembersSearchRequest(in *engine.ExportMembersRequest, page int) *model.SearchMemberRequest {
	req := &model.SearchMemberRequest{
		ListRequest: model.ListRequest{
			Q:       in.GetQ(),
			Page:    page,
			PerPage: exportMembersPageSize,
			Fields:  in.GetFields(),
		},
		Ids:        in.GetId(),
		QueueId:    &in.QueueId,
		BucketIds:  in.GetBucketId(),
		StopCauses: in.GetStopCause(),
		AgentIds:   in.GetAgentId(),
		Variables:  in.GetVariables(),
	}

	if in.Destination != "" {
		req.Destination = &in.Destination
	}
	if in.Name != "" {
		req.Name = &in.Name
	}
	if in.GetPriority() != nil {
		req.Priority = &model.FilterBetween{From: in.GetPriority().GetFrom(), To: in.GetPriority().GetTo()}
	}
	if in.GetAttempts() != nil {
		req.Attempts = &model.FilterBetween{From: in.GetAttempts().GetFrom(), To: in.GetAttempts().GetTo()}
	}
	if in.GetCreatedAt() != nil {
		req.CreatedAt = &model.FilterBetween{From: in.GetCreatedAt().GetFrom(), To: in.GetCreatedAt().GetTo()}
	}
	if in.GetOfferingAt() != nil {
		req.OfferingAt = &model.FilterBetween{From: in.GetOfferingAt().GetFrom(), To: in.GetOfferingAt().GetTo()}
	}

	return req
}

func (api *member) exportMembersCSV(ctx context.Context, domainId int64, in *engine.ExportMembersRequest, fields []string,
	stream engine.MemberService_ExportMembersServer) error {

	page := 1
	sentAnyChunk := false

	for {
		list, endList, err := api.app.SearchMembers(ctx, domainId, buildExportMembersSearchRequest(in, page))
		if err != nil {
			return err
		}

		if len(list) == 0 {
			break
		}

		chunk, genErr := generateMembersCSVChunk(fields, membersToExportRows(list, fields), page, in.GetSeparator())
		if genErr != nil {
			return model.NewInternalError("grpc.member.export_members.csv", genErr.Error())
		}

		if len(chunk) > 0 {
			if sendErr := stream.Send(&engine.ExportMembersResponse{Data: chunk}); sendErr != nil {
				return sendErr
			}
			sentAnyChunk = true
		}

		if endList {
			break
		}
		page++
	}

	if !sentAnyChunk {
		chunk, genErr := generateMembersCSVChunk(fields, nil, 1, in.GetSeparator())
		if genErr != nil {
			return model.NewInternalError("grpc.member.export_members.csv", genErr.Error())
		}
		return stream.Send(&engine.ExportMembersResponse{Data: chunk})
	}

	return nil
}

func (api *member) exportMembersXLSX(ctx context.Context, domainId int64, in *engine.ExportMembersRequest, fields []string,
	stream engine.MemberService_ExportMembersServer) error {

	var allRows [][]string
	page := 1

	for {
		list, endList, err := api.app.SearchMembers(ctx, domainId, buildExportMembersSearchRequest(in, page))
		if err != nil {
			return err
		}

		if len(list) == 0 {
			break
		}

		allRows = append(allRows, membersToExportRows(list, fields)...)

		if endList {
			break
		}
		page++
	}

	data, err := generateMembersXLSX(fields, allRows)
	if err != nil {
		return model.NewInternalError("grpc.member.export_members.xlsx", err.Error())
	}

	const chunkSize = 1024 * 1024
	for i := 0; i < len(data); i += chunkSize {
		end := min(i+chunkSize, len(data))
		if err = stream.Send(&engine.ExportMembersResponse{Data: data[i:end]}); err != nil {
			return err
		}
	}

	return nil
}

func membersToExportRows(list []*model.Member, fields []string) [][]string {
	rows := make([][]string, 0, len(list))
	for _, m := range list {
		row := make([]string, len(fields))
		for i, f := range fields {
			row[i] = memberExportFieldValue(m, f)
		}
		rows = append(rows, row)
	}
	return rows
}

func memberExportFieldValue(m *model.Member, field string) string {
	switch field {
	case "id":
		return strconv.FormatInt(m.Id, 10)
	case "name":
		return m.Name
	case "priority":
		return strconv.Itoa(m.Priority)
	case "queue":
		return m.Queue.Name
	case "bucket":
		if m.Bucket != nil {
			return m.Bucket.Name
		}
		return ""
	case "agent":
		if m.Agent != nil {
			return m.Agent.Name
		}
		return ""
	case "skill":
		if m.Skill != nil {
			return m.Skill.Name
		}
		return ""
	case "timezone":
		return m.Timezone.Name
	case "attempts":
		return strconv.Itoa(m.Attempts)
	case "reserved":
		return strconv.FormatBool(m.Reserved)
	case "stop_cause":
		if m.StopCause != nil {
			return *m.StopCause
		}
		return ""
	case "stop_at":
		return formatMemberTimeExport(m.StopAt)
	case "created_at":
		return m.CreatedAt.Format("2006-01-02 15:04:05")
	case "expire_at":
		return formatMemberTimeExport(m.ExpireAt)
	case "ready_at", "min_offering_at":
		return formatMemberTimeExport(m.MinOfferingAt)
	case "last_hangup_at", "last_activity_at":
		if m.LastActivityAt == 0 {
			return ""
		}
		return time.UnixMilli(m.LastActivityAt).Format("2006-01-02 15:04:05")
	case "communications", "destination":
		destinations := make([]string, 0, len(m.Communications))
		for _, c := range m.Communications {
			destinations = append(destinations, c.Destination)
		}
		return strings.Join(destinations, ", ")
	case "variables":
		pairs := make([]string, 0, len(m.Variables))
		for k, v := range m.Variables {
			pairs = append(pairs, k+"="+v)
		}
		return strings.Join(pairs, ", ")
	default:
		return ""
	}
}

func formatMemberTimeExport(t *time.Time) string {
	if t == nil || t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02 15:04:05")
}

func generateMembersCSVChunk(headers []string, rows [][]string, page int, separator string) ([]byte, error) {
	buf := &bytes.Buffer{}

	if page == 1 {
		buf.Write([]byte{0xEF, 0xBB, 0xBF})
	}

	if separator != "" {
		if page == 1 {
			buf.WriteString(strings.Join(headers, separator))
			buf.WriteByte('\n')
		}
		for _, row := range rows {
			buf.WriteString(strings.Join(row, separator))
			buf.WriteByte('\n')
		}
		return buf.Bytes(), nil
	}

	writer := csv.NewWriter(buf)

	if page == 1 {
		if err := writer.Write(headers); err != nil {
			return nil, fmt.Errorf("failed to write CSV headers: %w", err)
		}
	}

	for _, row := range rows {
		if err := writer.Write(row); err != nil {
			return nil, fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("CSV writer error: %w", err)
	}

	return buf.Bytes(), nil
}

func generateMembersXLSX(headers []string, rows [][]string) ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()

	sheet := f.GetSheetName(0)
	sw, err := f.NewStreamWriter(sheet)
	if err != nil {
		return nil, fmt.Errorf("failed to create stream writer: %w", err)
	}

	for i := range headers {
		if err = sw.SetColWidth(i+1, i+1, 25); err != nil {
			return nil, fmt.Errorf("failed to set column width: %w", err)
		}
	}

	style, err := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true}})
	if err != nil {
		return nil, fmt.Errorf("failed to create style: %w", err)
	}

	headerRow := make([]any, len(headers))
	for i, h := range headers {
		headerRow[i] = excelize.Cell{StyleID: style, Value: h}
	}
	headerCell, _ := excelize.CoordinatesToCellName(1, 1)
	if err = sw.SetRow(headerCell, headerRow); err != nil {
		return nil, fmt.Errorf("failed to write header row: %w", err)
	}

	for rowIdx, row := range rows {
		rowData := make([]any, len(row))
		for i, v := range row {
			rowData[i] = v
		}
		cell, err := excelize.CoordinatesToCellName(1, rowIdx+2)
		if err != nil {
			return nil, fmt.Errorf("failed to get cell name: %w", err)
		}
		if err = sw.SetRow(cell, rowData); err != nil {
			return nil, fmt.Errorf("failed to write row %d: %w", rowIdx, err)
		}
	}

	if err = sw.Flush(); err != nil {
		return nil, fmt.Errorf("failed to flush stream writer: %w", err)
	}

	buf := &bytes.Buffer{}
	if err = f.Write(buf); err != nil {
		return nil, fmt.Errorf("failed to write XLSX: %w", err)
	}

	return buf.Bytes(), nil
}
