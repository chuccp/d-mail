package rest

import (
	"encoding/json"
	"errors"
	"os"
	"strconv"

	wf "github.com/chuccp/go-web-frame"
	auth2 "github.com/chuccp/go-web-frame/component/auth"
	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/log"
	"github.com/chuccp/go-web-frame/web"
	"github.com/chuccp/http2smtp/auth"
	"github.com/chuccp/http2smtp/entity"
	"github.com/chuccp/http2smtp/model"
	"github.com/chuccp/http2smtp/smtp"
	"go.uber.org/zap"
)

type Log struct {
	context  *core.Context
	logModel *model.LogModel
}

func (l *Log) getOne(req *web.Request) (any, error) {
	id := req.Param("id")
	atoi, err := strconv.Atoi(id)
	if err != nil {
		return nil, err
	}
	user, err := auth.User(req, l.context)
	if user == nil {
		return nil, err
	}
	one, err := l.logModel.Query().Where("id = ? AND user_id = ?", uint(atoi), user.Id).One()
	if err != nil {
		return nil, err
	}
	if one != nil {
		one.StatusStr = entity.StatusText(one.Status)
	}
	return one, nil
}

func (l *Log) getPage(req *web.Request) (any, error) {
	page, err := req.Page()
	if err != nil {
		return nil, err
	}
	user, err := auth.User(req, l.context)
	if user == nil {
		return nil, err
	}
	searchKey := req.GetFormParam("searchKey")
	query := l.logModel.Query().Where("user_id = ?", user.Id)
	if searchKey != "" {
		query = query.Where("(name LIKE ? OR mail LIKE ? OR subject LIKE ?)", "%"+searchKey+"%", "%"+searchKey+"%", "%"+searchKey+"%")
	}
	result, err := query.Order("id desc").PageForWeb(page)
	if err != nil {
		return nil, err
	}
	for _, one := range result.List {
		if one != nil {
			one.StatusStr = entity.StatusText(one.Status)
		}
	}
	return result, nil
}

// downLoad serves an attachment belonging to one of the caller's own mail logs.
// The path is resolved from the log record, never from client input.
func (l *Log) downLoad(req *web.Request) (any, error) {
	logId, err := strconv.Atoi(req.GetFormParam("logId"))
	if err != nil {
		return nil, errors.New("logId is required")
	}
	fileName := req.GetFormParam("file")
	if fileName == "" {
		return nil, errors.New("file is required")
	}
	user, err := auth.User(req, l.context)
	if user == nil {
		return nil, err
	}
	one, err := l.logModel.Query().Where("id = ? AND user_id = ?", uint(logId), user.Id).One()
	if err != nil {
		return nil, err
	}
	if one == nil {
		return nil, errors.New("log not found")
	}
	filePath, err := FilePathOf(one.Files, fileName)
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(filePath); err != nil {
		return nil, errors.New("attachment file is no longer available")
	}
	log.Info("downLoad", zap.Uint("logId", uint(logId)), zap.String("file", fileName), zap.Uint("userId", user.Id))
	return web.CreateFileResponse(filePath), nil
}

// FilePathOf resolves the stored path of a named attachment from a log's files JSON.
func FilePathOf(filesJSON string, name string) (string, error) {
	if filesJSON == "" {
		return "", errors.New("log has no attachments")
	}
	var files []*smtp.File
	if err := json.Unmarshal([]byte(filesJSON), &files); err != nil {
		return "", errors.New("invalid attachment record")
	}
	for _, file := range files {
		if file != nil && file.Name == name && file.FilePath != "" {
			return file.FilePath, nil
		}
	}
	return "", errors.New("attachment not found")
}

func (l *Log) Init(context *core.Context) error {
	l.context = context
	l.logModel = wf.GetModel[*model.LogModel](context)
	context.Get("/log/:id", l.getOne).WithMeta(auth2.WithLogin())
	context.Get("/log", l.getPage).WithMeta(auth2.WithLogin())
	context.Get("/download", l.downLoad).WithMeta(auth2.WithLogin())
	return nil

}
