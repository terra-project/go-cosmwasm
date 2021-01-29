package typesv3

import "github.com/CosmWasm/wasmvm/types"

//---------- Env ---------

// Env defines the state of the blockchain environment this contract is
// running in. This must contain only trusted data - nothing from the Tx itself
// that has not been verfied (like Signer).
//
// Env are json encoded to a byte slice before passing to the wasm contract.
type Env struct {
	Block    BlockInfo          `json:"block"`
	Message  types.MessageInfo  `json:"message"`
	Contract types.ContractInfo `json:"contract"`
}

type BlockInfo struct {
	// block height this transaction is executed
	Height uint64 `json:"height"`
	// time in seconds since unix epoch - since cosmwasm 0.3
	Time    uint64 `json:"time"`
	ChainID string `json:"chain_id"`
}

type HumanizeAddress func([]byte) (string, uint64, error)
type CanonicalizeAddress func(string) ([]byte, uint64, error)
