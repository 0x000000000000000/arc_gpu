//go:build cuda
// +build cuda

package atomicals

// #include <stdint.h>
//uint32_t scanhash_sha256d(int thr_id, unsigned char* in, unsigned int inlen, unsigned char *target, unsigned int target_len, char pp, char ext, unsigned int threads, unsigned int start_seq, unsigned int *hashes_done);
//#cgo LDFLAGS: -L. -L../../cuda -lhash
import "C"
import (
	"bytes"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/btcsuite/btcd/wire"
)

// go build -ldflags="-L/home/mask/project/atomicals-cuda/web3-go/cuda" main.go
func Mine(input Input, threads uint32, result chan<- Result) {
	deviceNum := 1
	devcieNumStr := os.Getenv("CUDA_DEVICE_NUM")
	if devcieNumStr != "" {
		deviceNum = int(devcieNumStr[0] - '0')
	}
	for i := 0; i < deviceNum; i++ {
		go mine(i, input, threads, result)
	}
}

func mine(i int, input Input, threads uint32, result chan<- Result) {
	// set different time for each goroutine
	// input.CopiedData.Args.Time += uint32(i)
	// // use uint32 so we can avoid cbor encoding at runtime
	// input.CopiedData.Args.Nonce = uint32(^uint16(0)) + 1
	// input.Init()

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

	msgTx := new(wire.MsgTx)
	by, err := hex.DecodeString("01000000011a3059f8c406e31a936cd07d7522c34ad2eb5035ee2eae65f039aa43da08bbaa010000000000000000029709000000000000225120a579cc302987845e990343cc5a60bfffdf290b45d011b16d9269cc86349385d497d7f00500000000225120b0cc121a1e5b6c2f2ea18ed079e0fd25f700490ad048c0dc9e24671f1a6a5ea600000000")
	if err != nil {
		fmt.Println("err0", err)
		return
	}

	err = msgTx.Deserialize(bytes.NewBuffer(by))
	if err != nil {
		fmt.Println("err", err)
		return
	}
	buf := bytes.NewBuffer(make([]byte, 0, msgTx.SerializeSizeStripped()))
	txIn := msgTx.TxIn[0]
	txOut := msgTx.TxOut[0]
	msgTx.SerializeNoWitness(buf)
	serializedTx := buf.Bytes()

	hashesDone := C.uint(0)
	var (
		pp  = -1
		ext = -1
	)
	if input.WorkerBitworkInfoCommit.PrefixPartial != nil {
		pp = int(*input.WorkerBitworkInfoCommit.PrefixPartial)
	}
	if input.WorkerBitworkInfoCommit.Ext != 0 {
		ext = int(input.WorkerBitworkInfoCommit.Ext)
	}
	for {
		start := time.Now()
		seq := C.scanhash_sha256d(
			C.int(i), // device id
			(*C.uchar)(&serializedTx[0]),
			C.uint(len(serializedTx)),
			(*C.uchar)(&input.WorkerBitworkInfoCommit.PrefixBytes[0]),
			C.uint(len(input.WorkerBitworkInfoCommit.PrefixBytes)),
			C.char(pp),
			C.char(ext),
			C.uint(1<<threads),
			C.uint(txIn.Sequence),
			&hashesDone,
		)
		log.Printf("device: %d, hashrate: %d/s", i, int64(float64(hashesDone)/time.Since(start).Seconds()))
		if uint32(seq) != MAX_SEQUENCE {
			txIn.Sequence = uint32(seq)
			break
		}

		input.CopiedData.Args.Nonce++
		scriptP2TR := input.MustBuildScriptP2TR()
		txOut.PkScript = scriptP2TR.Output
		txIn.Sequence = 0
		buf := bytes.NewBuffer(make([]byte, 0, msgTx.SerializeSizeStripped()))
		msgTx.SerializeNoWitness(buf)
		serializedTx = buf.Bytes()
	}

	PrintMsgTx(msgTx)
	result <- Result{
		FinalCopyData: input.CopiedData,
		FinalSequence: txIn.Sequence,
	}
}
