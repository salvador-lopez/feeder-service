package domain

import (
	"errors"
	"fmt"
	"regexp"
)

type SkuId struct {
	value string
}

func (i SkuId) Equal(anotherSkuId *SkuId) bool {
	return anotherSkuId.value == i.value
}

var ErrInvalidSku = errors.New("invalid Sku provided")
func NewSkuId(value string) (*SkuId, error) {
	match, err := regexp.MatchString("^[A-Z]{4}-[0-9]{4}$", value)
	if err != nil || !match {
		return nil, fmt.Errorf("%w: %s", ErrInvalidSku, value)
	}

	return &SkuId{value: value}, nil
}

func (i *SkuId) Value() string {
	return i.value
}