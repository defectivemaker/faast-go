package worker

import (
	"faast-go/internal/config"
	"faast-go/internal/curl"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync/atomic"
	"testing"
	"time"

	"github.com/schollz/progressbar/v3"
)

func createTestYamlConfig() *config.YamlConfig {
	return &config.YamlConfig{
		Type:         "payload",
		Endpoint:     "http://example.com",
		Fields:       []string{"field1", "field2"},
		Wordlists:    []string{"wordlist1.txt"},
		StaticValues: []string{"staticValue"},
		Cookies:      []string{"session=abc123"},
		ValidateType: "code",
		SizeDefault:  100,
		CodeDefault:  404,
		RateLimit:    10,
		Timeout:      5,
		ShardIndex:   0,
		NumShards:    1,
	}
}

func TestNewWorkerPool(t *testing.T) {
	yamlConfig := createTestYamlConfig()
	curlConfig, err := curl.NewCurlConfig(yamlConfig)
	if err != nil {
		t.Fatalf("Failed to create CurlConfig: %v", err)
	}

	permChan := make(chan []string)
	resultChan := make(chan CurlResult)
	progressBar := progressbar.New(100)

	wp := NewWorkerPool(curlConfig, permChan, resultChan, progressBar)

	if wp.config != curlConfig {
		t.Errorf("Expected config to be %v, got %v", curlConfig, wp.config)
	}
	if wp.permChan != permChan {
		t.Errorf("Expected permChan to be %v, got %v", permChan, wp.permChan)
	}
	if wp.resultChan != resultChan {
		t.Errorf("Expected resultChan to be %v, got %v", resultChan, wp.resultChan)
	}
	if wp.progressBar != progressBar {
		t.Errorf("Expected progressBar to be %v, got %v", progressBar, wp.progressBar)
	}
	if wp.numWorkers != 10 {
		t.Errorf("Expected numWorkers to be 10, got %d", wp.numWorkers)
	}
}

func TestWorkerPool_Start(t *testing.T) {
	yamlConfig := createTestYamlConfig()
	curlConfig, _ := curl.NewCurlConfig(yamlConfig)
	permChan := make(chan []string)
	resultChan := make(chan CurlResult)
	progressBar := progressbar.New(100)

	wp := NewWorkerPool(curlConfig, permChan, resultChan, progressBar)
	wp.numWorkers = 3 // Reduce number of workers for testing

	wp.Start()

	// Check if the correct number of workers were started
	time.Sleep(100 * time.Millisecond) // Give some time for goroutines to start
	if int(atomic.LoadInt32(&wp.workerCount)) != wp.numWorkers {
		t.Errorf("Expected %d workers to be started, got %d", wp.numWorkers, atomic.LoadInt32(&wp.workerCount))
	}

	close(permChan) // Cause workers to finish
	wp.Wait()
}

func TestWorkerPool_Wait(t *testing.T) {
	yamlConfig := createTestYamlConfig()
	curlConfig, _ := curl.NewCurlConfig(yamlConfig)
	permChan := make(chan []string)
	resultChan := make(chan CurlResult)
	progressBar := progressbar.New(100)

	wp := NewWorkerPool(curlConfig, permChan, resultChan, progressBar)
	wp.numWorkers = 3 // Reduce number of workers for testing

	wp.Start()
	close(permChan) // This will cause workers to finish

	done := make(chan struct{})
	go func() {
		wp.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Wait completed successfully
	case <-time.After(2 * time.Second):
		t.Error("Wait timed out")
	}
}

func TestWorkerPool_worker(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	yamlConfig := createTestYamlConfig()
	yamlConfig.Endpoint = server.URL
	curlConfig, _ := curl.NewCurlConfig(yamlConfig)

	permChan := make(chan []string)
	resultChan := make(chan CurlResult)
	progressBar := progressbar.New(2)

	wp := NewWorkerPool(curlConfig, permChan, resultChan, progressBar)

	go wp.worker()

	// Test successful case
	permChan <- []string{"test1"}

	select {
	case result := <-resultChan:
		if result.Err != nil {
			t.Errorf("Expected no error, got %v", result.Err)
		}
		if !reflect.DeepEqual(result.Payload, []string{"test1"}) {
			t.Errorf("Expected payload [test1], got %v", result.Payload)
		}
		if result.Response.StatusCode != 200 {
			t.Errorf("Expected status code 200, got %d", result.Response.StatusCode)
		}
	case <-time.After(time.Second):
		t.Error("Timed out waiting for result")
	}

	// Test error case (invalid permutation length)
	permChan <- []string{"test2", "extraValue"}

	select {
	case result := <-resultChan:
		if result.Err == nil || result.Err.Error() != "error: length of permutation and values are not equal" {
			t.Errorf("Expected 'length of permutation and values are not equal' error, got %v", result.Err)
		}
		if !reflect.DeepEqual(result.Payload, []string{"test2", "extraValue"}) {
			t.Errorf("Expected payload [test2 extraValue], got %v", result.Payload)
		}
	case <-time.After(time.Second):
		t.Error("Timed out waiting for result")
	}

	close(permChan)
}
