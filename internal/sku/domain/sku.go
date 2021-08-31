package domain

import "context"

//go:generate mockgen -destination=mock/sku_repository_mockgen_mock.go -package=mock . SkuRepository
type SkuRepository interface {
	Find(context.Context, *SkuId) (*Sku, error)
	Save(context.Context, *Sku) error
}

type Sku struct {
	id *SkuId
}

func NewSku(id *SkuId) *Sku {
	return &Sku{id: id}
}

func (s Sku) Id() *SkuId {
	return s.id
}
