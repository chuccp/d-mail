package manage

import (
	"github.com/chuccp/d-mail/core"
	"github.com/chuccp/d-mail/db"
	"github.com/chuccp/d-mail/web"
	"net/mail"
	"strconv"
)

type Mail struct {
	context *core.Context
}

func (m *Mail) getOne(req *web.Request) (any, error) {
	id := req.Param("id")
	atoi, err := strconv.Atoi(id)
	if err != nil {
		return nil, err
	}
	one, err := m.context.GetDb().GetMailModel().GetOne(uint(atoi))
	if err != nil {
		return nil, err
	}
	return one, nil
}

func (m *Mail) deleteOne(req *web.Request) (any, error) {
	id := req.Param("id")
	atoi, err := strconv.Atoi(id)
	if err != nil {
		return nil, err
	}
	err = m.context.GetDb().GetMailModel().DeleteOne(uint(atoi))
	if err != nil {
		return nil, err
	}
	return "ok", nil
}
func (m *Mail) getPage(req *web.Request) (any, error) {
	page := req.GetPage()
	p, err := m.context.GetDb().GetMailModel().Page(page)
	if err != nil {
		return nil, err
	}
	return p, nil
}
func (m *Mail) postOne(req *web.Request) (any, error) {
	var st db.Mail
	err := req.ShouldBindBodyWithJSON(&st)
	if err != nil {
		return nil, err
	}
	_, err = mail.ParseAddress(st.Mail)
	if err != nil {
		return nil, err
	}
	err = m.context.GetDb().GetMailModel().Save(&st)
	if err != nil {
		return nil, err
	}
	return "ok", nil
}
func (m *Mail) putOne(req *web.Request) (any, error) {
	var st db.Mail
	err := req.ShouldBindBodyWithJSON(&st)
	if err != nil {
		return nil, err
	}

	_, err = mail.ParseAddress(st.Mail)
	if err != nil {
		return nil, err
	}

	err = m.context.GetDb().GetMailModel().Edit(&st)
	if err != nil {
		return nil, err
	}
	return "ok", nil
}
func (m *Mail) Init(context *core.Context, server core.IHttpServer) {
	m.context = context
	server.GET("/mail/:id", m.getOne)
	server.DELETE("/mail/:id", m.deleteOne)
	server.GET("/mail", m.getPage)
	server.POST("/mail", m.postOne)
	server.PUT("/mail", m.putOne)
}
