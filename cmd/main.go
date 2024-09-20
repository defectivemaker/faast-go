package main

import (
	"fmt"
	"log"
	"os"
	"sync"

	"faast-go/internal/config"
	"faast-go/internal/curl"
	"faast-go/internal/permute"
	"faast-go/internal/worker"

	"github.com/schollz/progressbar/v3"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Please provide a YAML config file")
	}

	loadedConfig, err := config.LoadConfig(os.Args[1])
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	curlConfig, err := curl.NewCurlConfig(loadedConfig)
	if err != nil {
		log.Fatalf("Error creating curl config: %v", err)
	}

	wordlists, err := config.LoadWordlists(loadedConfig.Wordlists)
	if err != nil {
		log.Fatalf("Error loading wordlists: %v", err)
	}

	shardedLists := permute.ShardLists(wordlists, loadedConfig.ShardIndex, loadedConfig.NumShards)

	totalPermutations := permute.CalculateTotalPermutations(shardedLists)
	progressBar := progressbar.Default(int64(totalPermutations))

	permChan := make(chan []string, 10000)
	resultChan := make(chan worker.CurlResult, 1000)

	var wg sync.WaitGroup
	for _, list := range shardedLists {
		wg.Add(1)
		go func(l [][]string) {
			defer wg.Done()
			permute.IteratePermutations(permute.NewPermutationIterator(l), permChan)
		}(list)
	}

	go func() {
		wg.Wait()
		close(permChan)
	}()

	workerPool := worker.NewWorkerPool(curlConfig, permChan, resultChan, progressBar)
	workerPool.Start()

	go func() {
		workerPool.Wait()
		close(resultChan)
	}()

	ProcessResults(resultChan, curlConfig)
}

func ProcessResults(resultChan <-chan worker.CurlResult, loadedConfig *curl.CurlConfig) {
	for result := range resultChan {
		if result.Err != nil {
			fmt.Printf("Error: %v\n", result.Err)
			continue
		}
		if !loadedConfig.ValidateResponse(result.Response) {
			fmt.Printf("Payload %v caused an anomaly\n", result.Payload)
			result.Response.Body.Close()
		}
	}
}
