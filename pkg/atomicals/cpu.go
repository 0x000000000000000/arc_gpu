//go:build cpu
// +build cpu

package atomicals

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"runtime"

	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
)

func mine(i int, input Input, resultCh chan<- Result, rawTx string) {
	// set different time for each goroutine
	input.CopiedData.Args.Time += uint32(i)
	// use uint32 so we can avoid cbor encoding at runtime
	input.CopiedData.Args.Nonce = uint32(^uint16(0)) + 1
	input.Init()

	result := MineCommitTx(i, &input, rawTx)
	revealTxHash, embed := MineRevealTx(&input, result)
	result.RevealTxhash = &revealTxHash
	result.Embed = string(embed)
	resultCh <- *result
}

func MineCommitTx(i int, input *Input, rawTx string) *Result {

	// msgTx := wire.NewMsgTx(wire.TxVersion)
	// output := wire.NewOutPoint(input.FundingUtxo.Txid, input.FundingUtxo.Index)
	// txIn := wire.NewTxIn(output, nil, nil)
	// txIn.Sequence = 0
	// msgTx.AddTxIn(txIn)

	// scriptP2TR := input.MustBuildScriptP2TR()
	// txOut := wire.NewTxOut(int64(input.Fees.RevealFeePlusOutputs), scriptP2TR.Output)
	// msgTx.AddTxOut(txOut)
	// // add change utxo
	// if change := input.GetCommitChange(); change != 0 {
	// 	msgTx.AddTxOut(wire.NewTxOut(change, input.KeyPairInfo.Ouput))
	// }
	// pkscript, _ := hex.DecodeString("5120b0cc121a1e5b6c2f2ea18ed079e0fd25f700490ad048c0dc9e24671f1a6a5ea6")
	// txOut := wire.NewTxOut(99673911, pkscript)
	// tid := new(chainhash.Hash)
	// b, _ := hex.DecodeString("aabb3abf51b8e9cbcda15b92573a7c18f1e93871523ffa36410173d35e5d00cd")
	// tid.SetBytes(b)
	// output := wire.NewOutPoint(tid, 1)
	// txIn := wire.NewTxIn(output, nil, nil)
	// txIn.Sequence = 0
	msgTx := new(wire.MsgTx)
	by, err := hex.DecodeString(rawTx)
	if err != nil {
		fmt.Println("err0", err)
		return nil
	}
	err = msgTx.Deserialize(bytes.NewBuffer(by))
	if err != nil {
		fmt.Println("err", err)
		return nil
	}

	txIn := msgTx.TxIn[0]
	txOut := msgTx.TxOut[0]
	buf := bytes.NewBuffer(make([]byte, 0, msgTx.SerializeSizeStripped()))
	msgTx.SerializeNoWitness(buf)

	serializedTx := buf.Bytes()
	var hash chainhash.Hash
	fmt.Println("Sequence:", msgTx.TxIn[0].Sequence)
	for {
		hash = chainhash.DoubleHashH(serializedTx)
		if input.WorkerBitworkInfoCommit.HasValidBitwork(&hash) {
			break
		}
		txIn.Sequence++
		binary.LittleEndian.PutUint32(serializedTx[42:], txIn.Sequence)
		if txIn.Sequence == MAX_SEQUENCE {
			input.CopiedData.Args.Nonce++
			scriptP2TR := input.MustBuildScriptP2TR()
			txOut.PkScript = scriptP2TR.Output
			txIn.Sequence = 0
			buf := bytes.NewBuffer(make([]byte, 0, msgTx.SerializeSizeStripped()))
			msgTx.SerializeNoWitness(buf)
			serializedTx = buf.Bytes()

		}
	}

	fmt.Println(hex.EncodeToString(serializedTx))
	return &Result{
		FinalCopyData: input.CopiedData,
		FinalSequence: txIn.Sequence,
		CommitTxHash:  &hash,
	}
}

func Mine(input Input, threads int32, result chan<- Result, rawTx string) {
	for i := 0; i < runtime.NumCPU(); i++ {
		go mine(i, input, result, rawTx)
	}
}
