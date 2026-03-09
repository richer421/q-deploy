package hello_world

import (
	"context"

	"github.com/richer421/q-deploy/infra/mysql/dao"
	"github.com/richer421/q-deploy/infra/mysql/model"
)

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) Create(ctx context.Context, m *model.HelloWorld) error {
	return dao.HelloWorld.WithContext(ctx).Create(m)
}

func (s *Service) GetByID(ctx context.Context, id uint) (*model.HelloWorld, error) {
	return dao.HelloWorld.WithContext(ctx).Where(dao.HelloWorld.ID.Eq(id)).First()
}

func (s *Service) List(ctx context.Context, offset, limit int) ([]*model.HelloWorld, int64, error) {
	return dao.HelloWorld.WithContext(ctx).
		Order(dao.HelloWorld.ID.Desc()).
		FindByPage(offset, limit)
}

func (s *Service) Update(ctx context.Context, id uint, updates map[string]any) error {
	_, err := dao.HelloWorld.WithContext(ctx).
		Where(dao.HelloWorld.ID.Eq(id)).
		Updates(updates)
	return err
}

func (s *Service) Delete(ctx context.Context, id uint) error {
	_, err := dao.HelloWorld.WithContext(ctx).
		Where(dao.HelloWorld.ID.Eq(id)).
		Delete()
	return err
}
