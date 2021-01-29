package apiv3

/*
#include "bindings.h"

// typedefs for _cgo functions (db)
typedef V3_GoResult (*read_db_fn)(V3_db_t *ptr, V3_gas_meter_t *gas_meter, uint64_t *used_gas, V3_Buffer key, V3_Buffer *val, V3_Buffer *errOut);
typedef V3_GoResult (*write_db_fn)(V3_db_t *ptr, V3_gas_meter_t *gas_meter, uint64_t *used_gas, V3_Buffer key, V3_Buffer val, V3_Buffer *errOut);
typedef V3_GoResult (*remove_db_fn)(V3_db_t *ptr, V3_gas_meter_t *gas_meter, uint64_t *used_gas, V3_Buffer key, V3_Buffer *errOut);
typedef V3_GoResult (*scan_db_fn)(V3_db_t *ptr, V3_gas_meter_t *gas_meter, uint64_t *used_gas, V3_Buffer start, V3_Buffer end, int32_t order, V3_GoIter *out, V3_Buffer *errOut);
// iterator
typedef V3_GoResult (*next_db_fn)(V3_iterator_t idx, V3_gas_meter_t *gas_meter, uint64_t *used_gas, V3_Buffer *key, V3_Buffer *val, V3_Buffer *errOut);
// and api
typedef V3_GoResult (*humanize_address_fn)(V3_api_t *ptr, V3_Buffer canon, V3_Buffer *human, V3_Buffer *errOut, uint64_t *used_gas);
typedef V3_GoResult (*canonicalize_address_fn)(V3_api_t *ptr, V3_Buffer human, V3_Buffer *canon, V3_Buffer *errOut, uint64_t *used_gas);
typedef V3_GoResult (*query_external_fn)(V3_querier_t *ptr, uint64_t gas_limit, uint64_t *used_gas, V3_Buffer request, V3_Buffer *result, V3_Buffer *errOut);

// forward declarations (db)
V3_GoResult cGetV3_cgo(V3_db_t *ptr, V3_gas_meter_t *gas_meter, uint64_t *used_gas, V3_Buffer key, V3_Buffer *val, V3_Buffer *errOut);
V3_GoResult cSetV3_cgo(V3_db_t *ptr, V3_gas_meter_t *gas_meter, uint64_t *used_gas, V3_Buffer key, V3_Buffer val, V3_Buffer *errOut);
V3_GoResult cDeleteV3_cgo(V3_db_t *ptr, V3_gas_meter_t *gas_meter, uint64_t *used_gas, V3_Buffer key, V3_Buffer *errOut);
V3_GoResult cScanV3_cgo(V3_db_t *ptr, V3_gas_meter_t *gas_meter, uint64_t *used_gas, V3_Buffer start, V3_Buffer end, int32_t order, V3_GoIter *out, V3_Buffer *errOut);
// iterator
V3_GoResult cNextV3_cgo(V3_iterator_t *ptr, V3_gas_meter_t *gas_meter, uint64_t *used_gas, V3_Buffer *key, V3_Buffer *val, V3_Buffer *errOut);
// api
V3_GoResult cHumanAddressV3_cgo(V3_api_t *ptr, V3_Buffer canon, V3_Buffer *human, V3_Buffer *errOut, uint64_t *used_gas);
V3_GoResult cCanonicalAddressV3_cgo(V3_api_t *ptr, V3_Buffer human, V3_Buffer *canon, V3_Buffer *errOut, uint64_t *used_gas);
// and querier
V3_GoResult cQueryExternalV3_cgo(V3_querier_t *ptr, uint64_t gas_limit, uint64_t *used_gas, V3_Buffer request, V3_Buffer *result, V3_Buffer *errOut);


*/
import "C"

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"unsafe"

	"github.com/CosmWasm/wasmvm/typesv3"
	dbm "github.com/tendermint/tm-db"
)

// Note: we have to include all exports in the same file (at least since they both import bindings.h),
// or get odd cgo build errors about duplicate definitions

