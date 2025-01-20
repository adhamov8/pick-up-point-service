package returns

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

type Return struct {
	ID           string `json:"id"`
	RecipientID  string `json:"recipient_id"`
	ReceivedDate string `json:"received_date"`
}

var returnsFile = "returns.json"

func AddReturn(returnInfo Return) {
	returns := ListReturns()

	for _, r := range returns {
		if r.ID == returnInfo.ID {
			fmt.Println("Return with this ID already exists.")
			return
		}
	}

	returns = append(returns, returnInfo)
	saveReturns(returns)
	fmt.Println("Return added successfully.")
}

func ListReturns() []Return {
	file, err := os.ReadFile(returnsFile)
	if err != nil {
		if os.IsNotExist(err) {
			return []Return{}
		}
		log.Fatal(err)
	}

	var returns []Return
	err = json.Unmarshal(file, &returns)
	if err != nil {
		log.Fatal(err)
	}

	return returns
}

func saveReturns(returns []Return) {
	file, err := os.Create(returnsFile)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	err = encoder.Encode(returns)
	if err != nil {
		log.Fatal(err)
	}
}
