package domain

type SkuDTO struct {
	ID string `bson:"_id"`
}

type Hydrator struct {}

func NewHydrator() *Hydrator {
	return &Hydrator{}
}

func (h *Hydrator) Hydrate(dto *SkuDTO) *Sku {
	return &Sku{
		id: &SkuId{
			value: dto.ID,
		},
	}
}

func (h *Hydrator) Dehydrate(sku *Sku) *SkuDTO {
	return &SkuDTO{ID: sku.id.value}
}