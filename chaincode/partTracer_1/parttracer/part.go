package parttracer

import (
	"encoding/json"
	"fmt"
)

// State : enum for maintaining the Part condition
type State uint

const (
	// NEW : 1
	NEW State = iota + 1
	// USED : 2
	USED
	// REFURBISHED : 3
	REFURBISHED
)

func (state State) String() string {
	names := []string{"NEW", "USED", "REFURBISHED"}

	if state < NEW || state > REFURBISHED {
		return "UNKNOWN"
	}

	return names[state-1]
}

// Part : structure to capture part details
type Part struct {
	PartID       string `json:"partId"`
	PartName     string `json:"partName"`
	Description  string `json:"description"`
	QuotePrice   uint32 `json:"quotePrice"`
	Manufacturer string `json:"manufacturer"`
	Owner        string `json:"owner"`
	EventTime    string `json:"eventTime"`
	Condition    State  `json:"condition"`
}

// PartDealData : structure to capture deal price for the part
type PartDealData struct {
	ObjectType string `json:"docType"` //docType is used to distinguish the various types of objects in state database
	PartID     string `json:"partId"`
	DealPrice  uint32 `json:"dealPrice"`
}

// SetDealPrice method
func (pdd *PartDealData) SetDealPrice(price uint32) {
	pdd.DealPrice = price
}

// SetOwner method
func (p *Part) SetOwner(newOwner string) {
	p.Owner = newOwner
}

//GetCondition method
func (p *Part) GetCondition() State {
	return p.Condition
}

//SetNew method
func (p *Part) SetNew() {
	p.Condition = NEW
}

//SetUsed method
func (p *Part) SetUsed() {
	p.Condition = USED
}

//SetRefurbished method
func (p *Part) SetRefurbished() {
	p.Condition = REFURBISHED
}

//IsNew method
func (p *Part) IsNew() bool {
	return p.Condition == NEW
}

//IsUsed method
func (p *Part) IsUsed() bool {
	return p.Condition == USED
}

//IsRefurbished method
func (p *Part) IsRefurbished() bool {
	return p.Condition == REFURBISHED
}

//Serialize : Method to Serialize the Part details
func (p *Part) Serialize() ([]byte, error) {
	return json.Marshal(p)
}

//Deserialize : Function to Deserialize the Part details
func Deserialize(bytes []byte, p *Part) error {
	err := json.Unmarshal(bytes, p)

	if err != nil {
		return fmt.Errorf("Error deserializing part. %s", err.Error())
	}

	return nil
}

//Serialize : Method to Serialize the Part Data details
func (pdd *PartDealData) Serialize() ([]byte, error) {
	return json.Marshal(pdd)
}

//DeserializeDealDetails : Function to Deserialize the Part Deal details
func DeserializeDealDetails(bytes []byte, pdd *PartDealData) error {
	err := json.Unmarshal(bytes, pdd)

	if err != nil {
		return fmt.Errorf("Error deserializing deal data. %s", err.Error())
	}

	return nil
}
