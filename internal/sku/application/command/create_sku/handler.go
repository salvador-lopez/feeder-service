package create_sku

import (
	"errors"
	"feeder-service/internal/sku/domain"
	"fmt"
)

type Command struct {
	Sku string
}

//go:generate mockgen -destination=mock/command_handler_interface_mockgen_mock.go -package=mock . CommandHandlerInterface
type CommandHandlerInterface interface {
	Handle(command Command) error
}

type CommandHandler  struct {
	repository domain.SkuRepository
}

func NewCommandHandler(repository domain.SkuRepository) *CommandHandler {
	return &CommandHandler{repository: repository}
}

var ErrSkuAlreadyExists = errors.New("sku already exists")
var ErrCreatingSku = errors.New("error creating sku")

func (h *CommandHandler) Handle(command Command) error {
	skuId, err := domain.NewSkuId(command.Sku)
	if err != nil {
		return err
	}
	//h.transactionalSession.BeginTransaction()
	skuEntity, err := h.repository.Find(skuId)
	if err != nil {
		return h.buildErrCreatingSku(skuId, err.Error())
	}
	if skuEntity != nil {
		return fmt.Errorf("%w: %s", ErrSkuAlreadyExists, skuId.Value())
	}

	err = h.repository.Save(domain.NewSku(skuId))
	if err != nil {
		//h.transactionalSession.Rollback()
		return h.buildErrCreatingSku(skuId, err.Error())
	}
	//h.transactionalSession.Commit()
	return nil
}

func (h *CommandHandler) buildErrCreatingSku(skuId *domain.SkuId, errorMsg string) error {
	return fmt.Errorf("%w %s: %s", ErrCreatingSku, skuId.Value(), errorMsg)
}


