package domain

import "fmt"

type PackagingStrategy interface {
	Apply(order *Order) error
	GetCostIncrease() float32
}

type BagPackaging struct{}

func (p *BagPackaging) Apply(order *Order) error {
	if order.Weight >= 10 {
		return fmt.Errorf("order weight exceeds 10 kg, cannot use bag")
	}
	return nil
}

func (p *BagPackaging) GetCostIncrease() float32 {
	return 5
}

type BoxPackaging struct{}

func (p *BoxPackaging) Apply(order *Order) error {
	if order.Weight >= 30 {
		return fmt.Errorf("order weight exceeds 30 kg, cannot use box")
	}
	return nil
}

func (p *BoxPackaging) GetCostIncrease() float32 {
	return 20
}

type FilmPackaging struct{}

func (p *FilmPackaging) Apply(order *Order) error {
	return nil
}

func (p *FilmPackaging) GetCostIncrease() float32 {
	return 1
}

func GetPackagingStrategy(packagingType string) (PackagingStrategy, error) {
	switch packagingType {
	case "bag":
		return &BagPackaging{}, nil
	case "box":
		return &BoxPackaging{}, nil
	case "film":
		return &FilmPackaging{}, nil
	default:
		return nil, fmt.Errorf("invalid packaging type")
	}
}
