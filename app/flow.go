package app

import (
	workflow "buf.build/gen/go/webitel/workflow/protocolbuffers/go"
	"github.com/webitel/engine/model"
)

func (app *App) StartAsyncFlow(domainId int64, schemaId int, variables map[string]string) (string, model.AppError) {
	execId, err := app.flowManager.Queue().StartFlow(&workflow.StartFlowRequest{
		SchemaId:  uint32(schemaId), // todo
		DomainId:  domainId,
		Variables: variables,
	})

	if err != nil {
		return "", model.NewInternalError("app.flow.async", err.Error())
	}

	return execId, nil
}
