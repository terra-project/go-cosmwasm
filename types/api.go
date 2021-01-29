package types

//---------- Api ---------
type HumanizeAddress func([]byte) (string, uint64, error)
type CanonicalizeAddress func(string) ([]byte, uint64, error)

type GoAPI struct {
	HumanAddress     HumanizeAddress
	CanonicalAddress CanonicalizeAddress
}