func recoverPanic(ret *C.V3_GoResult) {
	rec := recover()
	// we don't want to import cosmos-sdk
	// we also cannot use interfaces to detect these error types (as they have no methods)
	// so, let's just rely on the descriptive names
	// this is used to detect "out of gas panics"
	if rec != nil {
		name := reflect.TypeOf(rec).Name()
		switch name {
		// These two cases are for types thrown in panics from this module:
		// https://github.com/cosmos/cosmos-sdk/blob/4ffabb65a5c07dbb7010da397535d10927d298c1/store/types/gas.go
		// ErrorOutOfGas needs to be propagated through the rust code and back into go code, where it should
		// probably be thrown in a panic again.
		// TODO figure out how to pass the text in its `Descriptor` field through all the FFI
		// TODO handle these cases on the Rust side in the first place
		case "ErrorOutOfGas":
			*ret = C.V3_GoResult_OutOfGas
		// Looks like this error is not treated specially upstream:
		// https://github.com/cosmos/cosmos-sdk/blob/4ffabb65a5c07dbb7010da397535d10927d298c1/baseapp/baseapp.go#L818-L853
		// but this needs to be periodically verified, in case they do start checking for this type
		// 	case "ErrorGasOverflow":
		default:
			log.Printf("Panic in Go callback: %#v\n", rec)
			*ret = C.V3_GoResult_Panic
		}
	}
}

type Gas = uint64

// GasMeter is a copy of an interface declaration from cosmos-sdk
// https://github.com/cosmos/cosmos-sdk/blob/18890a225b46260a9adc587be6fa1cc2aff101cd/store/types/gas.go#L34
type GasMeter interface {
	GasConsumed() Gas
}

/****** DB ********/

// KVStore copies a subset of types from cosmos-sdk
// We may wish to make this more generic sometime in the future, but not now
// https://github.com/cosmos/cosmos-sdk/blob/bef3689245bab591d7d169abd6bea52db97a70c7/store/types/store.go#L170
type KVStore interface {
	Get(key []byte) []byte
	Set(key, value []byte)
	Delete(key []byte)

	// Iterator over a domain of keys in ascending order. End is exclusive.
	// Start must be less than end, or the Iterator is invalid.
	// Iterator must be closed by caller.
	// To iterate over entire domain, use store.Iterator(nil, nil)
	Iterator(start, end []byte) dbm.Iterator

	// Iterator over a domain of keys in descending order. End is exclusive.
	// Start must be less than end, or the Iterator is invalid.
	// Iterator must be closed by caller.
	ReverseIterator(start, end []byte) dbm.Iterator
}

var db_vtable = C.V3_DB_vtable{
	read_db:   (C.read_db_fn)(C.cGetV3_cgo),
	write_db:  (C.write_db_fn)(C.cSetV3_cgo),
	remove_db: (C.remove_db_fn)(C.cDeleteV3_cgo),
	scan_db:   (C.scan_db_fn)(C.cScanV3_cgo),
}

type DBState struct {
	Store KVStore
	// IteratorStackID is used to lookup the proper stack frame for iterators associated with this DB (iterator.go)
	IteratorStackID uint64
}

// use this to create C.DB in two steps, so the pointer lives as long as the calling stack
//   state := buildDBState(kv, counter)
//   db := buildDB(&state, &gasMeter)
//   // then pass db into some FFI function
func buildDBState(kv KVStore, counter uint64) DBState {
	return DBState{
		Store:           kv,
		IteratorStackID: counter,
	}
}

// contract: original pointer/struct referenced must live longer than C.DB struct
// since this is only used internally, we can verify the code that this is the case
func buildDB(state *DBState, gm *GasMeter) C.V3_DB {
	return C.V3_DB{
		gas_meter: (*C.V3_gas_meter_t)(unsafe.Pointer(gm)),
		state:     (*C.V3_db_t)(unsafe.Pointer(state)),
		vtable:    db_vtable,
	}
}

var iterator_vtable = C.V3_Iterator_vtable{
	next_db: (C.next_db_fn)(C.cNextV3_cgo),
}

// contract: original pointer/struct referenced must live longer than C.DB struct
// since this is only used internally, we can verify the code that this is the case
func buildIterator(dbCounter uint64, it dbm.Iterator) C.V3_iterator_t {
	idx := storeIterator(dbCounter, it)
	return C.V3_iterator_t{
		db_counter:     u64(dbCounter),
		iterator_index: u64(idx),
	}
}

//export cGetV3
func cGetV3(ptr *C.V3_db_t, gasMeter *C.V3_gas_meter_t, usedGas *u64, key C.V3_Buffer, val *C.V3_Buffer, errOut *C.V3_Buffer) (ret C.V3_GoResult) {
	defer recoverPanic(&ret)
	if ptr == nil || gasMeter == nil || usedGas == nil || val == nil {
		// we received an invalid pointer
		return C.V3_GoResult_BadArgument
	}

	gm := *(*GasMeter)(unsafe.Pointer(gasMeter))
	kv := *(*KVStore)(unsafe.Pointer(ptr))
	k := receiveSlice(key)

	gasBefore := gm.GasConsumed()
	v := kv.Get(k)
	gasAfter := gm.GasConsumed()
	*usedGas = (u64)(gasAfter - gasBefore)

	// v will equal nil when the key is missing
	// https://github.com/cosmos/cosmos-sdk/blob/1083fa948e347135861f88e07ec76b0314296832/store/types/store.go#L174
	if v != nil {
		*val = allocateRust(v)
	}
	// else: the Buffer on the rust side is initialised as a "null" buffer,
	// so if we don't write a non-null address to it, it will understand that
	// the key it requested does not exist in the kv store

	return C.V3_GoResult_Ok
}

