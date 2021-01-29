// +build linux,muslc

package apiv3

// #cgo LDFLAGS: -Wl,-rpath,${SRCDIR} -L${SRCDIR} -lgo_cosmwasm_muslc
import "C"
