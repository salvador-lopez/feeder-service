package mongo

import "feeder-service/internal/sku/domain"

type SkuRepository struct {

}

func NewSkuRepository() *SkuRepository {
	return &SkuRepository{}
}

func (r *SkuRepository) Find(id *domain.SkuId) (*domain.Sku, error) {
	return nil, nil
}
func (r *SkuRepository) Save(*domain.Sku) error {
	return nil
}
