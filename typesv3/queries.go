package typesv3

import (
	"encoding/json"

	"github.com/CosmWasm/wasmvm/types"
)

//-------- Queries --------

type QueryResponse struct {
	Ok  []byte    `json:"Ok,omitempty"`
	Err *StdError `json:"Err,omitempty"`
}

// this is a thin wrapper around the desired Go API to give us types closer to Rust FFI
func RustQuery(querier types.Querier, binRequest []byte, gasLimit uint64) QuerierResult {
	var request types.QueryRequest
	err := json.Unmarshal(binRequest, &request)
	if err != nil {
		return QuerierResult{
			Err: &types.SystemError{
				InvalidRequest: &types.InvalidRequest{
					Err:     err.Error(),
					Request: binRequest,
				},
			},
		}
	}
	bz, err := querier.Query(request, gasLimit)
	return ToQuerierResult(bz, err)
}

// This is a 2-level result
type QuerierResult struct {
	Ok  *QueryResponse     `json:"Ok,omitempty"`
	Err *types.SystemError `json:"Err,omitempty"`
}

func ToQuerierResult(response []byte, err error) QuerierResult {
	if err == nil {
		return QuerierResult{
			Ok: &QueryResponse{
				Ok: response,
			},
		}
	}
	syserr := types.ToSystemError(err)
	if syserr != nil {
		return QuerierResult{
			Err: syserr,
		}
	}
	stderr := ToStdError(err)
	return QuerierResult{
		Ok: &QueryResponse{
			Err: stderr,
		},
	}
}
