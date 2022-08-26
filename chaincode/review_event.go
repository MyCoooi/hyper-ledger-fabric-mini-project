package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing a event
type SmartContract struct {
	contractapi.Contract
}

// Event 구조체 정의
type Event struct {
	Type	string `json:"type"` // 유형
	Host	string `json:"host"` // 주최자
	Target	string `json:"target"` // 타겟 상품
	ServiceProduct	string `json:"serviceProduct"` // 서비스 상품
	MinPrice	int `json:"minPrice"` // 참여 가능 최소금액
	MaxNum	int `json:"maxLimit"` // 참여 가능 최대인원
	ExpireDate	string `json:"expireDate"` // 이벤트 종료일
	User	string 	`json:"user"` // 참여한 사용자
	Status	string `json:"status"` // 진행 상태(등록/진행중/완료/미완료)
	Timestamp time.Time `json:"timestamp"`
}

// history 결과 저장 구조체
type HistoryQueryResult struct {
	Record    *Event    `json:"record"`
	TxId     string    `json:"txId"`
	Timestamp time.Time `json:"timestamp"`
	IsDelete  bool      `json:"isDelete"`
}

// QueryResult structure used for handling result of query
type QueryResult struct {
	Key    string `json:"Key"`
	Record *Event
}

// InitLedger adds a base set of review-events to the ledger
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	events := []Event{
		Event{
			Type:"default",
			Host:"starbucks",
			Target:"caffe-latte",
			ServiceProduct:"banila-cookie_1,americano_1",
			MinPrice:14000,
			MaxNum:50,
			ExpireDate:"2022-08-26",
			User:"",
			Status:"regi",
		},
		Event{Type:"receipt",Host:"seoul-milk",Target:"milk",ServiceProduct:"cheese_10",MinPrice:6000,MaxNum:100,ExpireDate:"2022-10-24",User:"",Status:"regi"},
		Event{Type:"photo",Host:"yellow-chicken",Target:"snow-chicken",ServiceProduct:"cheese-ball_2,hotdog_1,coke_1",MinPrice:19000,MaxNum:0,ExpireDate:"2022-09-01",User:"",Status:"regi"},
		Event{Type:"text",Host:"pizza-hut",Target:"bulgogi-pizza,mozza-pizza",ServiceProduct:"coke_1,cheese-crust_1",MinPrice:24000,MaxNum:100,ExpireDate:"2022-08-30",User:"",Status:"regi"},
	}

	for i, event := range events {
		eventAsBytes, _ := json.Marshal(event)
		err := ctx.GetStub().PutState("Event"+strconv.Itoa(i), eventAsBytes)

		if err != nil {
			return fmt.Errorf("Failed to put to world state. %s", err.Error())
		}
	}

	return nil
}

// RegisterEvent adds a new event to the world state with given details
func (s *SmartContract) RegisterEvent(ctx contractapi.TransactionContextInterface, eventNum string, type1 string, host string, target string, serviceProduct string, minPrice int, maxNum int, expireDate string) error {
	event := Event{
		Type:	type1,
		Host:	host,
		Target: target,
		ServiceProduct: serviceProduct,
		MinPrice: minPrice,
		MaxNum: maxNum,
		ExpireDate: expireDate,
		User: "",
		Status: "regi",
	}

	eventAsBytes, _ := json.Marshal(event)

	return ctx.GetStub().PutState(eventNum, eventAsBytes)
}

// QueryEvent returns the event stored in the world state with given id
func (s *SmartContract) QueryEvent(ctx contractapi.TransactionContextInterface, eventNum string) (*Event, error) {
	eventAsBytes, err := ctx.GetStub().GetState(eventNum)

	if err != nil {
		return nil, fmt.Errorf("Failed to read from world state. %s", err.Error())
	}

	if eventAsBytes == nil {
		return nil, fmt.Errorf("%s does not exist", eventNum)
	}

	event := new(Event)
	_ = json.Unmarshal(eventAsBytes, event)

	return event, nil
}

// QueryAllEvents returns all events found in world state
func (s *SmartContract) QueryAllEvents(ctx contractapi.TransactionContextInterface) ([]QueryResult, error) {
	startKey := ""
	endKey := ""

	resultsIterator, err := ctx.GetStub().GetStateByRange(startKey, endKey)

	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	results := []QueryResult{}

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()

		if err != nil {
			return nil, err
		}

		event := new(Event)
		_ = json.Unmarshal(queryResponse.Value, event)

		queryResult := QueryResult{Key: queryResponse.Key, Record: event}
		results = append(results, queryResult)
	}

	return results, nil
}

// ChangeEventUser updates the user field of event with given id in world state
func (s *SmartContract) ChangeEventUser(ctx contractapi.TransactionContextInterface, eventNum string, newUser string) error {
	event, err := s.QueryEvent(ctx, eventNum)

	if err != nil {
		return err
	}

	event.User = newUser

	eventAsBytes, _ := json.Marshal(event)

	return ctx.GetStub().PutState(eventNum, eventAsBytes)
}

// ChangeEventStatus updates the user field of event with given id in world state
func (s *SmartContract) ChangeEventStatus(ctx contractapi.TransactionContextInterface, eventNum string, newStatus string) error {
	event, err := s.QueryEvent(ctx, eventNum)

	if err != nil {
		return err
	}

	event.Status = newStatus

	eventAsBytes, _ := json.Marshal(event)

	return ctx.GetStub().PutState(eventNum, eventAsBytes)
}

func (t *SmartContract) GetHistory(ctx contractapi.TransactionContextInterface, key string) ([]HistoryQueryResult, error) {
	log.Printf("GetHistory: ID %v", key)

	resultsIterator, err := ctx.GetStub().GetHistoryForKey(key)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var records []HistoryQueryResult
	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var asset Event
		if len(response.Value) > 0 {
			err = json.Unmarshal(response.Value, &asset)
			if err != nil {
				return nil, err
			}
		} else {
			asset = Event{
			}
		}

		timestamp, err := ptypes.Timestamp(response.Timestamp)
		if err != nil {
			return nil, err
		}

		record := HistoryQueryResult{
			TxId:      response.TxId,
			Timestamp: timestamp,
			Record:    &asset,
			IsDelete:  response.IsDelete,
		}
		records = append(records, record)
	}

	return records, nil
}

func main() {

	chaincode, err := contractapi.NewChaincode(new(SmartContract))

	if err != nil {
		fmt.Printf("Error create reviewevent chaincode: %s", err.Error())
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting reviewevent chaincode: %s", err.Error())
	}
}
