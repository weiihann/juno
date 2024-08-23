package core

import (
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/NethermindEth/juno/core/crypto"
	"github.com/NethermindEth/juno/core/felt"
)

var _ Class = (*Cairo1Class)(nil)

// Class unambiguously defines a [Contract]'s semantics.
type Class interface {
	Version() uint64
	Hash() (*felt.Felt, error)
	CompareClashHash(felt.Felt) error
}

// EntryPoint uniquely identifies a Cairo function to execute.
type EntryPoint struct {
	// starknet_keccak hash of the function signature.
	Selector *felt.Felt
	// The offset of the instruction in the class's bytecode.
	Offset *felt.Felt
}

// Cairo1Class unambiguously defines a [Contract]'s semantics.
type Cairo1Class struct {
	Abi         string
	AbiHash     *felt.Felt
	EntryPoints struct {
		Constructor []SierraEntryPoint
		External    []SierraEntryPoint
		L1Handler   []SierraEntryPoint
	}
	Program         []*felt.Felt
	ProgramHash     *felt.Felt
	SemanticVersion string
	Compiled        *CompiledClass
}

type SegmentLengths struct {
	Children []SegmentLengths
	Length   uint64
}

type CompiledClass struct {
	Bytecode               []*felt.Felt
	PythonicHints          json.RawMessage
	CompilerVersion        string
	Hints                  json.RawMessage
	Prime                  *big.Int
	External               []CompiledEntryPoint
	L1Handler              []CompiledEntryPoint
	Constructor            []CompiledEntryPoint
	BytecodeSegmentLengths SegmentLengths
}

type CompiledEntryPoint struct {
	Offset   uint64
	Builtins []string
	Selector *felt.Felt
}

type SierraEntryPoint struct {
	Index    uint64
	Selector *felt.Felt
}

func (c *Cairo1Class) Version() uint64 {
	return 1
}

func (c *Cairo1Class) Hash() (*felt.Felt, error) {
	return crypto.PoseidonArray(
		new(felt.Felt).SetBytes([]byte("CONTRACT_CLASS_V"+c.SemanticVersion)),
		crypto.PoseidonArray(flattenSierraEntryPoints(c.EntryPoints.External)...),
		crypto.PoseidonArray(flattenSierraEntryPoints(c.EntryPoints.L1Handler)...),
		crypto.PoseidonArray(flattenSierraEntryPoints(c.EntryPoints.Constructor)...),
		c.AbiHash,
		c.ProgramHash,
	), nil
}

func (c *Cairo1Class) CompareClashHash(cHash felt.Felt) error {
	hash, err := c.Hash()
	if err != nil {
		return err
	}

	if !hash.Equal(&cHash) {
		return fmt.Errorf("cannot verify class hash: calculated hash %v, received hash %v", hash.String(), cHash.String())
	}

	return nil
}

var compiledClassV1Prefix = new(felt.Felt).SetBytes([]byte("COMPILED_CLASS_V1"))

func (c *CompiledClass) Hash() *felt.Felt {
	var bytecodeHash *felt.Felt
	if len(c.BytecodeSegmentLengths.Children) == 0 {
		bytecodeHash = crypto.PoseidonArray(c.Bytecode...)
	} else {
		bytecodeHash = SegmentedBytecodeHash(c.Bytecode, c.BytecodeSegmentLengths.Children)
	}

	return crypto.PoseidonArray(
		compiledClassV1Prefix,
		crypto.PoseidonArray(flattenCompiledEntryPoints(c.External)...),
		crypto.PoseidonArray(flattenCompiledEntryPoints(c.L1Handler)...),
		crypto.PoseidonArray(flattenCompiledEntryPoints(c.Constructor)...),
		bytecodeHash,
	)
}

func SegmentedBytecodeHash(bytecode []*felt.Felt, segmentLengths []SegmentLengths) *felt.Felt {
	var startingOffset uint64
	var digestSegment func(segments []SegmentLengths) (uint64, *felt.Felt)
	digestSegment = func(segments []SegmentLengths) (uint64, *felt.Felt) {
		var totalLength uint64
		var digest crypto.PoseidonDigest
		for _, segment := range segments {
			var curSegmentLength uint64
			var curSegmentHash *felt.Felt

			if len(segment.Children) == 0 {
				curSegmentLength = segment.Length
				segmentBytecode := bytecode[startingOffset : startingOffset+segment.Length]
				curSegmentHash = crypto.PoseidonArray(segmentBytecode...)
			} else {
				curSegmentLength, curSegmentHash = digestSegment(segment.Children)
			}

			digest.Update(new(felt.Felt).SetUint64(curSegmentLength))
			digest.Update(curSegmentHash)

			startingOffset += curSegmentLength
			totalLength += curSegmentLength
		}
		digestRes := digest.Finish()
		return totalLength, digestRes.Add(digestRes, new(felt.Felt).SetUint64(1))
	}

	_, hash := digestSegment(segmentLengths)
	return hash
}

func flattenSierraEntryPoints(entryPoints []SierraEntryPoint) []*felt.Felt {
	result := make([]*felt.Felt, len(entryPoints)*2)
	for i, entryPoint := range entryPoints {
		// It is important that Selector is first because the order
		// influences the class hash.
		result[2*i] = entryPoint.Selector
		result[2*i+1] = new(felt.Felt).SetUint64(entryPoint.Index)
	}
	return result
}

func flattenCompiledEntryPoints(entryPoints []CompiledEntryPoint) []*felt.Felt {
	result := make([]*felt.Felt, len(entryPoints)*3)
	for i, entryPoint := range entryPoints {
		// It is important that Selector is first, then Offset is second because the order
		// influences the class hash.
		result[3*i] = entryPoint.Selector
		result[3*i+1] = new(felt.Felt).SetUint64(entryPoint.Offset)
		builtins := make([]*felt.Felt, len(entryPoint.Builtins))
		for idx, buil := range entryPoint.Builtins {
			builtins[idx] = new(felt.Felt).SetBytes([]byte(buil))
		}
		result[3*i+2] = crypto.PoseidonArray(builtins...)
	}

	return result
}

func VerifyClassHashes(classes map[felt.Felt]Class) error {
	for hash, class := range classes {
		err := class.CompareClashHash(hash)
		if err != nil {
			return err
		}
	}

	return nil
}
