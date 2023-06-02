package vm

//#include <stdint.h>
//#include <stdlib.h>
// extern void cairoVMCall(char* contract_address, char* entry_point_selector, char** calldata, uintptr_t len_calldata,
//					uintptr_t readerHandle, unsigned long long block_number, unsigned long long block_timestamp,
//					char* chain_id);
//
// extern void cairoVMExecute(char* txn_json, char* class_json, uintptr_t readerHandle, unsigned long long block_number,
//					unsigned long long block_timestamp, char* chain_id);
//
// #cgo LDFLAGS: -lm -lssl -lcrypto -L./juno-starknet-rs/target/release -ljuno_starknet_rs
import "C"

import (
	"errors"
	"runtime/cgo"
	"unsafe"

	"github.com/NethermindEth/juno/core"
	"github.com/NethermindEth/juno/core/felt"
	"github.com/NethermindEth/juno/utils"
)

type callContext struct {
	state       core.StateReader
	err         string
	response    []*felt.Felt
	gasConsumed *felt.Felt
}

func unwrapContext(readerHandle C.uintptr_t) *callContext {
	context, ok := cgo.Handle(readerHandle).Value().(*callContext)
	if !ok {
		panic("cannot cast reader")
	}

	return context
}

//export JunoReportError
func JunoReportError(readerHandle C.uintptr_t, str *C.char) {
	context := unwrapContext(readerHandle)
	context.err = C.GoString(str)
}

//export JunoAppendResponse
func JunoAppendResponse(readerHandle C.uintptr_t, ptr unsafe.Pointer) {
	context := unwrapContext(readerHandle)
	context.response = append(context.response, makeFeltFromPtr(ptr))
}

//export JunoSetGasConsumed
func JunoSetGasConsumed(readerHandle C.uintptr_t, ptr unsafe.Pointer) {
	context := unwrapContext(readerHandle)
	context.gasConsumed = makeFeltFromPtr(ptr)
}

func makeFeltFromPtr(ptr unsafe.Pointer) *felt.Felt {
	return new(felt.Felt).SetBytes(C.GoBytes(ptr, felt.Bytes))
}

func makePtrFromFelt(val *felt.Felt) unsafe.Pointer {
	feltBytes := val.Bytes()
	return C.CBytes(feltBytes[:])
}

func Call(contractAddr, selector *felt.Felt, calldata []*felt.Felt, blockNumber,
	blockTimestamp uint64, state core.StateReader, network utils.Network,
) ([]*felt.Felt, error) {
	context := &callContext{
		state: state,
	}
	handle := cgo.NewHandle(context)
	defer handle.Delete()

	addrBytes := contractAddr.Bytes()
	selectorBytes := selector.Bytes()
	calldataPtrs := []*C.char{}
	for _, data := range calldata {
		bytes := data.Bytes()
		calldataPtrs = append(calldataPtrs, (*C.char)(C.CBytes(bytes[:])))
	}
	calldataArrPtr := unsafe.Pointer(nil)
	if len(calldataPtrs) > 0 {
		calldataArrPtr = unsafe.Pointer(&calldataPtrs[0])
	}

	chainID := C.CString(network.ChainIDString())
	C.cairoVMCall((*C.char)(unsafe.Pointer(&addrBytes[0])),
		(*C.char)(unsafe.Pointer(&selectorBytes[0])),
		(**C.char)(calldataArrPtr), C.ulong(len(calldataPtrs)),
		C.ulong(handle), C.ulonglong(blockNumber), C.ulonglong(blockTimestamp),
		chainID)

	for _, ptr := range calldataPtrs {
		C.free(unsafe.Pointer(ptr))
	}
	C.free(unsafe.Pointer(chainID))

	if len(context.err) > 0 {
		return nil, errors.New(context.err)
	}
	return context.response, nil
}

// Execute executes a given transaction and returns the gas spent
func Execute(txn core.Transaction, declaredClass core.Class, blockNumber, blockTimestamp uint64,
	state core.StateReader, network utils.Network,
) (*felt.Felt, error) {
	context := &callContext{
		state: state,
	}
	handle := cgo.NewHandle(context)
	defer handle.Delete()

	txnJson, err := marshalTxn(txn, true)
	if err != nil {
		return nil, err
	}

	var declaredClassJsonCstr *C.char
	if declaredClass != nil {
		declaredClassJson, cErr := marshalDeclaredClass(declaredClass)
		if cErr != nil {
			return nil, cErr
		}
		declaredClassJsonCstr = C.CString(string(declaredClassJson))
	}

	txnJsonCstr := C.CString(string(txnJson))
	chainID := C.CString(network.ChainIDString())
	C.cairoVMExecute(txnJsonCstr, declaredClassJsonCstr, C.ulong(handle),
		C.ulonglong(blockNumber), C.ulonglong(blockTimestamp), chainID)

	C.free(unsafe.Pointer(declaredClassJsonCstr))
	C.free(unsafe.Pointer(txnJsonCstr))
	C.free(unsafe.Pointer(chainID))

	if len(context.err) > 0 {
		return nil, errors.New(context.err)
	}
	return context.gasConsumed, nil
}
