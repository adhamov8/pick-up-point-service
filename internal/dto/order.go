package dto

type OrderDTO struct {
	OrderID       string
	RecipientID   string
	ExpiryDate    string
	Status        string
	DeliveryDate  string
	ReturnDate    string
	Weight        float32
	Cost          float32
	PackagingType string
}
