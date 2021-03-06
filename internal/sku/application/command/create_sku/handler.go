package create_sku

import (
	"context"
	"errors"
	"feeder-service/internal/sku/domain"
	"fmt"
)

type Command struct {
	Sku string
}

//go:generate mockgen -destination=mock/command_handler_interface_mockgen_mock.go -package=mock . CommandHandlerInterface
type CommandHandlerInterface interface {
	Handle(context.Context, Command) error
}

type CommandHandler  struct {
	repository domain.SkuRepository
}

func NewCommandHandler(repository domain.SkuRepository) *CommandHandler {
	return &CommandHandler{repository: repository}
}

var ErrCreatingSku = errors.New("error creating sku")

func (h *CommandHandler) Handle(ctx context.Context, command Command) error {
	skuId, err := domain.NewSkuId(command.Sku)
	if err != nil {
		return err
	}

	err = h.repository.Save(ctx, domain.NewSku(skuId))
	if err != nil {
		if errors.Is(err, domain.ErrSkuAlreadyExists) {
			return err
		}
		return h.buildErrCreatingSku(skuId, err.Error())
	}

	return nil
}

func (h *CommandHandler) buildErrCreatingSku(skuId *domain.SkuId, errorMsg string) error {
	return fmt.Errorf("%w %s: %s", ErrCreatingSku, skuId.Value(), errorMsg)
}


