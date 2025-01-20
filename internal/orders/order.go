package orders

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

type Order struct {
	ID          string `json:"id"`
	RecipientID string `json:"recipient_id"`
	ExpiryDate  string `json:"expiry_date"`
	Status      string `json:"status"`
}

var ordersFile = "orders.json"

func AddOrder(order Order) {
	orders := ListOrders()

	for _, o := range orders {
		if o.ID == order.ID {
			fmt.Println("Order with this ID already exists.")
			return
		}
	}

	orders = append(orders, order)
	saveOrders(orders)
	fmt.Println("Order added successfully.")
}

func ListOrders() []Order {
	file, err := os.ReadFile(ordersFile)
	if err != nil {
		if os.IsNotExist(err) {
			return []Order{}
		}
		log.Fatal(err)
	}

	var orders []Order
	err = json.Unmarshal(file, &orders)
	if err != nil {
		log.Fatal(err)
	}

	return orders
}

func saveOrders(orders []Order) {
	file, err := os.Create(ordersFile)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	err = encoder.Encode(orders)
	if err != nil {
		log.Fatal(err)
	}
}