//export cSetV3
func cSetV3(ptr *C.V3_db_t, gasMeter *C.V3_gas_meter_t, usedGas *C.uint64_t, key C.V3_Buffer, val C.V3_Buffer, errOut *C.V3_Buffer) (ret C.V3_GoResult) {
	defer recoverPanic(&ret)
	if ptr == nil || gasMeter == nil || usedGas == nil {
		// we received an invalid pointer
		return C.V3_GoResult_BadArgument
	}

	gm := *(*GasMeter)(unsafe.Pointer(gasMeter))
	kv := *(*KVStore)(unsafe.Pointer(ptr))
	k := receiveSlice(key)
	v := receiveSlice(val)

	gasBefore := gm.GasConsumed()
	kv.Set(k, v)
	gasAfter := gm.GasConsumed()
	*usedGas = (C.uint64_t)(gasAfter - gasBefore)

	return C.V3_GoResult_Ok
}

//export cDeleteV3
func cDeleteV3(ptr *C.V3_db_t, gasMeter *C.V3_gas_meter_t, usedGas *C.uint64_t, key C.V3_Buffer, errOut *C.V3_Buffer) (ret C.V3_GoResult) {
	defer recoverPanic(&ret)
	if ptr == nil || gasMeter == nil || usedGas == nil {
		// we received an invalid pointer
		return C.V3_GoResult_BadArgument
	}

	gm := *(*GasMeter)(unsafe.Pointer(gasMeter))
	kv := *(*KVStore)(unsafe.Pointer(ptr))
	k := receiveSlice(key)

	gasBefore := gm.GasConsumed()
	kv.Delete(k)
	gasAfter := gm.GasConsumed()
	*usedGas = (C.uint64_t)(gasAfter - gasBefore)

	return C.V3_GoResult_Ok
}

//export cScanV3
func cScanV3(ptr *C.V3_db_t, gasMeter *C.V3_gas_meter_t, usedGas *C.uint64_t, start C.V3_Buffer, end C.V3_Buffer, order i32, out *C.V3_GoIter, errOut *C.V3_Buffer) (ret C.V3_GoResult) {
	defer recoverPanic(&ret)
	if ptr == nil || gasMeter == nil || usedGas == nil || out == nil {
		// we received an invalid pointer
		return C.V3_GoResult_BadArgument
	}

	gm := *(*GasMeter)(unsafe.Pointer(gasMeter))
	state := (*DBState)(unsafe.Pointer(ptr))
	kv := state.Store
	// handle null as well as data
	var s, e []byte
	if start.ptr != nil {
		s = receiveSlice(start)
	}
	if end.ptr != nil {
		e = receiveSlice(end)
	}

	var iter dbm.Iterator
	gasBefore := gm.GasConsumed()
	switch order {
	case 1: // Ascending
		iter = kv.Iterator(s, e)
	case 2: // Descending
		iter = kv.ReverseIterator(s, e)
	default:
		return C.V3_GoResult_BadArgument
	}
	gasAfter := gm.GasConsumed()
	*usedGas = (C.uint64_t)(gasAfter - gasBefore)

	out.state = buildIterator(state.IteratorStackID, iter)
	out.vtable = iterator_vtable
	return C.V3_GoResult_Ok
}

//export cNextV3
func cNextV3(ref C.V3_iterator_t, gasMeter *C.V3_gas_meter_t, usedGas *C.uint64_t, key *C.V3_Buffer, val *C.V3_Buffer, errOut *C.V3_Buffer) (ret C.V3_GoResult) {
	// typical usage of iterator
	// 	for ; itr.Valid(); itr.Next() {
	// 		k, v := itr.Key(); itr.Value()
	// 		...
	// 	}

	defer recoverPanic(&ret)
	if ref.db_counter == 0 || gasMeter == nil || usedGas == nil || key == nil || val == nil {
		// we received an invalid pointer
		return C.V3_GoResult_BadArgument
	}

	gm := *(*GasMeter)(unsafe.Pointer(gasMeter))
	iter := retrieveIterator(uint64(ref.db_counter), uint64(ref.iterator_index))
	if !iter.Valid() {
		// end of iterator, return as no-op, nil key is considered end
		return C.V3_GoResult_Ok
	}

	gasBefore := gm.GasConsumed()
	// call Next at the end, upon creation we have first data loaded
	k := iter.Key()
	v := iter.Value()
	// check iter.Error() ????
	iter.Next()
	gasAfter := gm.GasConsumed()
	*usedGas = (C.uint64_t)(gasAfter - gasBefore)

	if k != nil {
		*key = allocateRust(k)
		*val = allocateRust(v)
	}
	return C.V3_GoResult_Ok
}

