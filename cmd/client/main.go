package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	order_service "gitlab.ozon.dev/ashadkhamov/homework/pkg/order_service/v1"
	"google.golang.org/grpc/status"

	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func initConfig() {
	viper.SetConfigName("config")
	viper.AddConfigPath("./config")
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}
}

func main() {
	initConfig()

	grpcAddress := viper.GetString("client.grpc_address")
	if grpcAddress == "" {
		log.Fatal("client.grpc_address not set in config")
	}

	creds := insecure.NewCredentials()

	conn, err := grpc.Dial(grpcAddress, grpc.WithTransportCredentials(creds))
	if err != nil {
		log.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	client := order_service.NewOrderServiceClient(conn)

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Enter the command (or 'exit' to exit):")

	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}
		line := scanner.Text()
		if strings.TrimSpace(line) == "exit" {
			fmt.Println("Exiting...")
			break
		}

		args := strings.Fields(line)
		if len(args) == 0 {
			continue
		}

		command := args[0]
		cmdArgs := args[1:]

		var err error

		switch command {
		case "add":
			err = Add(client, cmdArgs)
		case "remove":
			err = Remove(client, cmdArgs)
		case "deliver":
			err = Deliver(client, cmdArgs)
		case "orders":
			err = Orders(client, cmdArgs)
		case "return":
			err = Return(client, cmdArgs)
		case "returns":
			err = Returns(client, cmdArgs)
		default:
			fmt.Printf("Unknown command: %s\n", command)
			continue
		}

		if err != nil {
			fmt.Printf("Error executing command '%s': %v\n", command, err)
		}
	}
}

func Add(client order_service.OrderServiceClient, args []string) error {
	if len(args) != 5 {
		return fmt.Errorf("usage: add [orderID] [recipientID] [expiryDate YYYY-MM-DD] [weight] [packaging (bag|box|film)]")
	}

	orderID := args[0]
	recipientID := args[1]
	expiryDate := args[2]
	weightStr := args[3]
	packagingType := args[4]

	weight, err := strconv.ParseFloat(weightStr, 32)
	if err != nil {
		return fmt.Errorf("invalid weight: %v", err)
	}

	req := &order_service.AddOrderRequest{
		OrderId:       orderID,
		RecipientId:   recipientID,
		ExpiryDate:    expiryDate,
		Weight:        float32(weight),
		PackagingType: packagingType,
	}

	_, err = client.AddOrder(context.Background(), req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			return fmt.Errorf("AddOrder failed: %v", st.Message())
		}
		return err
	}

	fmt.Println("Order added successfully.")
	return nil
}

func Deliver(client order_service.OrderServiceClient, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: deliver [recipientID] [orderID1] [orderID2] ...")
	}

	recipientID := args[0]
	orderIDs := args[1:]

	req := &order_service.DeliverOrdersRequest{
		RecipientId: recipientID,
		OrderIds:    orderIDs,
	}

	_, err := client.DeliverOrders(context.Background(), req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			return fmt.Errorf("DeliverOrders failed: %v", st.Message())
		}
		return err
	}

	fmt.Println("Orders delivered successfully.")
	return nil
}

func Orders(client order_service.OrderServiceClient, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: orders [recipientID]")
	}

	recipientID := args[0]

	req := &order_service.GetOrdersRequest{
		RecipientId: recipientID,
	}

	res, err := client.GetOrders(context.Background(), req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			return fmt.Errorf("GetOrders failed: %v", st.Message())
		}
		return err
	}

	for _, order := range res.Orders {
		fmt.Printf("Order ID: %s, Status: %s, Delivery Date: %s\n", order.OrderId, order.Status, order.DeliveryDate)
	}

	return nil
}

func Remove(client order_service.OrderServiceClient, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: remove [orderID]")
	}

	orderID := args[0]

	req := &order_service.RemoveOrderRequest{
		OrderId: orderID,
	}

	_, err := client.RemoveOrder(context.Background(), req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			return fmt.Errorf("RemoveOrder failed: %v", st.Message())
		}
		return err
	}

	fmt.Println("Order removed successfully.")
	return nil
}

func Return(client order_service.OrderServiceClient, args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("usage: return [recipientID] [orderID]")
	}

	recipientID := args[0]
	orderID := args[1]

	req := &order_service.AcceptReturnRequest{
		RecipientId: recipientID,
		OrderId:     orderID,
	}

	_, err := client.AcceptReturn(context.Background(), req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			return fmt.Errorf("AcceptReturn failed: %v", st.Message())
		}
		return err
	}

	fmt.Println("Return accepted successfully.")
	return nil
}

func Returns(client order_service.OrderServiceClient, args []string) error {
	page := int32(1)
	if len(args) >= 1 {
		p, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid page number: %v", err)
		}
		page = int32(p)
	}

	req := &order_service.GetReturnsRequest{
		Page: page,
	}

	res, err := client.GetReturns(context.Background(), req)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			return fmt.Errorf("GetReturns failed: %v", st.Message())
		}
		return err
	}

	for _, ret := range res.Returns {
		fmt.Printf("Order ID: %s, Return Date: %s\n", ret.OrderId, ret.ReturnDate)
	}

	return nil
}
