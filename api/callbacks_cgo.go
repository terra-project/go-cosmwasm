package api

/*
#include "bindings.h"
#include <stdio.h>

// imports (db)
V3_GoResult cSetV3(V3_db_t *ptr, V3_gas_meter_t *gas_meter, uint64_t *used_gas, V3_Buffer key, V3_Buffer val, V3_Buffer *errOut);
V3_GoResult cGetV3(V3_db_t *ptr, V3_gas_meter_t *gas_meter, uint64_t *used_gas, V3_Buffer key, V3_Buffer *val, V3_Buffer *errOut);
V3_GoResult cDeleteV3(V3_db_t *ptr, V3_gas_meter_t *gas_meter, uint64_t *used_gas, V3_Buffer key, V3_Buffer *errOut);
V3_GoResult cScanV3(V3_db_t *ptr, V3_gas_meter_t *gas_meter, uint64_t *used_gas, V3_Buffer start, V3_Buffer end, int32_t order, V3_GoIter *out, V3_Buffer *errOut);
// imports (iterator)
V3_GoResult cNextV3(V3_iterator_t *ptr, V3_gas_meter_t *gas_meter, uint64_t *used_gas, V3_Buffer *key, V3_Buffer *val, V3_Buffer *errOut);
// imports (api)
V3_GoResult cHumanAddressV3(V3_api_t *ptr, V3_Buffer canon, V3_Buffer *human, V3_Buffer *errOut, uint64_t *used_gas);
V3_GoResult cCanonicalAddressV3(V3_api_t *ptr, V3_Buffer human, V3_Buffer *canon, V3_Buffer *errOut, uint64_t *used_gas);
// imports (querier)
V3_GoResult cQueryExternalV3(V3_querier_t *ptr, uint64_t gas_limit, uint64_t *used_gas, V3_Buffer request, V3_Buffer *result, V3_Buffer *errOut);

// Gateway functions (db)
V3_GoResult cGetV3_cgo(V3_db_t *ptr, V3_gas_meter_t *gas_meter, uint64_t *used_gas, V3_Buffer key, V3_Buffer *val, V3_Buffer *errOut) {
	return cGetV3(ptr, gas_meter, used_gas, key, val, errOut);
}
V3_GoResult cSetV3_cgo(V3_db_t *ptr, V3_gas_meter_t *gas_meter, uint64_t *used_gas, V3_Buffer key, V3_Buffer val, V3_Buffer *errOut) {
	return cSetV3(ptr, gas_meter, used_gas, key, val, errOut);
}
V3_GoResult cDeleteV3_cgo(V3_db_t *ptr, V3_gas_meter_t *gas_meter, uint64_t *used_gas, V3_Buffer key, V3_Buffer *errOut) {
	return cDeleteV3(ptr, gas_meter, used_gas, key, errOut);
}
V3_GoResult cScanV3_cgo(V3_db_t *ptr, V3_gas_meter_t *gas_meter, uint64_t *used_gas, V3_Buffer start, V3_Buffer end, int32_t order, V3_GoIter *out, V3_Buffer *errOut) {
	return cScanV3(ptr, gas_meter, used_gas, start, end, order, out, errOut);
}

// Gateway functions (iterator)
V3_GoResult cNextV3_cgo(V3_iterator_t *ptr, V3_gas_meter_t *gas_meter, uint64_t *used_gas, V3_Buffer *key, V3_Buffer *val, V3_Buffer *errOut) {
	return cNextV3(ptr, gas_meter, used_gas, key, val, errOut);
}

// Gateway functions (api)
V3_GoResult cCanonicalAddressV3_cgo(V3_api_t *ptr, V3_Buffer human, V3_Buffer *canon, V3_Buffer *errOut, uint64_t *used_gas) {
    return cCanonicalAddressV3(ptr, human, canon, errOut, used_gas);
}
V3_GoResult cHumanAddressV3_cgo(V3_api_t *ptr, V3_Buffer canon, V3_Buffer *human, V3_Buffer *errOut, uint64_t *used_gas) {
    return cHumanAddressV3(ptr, canon, human, errOut, used_gas);
}

// Gateway functions (querier)
V3_GoResult cQueryExternalV3_cgo(V3_querier_t *ptr, uint64_t gas_limit, uint64_t *used_gas, V3_Buffer request, V3_Buffer *result, V3_Buffer *errOut) {
    return cQueryExternalV3(ptr, gas_limit, used_gas, request, result, errOut);
}
*/
import "C"

// We need these gateway functions to allow calling back to a go function from the c code.
// At least I didn't discover a cleaner way.
// Also, this needs to be in a different file than `callbacks.go`, as we cannot create functions
// in the same file that has //export directives. Only import header types
