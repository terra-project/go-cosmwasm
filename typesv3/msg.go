package typesv3

import (
	"github.com/CosmWasm/wasmvm/types"
)

//------- Results / Msgs -------------

// HandleResult is the raw response from the handle call
type HandleResult struct {
	Ok  *HandleResponse `json:"Ok,omitempty"`
	Err *StdError       `json:"Err,omitempty"`
}

// HandleResponse defines the return value on a successful handle
type HandleResponse struct {
	// Messages comes directly from the contract and is it's request for action
	Messages []types.CosmosMsg `json:"messages"`
	// base64-encoded bytes to return as ABCI.Data field
	Data []byte `json:"data"`
	// log message to return over abci interface
	Log []types.EventAttribute `json:"log"`
}

// InitResult is the raw response from the handle call
type InitResult struct {
	Ok  *InitResponse `json:"Ok,omitempty"`
	Err *StdError     `json:"Err,omitempty"`
}

// InitResponse defines the return value on a successful handle
type InitResponse struct {
	// Messages comes directly from the contract and is it's request for action
	Messages []types.CosmosMsg `json:"messages"`
	// log message to return over abci interface
	Log []types.EventAttribute `json:"log"`
}

// MigrateResult is the raw response from the handle call
type MigrateResult struct {
	Ok  *MigrateResponse `json:"Ok,omitempty"`
	Err *StdError        `json:"Err,omitempty"`
}

// MigrateResponse defines the return value on a successful handle
type MigrateResponse struct {
	// Messages comes directly from the contract and is it's request for action
	Messages []types.CosmosMsg `json:"messages"`
	// base64-encoded bytes to return as ABCI.Data field
	Data []byte `json:"data"`
	// log message to return over abci interface
	Log []types.EventAttribute `json:"log"`
}
