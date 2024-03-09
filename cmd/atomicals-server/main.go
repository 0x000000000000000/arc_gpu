package main

import (
	"encoding/json"
	"fmt"
	"go-atomicals/pkg/atomicals"
	"log"
	"net/http"
	"time"
)

type PowRequst struct {
	Prefix string `json:"prefix"`
	Ext    uint8  `json:"ext"`
	RawTx  string `json:"rawTx"`
}

func main() {
	//curl http://localhost:9900/mine -d '{"Prefix": "aabb", ext:0, "rawTx": "0100000001d55dd0a41aea151cee5066e89ab7c7df66827c9d6c1b615485ad0143192bbbaa01000000000000000002970900000000000022512084fb43da5caee50dbf7d1ea81d2a6f0568d5ad2d7b62a147df1470126afb1126c6a8f00500000000225120b0cc121a1e5b6c2f2ea18ed079e0fd25f700490ad048c0dc9e24671f1a6a5ea600000000"}'
	http.HandleFunc("/mine", func(w http.ResponseWriter, r *http.Request) {

		var input atomicals.Input
		// data, _ := os.Open("./data.json")
		// dec := json.NewDecoder(data)
		// if dec.Decode(&input) != nil {
		// 	log.Fatalf("decode input error")
		// }
		start := time.Now()
		// //reporter := hashrate.NewReporter()
		// // core count
		result := make(chan atomicals.Result, 1)

		// var req PowRequst
		dec2 := json.NewDecoder(r.Body)
		if dec2.Decode(&input) != nil {
			log.Fatalf("decode input error")
		}

		go atomicals.Mine(input, 31, result, "")
		// go atomicals.Mine(input, result)
		finalData := <-result
		log.Printf("found solution cost: %v", time.Since(start))

		enc := json.NewEncoder(w)
		if err := enc.Encode(finalData); err != nil {
			log.Fatalf("encode output error")
		}

	})

	// start a http server
	fmt.Println("listening....")
	http.ListenAndServe("0.0.0.0:9900", nil)

}
