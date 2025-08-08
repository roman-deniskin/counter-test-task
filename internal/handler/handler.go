package handler

import "counter-test-task/internal/service"

type Handler struct {
	Srv *service.Service
}

func New(srv *service.Service) *Handler {
	return &Handler{Srv: srv}
}
