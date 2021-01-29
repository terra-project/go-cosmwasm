package typesv3

import (
	"encoding/json"
	"testing"

	"github.com/CosmWasm/wasmvm/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDelegationWithEmptyArray(t *testing.T) {
	var del types.Delegations
	bz, err := json.Marshal(&del)
	require.NoError(t, err)
	assert.Equal(t, string(bz), `[]`)

	var redel types.Delegations
	err = json.Unmarshal(bz, &redel)
	require.NoError(t, err)
	assert.Nil(t, redel)
}

func TestDelegationWithData(t *testing.T) {
	del := types.Delegations{{
		Validator: "foo",
		Delegator: "bar",
		Amount:    types.NewCoin(123, "stake"),
	}}
	bz, err := json.Marshal(&del)
	require.NoError(t, err)

	var redel types.Delegations
	err = json.Unmarshal(bz, &redel)
	require.NoError(t, err)
	assert.Equal(t, redel, del)
}

func TestValidatorWithEmptyArray(t *testing.T) {
	var val types.Validators
	bz, err := json.Marshal(&val)
	require.NoError(t, err)
	assert.Equal(t, string(bz), `[]`)

	var reval types.Validators
	err = json.Unmarshal(bz, &reval)
	require.NoError(t, err)
	assert.Nil(t, reval)
}

func TestValidatorWithData(t *testing.T) {
	val := types.Validators{{
		Address:       "1234567890",
		Commission:    "0.05",
		MaxCommission: "0.1",
		MaxChangeRate: "0.02",
	}}
	bz, err := json.Marshal(&val)
	require.NoError(t, err)

	var reval types.Validators
	err = json.Unmarshal(bz, &reval)
	require.NoError(t, err)
	assert.Equal(t, reval, val)
}
