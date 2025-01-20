package dto

import "github.com/go-playground/validator/v10"

type AddOrderDTO struct {
	OrderID       string  `validate:"required"`
	RecipientID   string  `validate:"required"`
	ExpiryDate    string  `validate:"required,datetime=2006-01-02"`
	Weight        float32 `validate:"gt=0"`
	PackagingType string  `validate:"required,oneof=bag box film"`
}

func (dto *AddOrderDTO) Validate() error {
	validate := validator.New()
	return validate.Struct(dto)
}
