package create_sku

import (
	"errors"
	"feeder-service/internal/sku/domain"
	"fmt"
)

type Command struct {
	Sku string
}

type CommandHandler  struct {
	repository domain.SkuRepository
}

func NewCommandHandler(repository domain.SkuRepository) *CommandHandler {
	return &CommandHandler{repository: repository}
}

var ErrCreatingSku = errors.New("error creating sku")
func (h *CommandHandler) Handle(command Command) error {
	skuId, err := domain.NewSkuId(command.Sku)
	if err != nil {
		return err
	}

	skuEntity, err := h.repository.Find(skuId)
	if err != nil {
		return h.buildErrCreatingSku(skuId, err.Error())
	}
	if skuEntity != nil {
		return h.buildErrCreatingSku(skuId, "already exists")
	}

	err = h.repository.Save(domain.NewSku(skuId))
	if err != nil {
		return h.buildErrCreatingSku(skuId, err.Error())
	}

	return nil
}

func (h *CommandHandler) buildErrCreatingSku(skuId *domain.SkuId, errorMsg string) error {
	return fmt.Errorf("%w %s: %s", ErrCreatingSku, skuId.Value(), errorMsg)
}


