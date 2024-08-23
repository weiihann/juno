package cairo0

//#include <stdint.h>
//#include <stdlib.h>
//#include <stddef.h>
//
// extern void Cairo0ClassHash(char* class_json_str, char* hash);
// #cgo vm_debug  LDFLAGS: -L./rust/target/debug   -ljuno_starknet_core_rs
// #cgo !vm_debug LDFLAGS: -L./rust/target/release -ljuno_starknet_core_rs
import "C"

import (
	"encoding/json"
	"errors"
	"unsafe"

	"github.com/NethermindEth/juno/core"
	"github.com/NethermindEth/juno/core/felt"
	"github.com/NethermindEth/juno/starknet"
	"github.com/NethermindEth/juno/utils"
)

var _ core.Class = (*Cairo0Class)(nil)

// Cairo0Class unambiguously defines a [Contract]'s semantics.
type Cairo0Class struct {
	Abi json.RawMessage
	// External functions defined in the class.
	Externals []core.EntryPoint
	// Functions that receive L1 messages. See
	// https://www.cairo-lang.org/docs/hello_starknet/l1l2.html#receiving-a-message-from-l1
	L1Handlers []core.EntryPoint
	// Constructors for the class. Currently, only one is allowed.
	Constructors []core.EntryPoint
	// Base64 encoding of compressed Program
	Program string
}

func (c *Cairo0Class) Version() uint64 {
	return 0
}

func (c *Cairo0Class) Hash() (*felt.Felt, error) {
	return cairo0ClassHash(c)
}

func (c *Cairo0Class) CompareClashHash(hash felt.Felt) error { return nil }

func cairo0ClassHash(class *Cairo0Class) (*felt.Felt, error) {
	definition, err := makeDeprecatedVMClass(class)
	if err != nil {
		return nil, err
	}

	classJSON, err := json.Marshal(definition)
	if err != nil {
		return nil, err
	}
	classJSONCStr := cstring(classJSON)

	var hash felt.Felt
	hashBytes := hash.Bytes()

	C.Cairo0ClassHash(classJSONCStr, (*C.char)(unsafe.Pointer(&hashBytes[0])))
	hash.SetBytes(hashBytes[:])
	C.free(unsafe.Pointer(classJSONCStr))
	if hash.IsZero() {
		return nil, errors.New("failed to calculate class hash")
	}
	return &hash, nil
}

func makeDeprecatedVMClass(class *Cairo0Class) (*starknet.Cairo0Definition, error) {
	adaptEntryPoint := func(ep core.EntryPoint) starknet.EntryPoint {
		return starknet.EntryPoint{
			Selector: ep.Selector,
			Offset:   ep.Offset,
		}
	}

	constructors := utils.Map(utils.NonNilSlice(class.Constructors), adaptEntryPoint)
	external := utils.Map(utils.NonNilSlice(class.Externals), adaptEntryPoint)
	handlers := utils.Map(utils.NonNilSlice(class.L1Handlers), adaptEntryPoint)

	decompressedProgram, err := utils.Gzip64Decode(class.Program)
	if err != nil {
		return nil, err
	}

	return &starknet.Cairo0Definition{
		Program: decompressedProgram,
		Abi:     class.Abi,
		EntryPoints: starknet.EntryPoints{
			Constructor: constructors,
			External:    external,
			L1Handler:   handlers,
		},
	}, nil
}

// cstring creates a null-terminated C string from the given byte slice.
// the caller is responsible for freeing the underlying memory
func cstring(data []byte) *C.char {
	str := unsafe.String(unsafe.SliceData(data), len(data))
	return C.CString(str)
}
