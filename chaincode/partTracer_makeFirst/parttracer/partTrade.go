package parttracer

import (
	"fmt"
	"github.com/golang/protobuf/ptypes"
	s "strings"

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

	msp, _:= ctx.GetClientIdentity().GetMSPID()
	fmt.Println("MSPID : ", msp)

	tx:= ctx.GetStub().GetTxID()
	fmt.Println("TXID : ", tx)

	chanl:= ctx.GetStub().GetChannelID()
	fmt.Println("ChannelID : ", chanl)

	tim, _ := ctx.GetStub().GetTxTimestamp()

	txTime, _ := ptypes.Timestamp(tim)

	PartID := s.Join([]string{"pName",txTime.Format("2006-01-02_5:04:05")},"_")
	
	fmt.Println("Tx timestamp : ", PartID)


	return nil
}

// AddPart : Method to add a part to the ledger
func (pt *PartTrade) AddPart(ctx contractapi.TransactionContextInterface, partID string, pName string, desc string, qprice uint32, maker string) (string, error) {

	if len(partID) == 0 {
		return "",fmt.Errorf("Invalid part ID")
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


	owner, _:= ctx.GetClientIdentity().GetMSPID()

	// partID := s.Join([]string{pName,currentTime.Format("2006-01-02_5:04:05")},"_")
	part := Part{ PartID: partID, PartName: pName, Description: desc, QuotePrice: qprice, Manufacturer:maker, Owner:owner }
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

	err =  ctx.GetStub().PutState(partID, partAsBytes)

	if err != nil {
		return partID, fmt.Errorf("Error while trying to add sell data to state: %s", err.Error())
	}

	return partID,err
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
func (pt *PartTrade) SellPart(ctx contractapi.TransactionContextInterface, partID string, buyer string, dprice uint32 ) (*Part, error) {

	if len(partID) == 0 {
		return nil, fmt.Errorf("Invalid part ID")
	}

	if dprice <= 0 {
		return nil, fmt.Errorf("Invalid dprice")
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

	seller, _:= ctx.GetClientIdentity().GetMSPID()

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
	part.DealPrice = dprice

	updatedPartAsBytes, err := part.Serialize()

	if err != nil {
		return nil, fmt.Errorf("Failed to update part while serializing data %s", err.Error())
	}

	fmt.Println("updated part ", partID)

	err = ctx.GetStub().PutState(partID, updatedPartAsBytes)

	if err != nil {
		return nil, fmt.Errorf("Error while trying to add sell data to state: %s", err.Error())
	}

	return part, nil
}


