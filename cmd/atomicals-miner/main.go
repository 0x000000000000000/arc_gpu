package main

import (
	"encoding/json"
	"go-atomicals/pkg/atomicals"
	"log"
	"os"
	"time"
)

func main() {
	// read from stdin, pase it to input
	var input atomicals.Input
	data, _ := os.Open("./data.json")
	dec := json.NewDecoder(data)
	if dec.Decode(&input) != nil {
		log.Fatalf("decode input error")
	}

	start := time.Now()
	// reporter := hashrate.NewReporter()
	// core count
	result := make(chan atomicals.Result, 1)

	go atomicals.Mine(input, 1, result, "")
	finalData := <-result
	log.Printf("found solution cost: %v", time.Since(start))

	enc := json.NewEncoder(os.Stdout)
	if err := enc.Encode(finalData); err != nil {
		log.Fatalf("encode output error")
	}
}
