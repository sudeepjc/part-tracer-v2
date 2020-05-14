package parttracer

import (
	"encoding/json"
	"fmt"
	s "strings"

	"github.com/golang/protobuf/ptypes"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// PartTrade : Business logic Contract
type PartTrade struct {
	contractapi.Contract
}

// InitLedger : Method to initialize the ledger
func (pt *PartTrade) InitLedger(ctx contractapi.TransactionContextInterface) error {
	fmt.Println("initLedger has been invoked")

	ci, _ := ctx.GetClientIdentity().GetID()
	fmt.Println("ClientIdentity : ", ci)

	msp, _ := ctx.GetClientIdentity().GetMSPID()
	fmt.Println("MSPID : ", msp)

	tx := ctx.GetStub().GetTxID()
	fmt.Println("TXID : ", tx)

	chanl := ctx.GetStub().GetChannelID()
	fmt.Println("ChannelID : ", chanl)

	tim, _ := ctx.GetStub().GetTxTimestamp()

	txTime, _ := ptypes.Timestamp(tim)

	PartID := s.Join([]string{"pName", txTime.Format("2006-01-02_5:04:05")}, "_")

	fmt.Println("Tx timestamp : ", PartID)

	return nil
}

// AddPart : Method to add a part to the ledger
func (pt *PartTrade) AddPart(ctx contractapi.TransactionContextInterface, partID string, pName string, desc string, qprice uint32, maker string) (string, error) {

	if len(partID) == 0 {
		return "", fmt.Errorf("Invalid part ID")
	}

	if len(pName) == 0 {
		return partID, fmt.Errorf("Invalid part Name info")
	}

	if len(desc) == 0 {
		return partID, fmt.Errorf("Invalid description ")
	}

	if len(maker) == 0 {
		return partID, fmt.Errorf("Invalid manufacturer info")
	}

	if qprice <= 0 {
		return partID, fmt.Errorf("Invalid quote price info")
	}

	// check if the part exists already and throw an error if it is so
	partAsBytes, err := ctx.GetStub().GetState(partID)

	if err != nil {
		return partID, fmt.Errorf("Failed to read from world state. %s", err.Error())
	}

	if partAsBytes != nil {
		return partID, fmt.Errorf("%s : already exists", partID)
	}

	owner, _ := ctx.GetClientIdentity().GetMSPID()

	part := Part{PartID: partID, PartName: pName, Description: desc, QuotePrice: qprice, Manufacturer: maker, Owner: owner}
	part.SetNew()

	// use tx time for deterministic behavior of the execution
	tim, _ := ctx.GetStub().GetTxTimestamp()
	txTime, _ := ptypes.Timestamp(tim)
	part.EventTime = txTime.Format("2006-01-02_5:04:05")

	partAsBytes, err = part.Serialize()

	if err != nil {
		return partID, fmt.Errorf("Failed to add part while serializing data %s", err.Error())
	}

	fmt.Println("added part ", partID)

	err = ctx.GetStub().PutState(partID, partAsBytes)

	if err != nil {
		return partID, fmt.Errorf("Error while trying to add sell data to state: %s", err.Error())
	}

	return partID, err
}

// QueryPart : Method to query the part given the partID
func (pt *PartTrade) QueryPart(ctx contractapi.TransactionContextInterface, partID string) (*Part, error) {

	if len(partID) == 0 {
		return nil, fmt.Errorf("Invalid part ID")
	}

	partAsBytes, err := ctx.GetStub().GetState(partID)

	if err != nil {
		return nil, fmt.Errorf("Failed to read from world state. %s", err.Error())
	}

	if partAsBytes == nil {
		return nil, fmt.Errorf("%s : does not exist", partID)
	}

	part := new(Part)
	_ = Deserialize(partAsBytes, part)

	return part, nil
}

// SellPart : Method to sell the part if the conditions meet
func (pt *PartTrade) SellPart(ctx contractapi.TransactionContextInterface, partID string, buyer string, privatePolicyName string) (*Part, error) {

	if len(partID) == 0 {
		return nil, fmt.Errorf("Invalid part ID")
	}

	if len(privatePolicyName) == 0 {
		return nil, fmt.Errorf("policy name should be one of the names in the policy collection json")
	}

	partAsBytes, err := ctx.GetStub().GetState(partID)

	if err != nil {
		return nil, fmt.Errorf("Failed to read from world state. %s", err.Error())
	}

	if partAsBytes == nil {
		return nil, fmt.Errorf("%s does not exist", partID)
	}

	part := new(Part)
	_ = Deserialize(partAsBytes, part)

	seller, _ := ctx.GetClientIdentity().GetMSPID()

	if part.Owner != seller {
		return nil, fmt.Errorf("Part %s is not owned by %s", partID, seller)
	}

	if part.IsNew() {
		part.SetUsed()
	}

	tim, _ := ctx.GetStub().GetTxTimestamp()
	txTime, _ := ptypes.Timestamp(tim)
	part.EventTime = txTime.Format("2006-01-02_5:04:05")

	part.SetOwner(buyer)

	updatedPartAsBytes, err := part.Serialize()

	if err != nil {
		return nil, fmt.Errorf("Failed to update part while serializing data %s", err.Error())
	}

	err = ctx.GetStub().PutState(partID, updatedPartAsBytes)

	if err != nil {
		return nil, fmt.Errorf("Error while trying to add sell data to state: %s", err.Error())
	}

	//handle DealData here
	type DealTransientInput struct {
		PartID    string `json:"partId"`
		DealPrice uint32 `json:"dealPrice"`
	}

	transMap, err := ctx.GetStub().GetTransient()
	if err != nil {
		return nil, fmt.Errorf("Error getting transient: " + err.Error())
	}

	partDealDataJSONBytes, ok := transMap["PartDealData"]
	if !ok {
		return nil, fmt.Errorf("PartDealData must be a key in the transient map")
	}

	if len(partDealDataJSONBytes) == 0 {
		return nil, fmt.Errorf("PartDealData value in the transient map must be a non-empty JSON string")
	}

	fmt.Println("Transmap :", transMap)
	fmt.Println("partDealDataJSONBytes : ", partDealDataJSONBytes)

	var partDealInput DealTransientInput
	err = json.Unmarshal(partDealDataJSONBytes, &partDealInput)
	if err != nil {
		return nil, fmt.Errorf("Failed to decode JSON of: " + string(partDealDataJSONBytes))
	}

	// ==== Create PartDealData object, marshal to JSON, and save to state ====
	ppd := &PartDealData{
		ObjectType: "PartDealData",
		PartID:     partDealInput.PartID,
		DealPrice:  partDealInput.DealPrice,
	}
	ppdJSONasBytes, err := json.Marshal(ppd)
	if err != nil {
		return nil, fmt.Errorf("Error during marshalling private data state" + err.Error())
	}

	err = ctx.GetStub().PutPrivateData(privatePolicyName, partID, ppdJSONasBytes)
	if err != nil {
		return nil, fmt.Errorf("Error during saving private data state" + err.Error())
	}

	fmt.Println("updated part ", partID)

	return part, nil
}

// QueryPartDealPrice : Method to query the part given the partID
func (pt *PartTrade) QueryPartDealPrice(ctx contractapi.TransactionContextInterface, partID string, privatePolicyName string) (*PartDealData, error) {

	if len(partID) == 0 {
		return nil, fmt.Errorf("Invalid part ID")
	}

	if len(privatePolicyName) == 0 {
		return nil, fmt.Errorf("policy name should be one of the names in the policy collection json")
	}

	partDealDataAsBytes, err := ctx.GetStub().GetPrivateData(privatePolicyName, partID)

	if err != nil {
		return nil, fmt.Errorf("Failed to read private details. %s", err.Error())
	}

	if partDealDataAsBytes == nil {
		return nil, fmt.Errorf("%s : does not exist", partID)
	}

	part := new(PartDealData)
	err = json.Unmarshal(partDealDataAsBytes, part)
	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshall private details. %s", err.Error())
	}

	return part, nil
}
