package vm

//#include <stdint.h>
//#include <stdlib.h>
import "C"

import (
	"unsafe"
)

//export JunoFree
func JunoFree(ptr unsafe.Pointer) {
	C.free(ptr)
}

//export JunoStateGetStorageAt
func JunoStateGetStorageAt(readerHandle C.uintptr_t, contractAddress, storageLocation unsafe.Pointer) unsafe.Pointer {
	context := unwrapContext(readerHandle)

	contractAddressFelt := makeFeltFromPtr(contractAddress)
	storageLocationFelt := makeFeltFromPtr(storageLocation)
	val, err := context.state.ContractStorage(contractAddressFelt, storageLocationFelt)
	if err != nil {
		return nil
	}

	return makePtrFromFelt(val)
}

//export JunoStateGetNonceAt
func JunoStateGetNonceAt(readerHandle C.uintptr_t, contractAddress unsafe.Pointer) unsafe.Pointer {
	context := unwrapContext(readerHandle)

	contractAddressFelt := makeFeltFromPtr(contractAddress)
	val, err := context.state.ContractNonce(contractAddressFelt)
	if err != nil {
		return nil
	}

	return makePtrFromFelt(val)
}

//export JunoStateGetClassHashAt
func JunoStateGetClassHashAt(readerHandle C.uintptr_t, contractAddress unsafe.Pointer) unsafe.Pointer {
	context := unwrapContext(readerHandle)

	contractAddressFelt := makeFeltFromPtr(contractAddress)
	val, err := context.state.ContractClassHash(contractAddressFelt)
	if err != nil {
		return nil
	}

	return makePtrFromFelt(val)
}

//export JunoStateGetClass
func JunoStateGetClass(readerHandle C.uintptr_t, classHash unsafe.Pointer) unsafe.Pointer {
	context := unwrapContext(readerHandle)

	classHashFelt := makeFeltFromPtr(classHash)
	val, err := context.state.Class(classHashFelt)
	if err != nil {
		return nil
	}

	compiledClass, err := marshalCompiledClass(val.Class)
	if err != nil {
		return nil
	}

	return unsafe.Pointer(C.CString(string(compiledClass)))
}