/***** GoAPI *******/

var api_vtable = C.V3_GoApi_vtable{
	humanize_address:     (C.humanize_address_fn)(C.cHumanAddressV3_cgo),
	canonicalize_address: (C.canonicalize_address_fn)(C.cCanonicalAddressV3_cgo),
}

// contract: original pointer/struct referenced must live longer than C.V3_GoApi struct
// since this is only used internally, we can verify the code that this is the case
func buildAPI(api *GoAPI) C.V3_GoApi {
	return C.V3_GoApi{
		state:  (*C.V3_api_t)(unsafe.Pointer(api)),
		vtable: api_vtable,
	}
}

//export cHumanAddressV3
func cHumanAddressV3(ptr *C.V3_api_t, canon C.V3_Buffer, human *C.V3_Buffer, errOut *C.V3_Buffer, used_gas *u64) (ret C.V3_GoResult) {
	defer recoverPanic(&ret)
	if human == nil {
		// we received an invalid pointer
		return C.V3_GoResult_BadArgument
	}
	api := (*GoAPI)(unsafe.Pointer(ptr))
	c := receiveSlice(canon)
	h, cost, err := api.HumanAddress(c)
	*used_gas = u64(cost)
	if err != nil {
		// store the actual error message in the return buffer
		*errOut = allocateRust([]byte(err.Error()))
		return C.V3_GoResult_User
	}
	if len(h) == 0 {
		panic(fmt.Sprintf("`api.HumanAddress()` returned an empty string for %q", c))
	}
	*human = allocateRust([]byte(h))
	return C.V3_GoResult_Ok
}

//export cCanonicalAddressV3
func cCanonicalAddressV3(ptr *C.V3_api_t, human C.V3_Buffer, canon *C.V3_Buffer, errOut *C.V3_Buffer, used_gas *u64) (ret C.V3_GoResult) {
	defer recoverPanic(&ret)

	if canon == nil {
		// we received an invalid pointer
		return C.V3_GoResult_BadArgument
	}

	api := (*GoAPI)(unsafe.Pointer(ptr))
	h := string(receiveSlice(human))
	c, cost, err := api.CanonicalAddress(h)
	*used_gas = u64(cost)
	if err != nil {
		// store the actual error message in the return buffer
		*errOut = allocateRust([]byte(err.Error()))
		return C.V3_GoResult_User
	}
	if len(c) == 0 {
		panic(fmt.Sprintf("`api.CanonicalAddress()` returned an empty string for %q", h))
	}
	*canon = allocateRust(c)

	// If we do not set canon to a meaningful value, then the other side will interpret that as an empty result.
	return C.V3_GoResult_Ok
}

/****** Go Querier ********/

var querier_vtable = C.V3_Querier_vtable{
	query_external: (C.query_external_fn)(C.cQueryExternalV3_cgo),
}

// contract: original pointer/struct referenced must live longer than C.V3_GoQuerier struct
// since this is only used internally, we can verify the code that this is the case
func buildQuerier(q *Querier) C.V3_GoQuerier {
	return C.V3_GoQuerier{
		state:  (*C.V3_querier_t)(unsafe.Pointer(q)),
		vtable: querier_vtable,
	}
}

//export cQueryExternalV3
func cQueryExternalV3(ptr *C.V3_querier_t, gasLimit C.uint64_t, usedGas *C.uint64_t, request C.V3_Buffer, result *C.V3_Buffer, errOut *C.V3_Buffer) (ret C.V3_GoResult) {
	defer recoverPanic(&ret)
	if ptr == nil || usedGas == nil || result == nil {
		// we received an invalid pointer
		return C.V3_GoResult_BadArgument
	}

	// query the data
	querier := *(*Querier)(unsafe.Pointer(ptr))
	req := receiveSlice(request)

	gasBefore := querier.GasConsumed()
	res := typesv3.RustQuery(querier, req, uint64(gasLimit))
	gasAfter := querier.GasConsumed()
	*usedGas = (C.uint64_t)(gasAfter - gasBefore)

	// serialize the response
	bz, err := json.Marshal(res)
	if err != nil {
		*errOut = allocateRust([]byte(err.Error()))
		return C.V3_GoResult_Other
	}
	*result = allocateRust(bz)
	return C.V3_GoResult_Ok
}
