package hello_world

import (
	"context"

	"github.com/richer421/q-deploy/app/hello_world/vo"
	domain "github.com/richer421/q-deploy/domain/hello_world"
	"github.com/richer421/q-deploy/infra/mysql/model"
)

type AppService struct {
	svc *domain.Service
}

func NewAppService() *AppService {
	return &AppService{svc: domain.NewService()}
}

func (a *AppService) Create(ctx context.Context, req *vo.CreateReq) (*vo.HelloWorldResp, error) {
	m := &model.HelloWorld{
		Title:       req.Title,
		Description: req.Description,
	}
	if err := a.svc.Create(ctx, m); err != nil {
		return nil, err
	}
	return toResp(m), nil
}

func (a *AppService) Get(ctx context.Context, id uint) (*vo.HelloWorldResp, error) {
	m, err := a.svc.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return toResp(m), nil
}

func (a *AppService) List(ctx context.Context, req *vo.ListReq) (*vo.ListResp, error) {
	offset := (req.Page - 1) * req.PageSize
	items, total, err := a.svc.List(ctx, offset, req.PageSize)
	if err != nil {
		return nil, err
	}
	resp := &vo.ListResp{Total: total}
	for _, m := range items {
		resp.Items = append(resp.Items, toResp(m))
	}
	return resp, nil
}

func (a *AppService) Update(ctx context.Context, id uint, req *vo.UpdateReq) error {
	updates := make(map[string]any)
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if len(updates) == 0 {
		return nil
	}
	return a.svc.Update(ctx, id, updates)
}

func (a *AppService) Delete(ctx context.Context, id uint) error {
	return a.svc.Delete(ctx, id)
}

func toResp(m *model.HelloWorld) *vo.HelloWorldResp {
	return &vo.HelloWorldResp{
		ID:          m.ID,
		Title:       m.Title,
		Description: m.Description,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}
