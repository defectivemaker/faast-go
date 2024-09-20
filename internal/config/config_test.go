package config

import (
	"math"
	"os"
	"reflect"
	"testing"
)

func TestLoadWordlists(t *testing.T) {
	// Create temporary wordlist files
	tempFiles := []string{"test_wordlist1.txt", "test_wordlist2.txt"}
	wordlistContents := [][]string{
		{"word1", "word2", "word3"},
		{"word4", "word5", "word6"},
	}

	for i, filename := range tempFiles {
		content := ""
		for _, word := range wordlistContents[i] {
			content += word + "\n"
		}
		err := os.WriteFile(filename, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test wordlist file: %v", err)
		}
		defer os.Remove(filename)
	}

	// Test LoadWordlists function
	wordlists, err := LoadWordlists(tempFiles)
	if err != nil {
		t.Fatalf("LoadWordlists failed: %v", err)
	}

	if !reflect.DeepEqual(wordlists, wordlistContents) {
		t.Errorf("LoadWordlists returned unexpected result. Got %v, want %v", wordlists, wordlistContents)
	}

	// Test with non-existent file
	_, err = LoadWordlists([]string{"non_existent_file.txt"})
	if err == nil {
		t.Error("LoadWordlists should have returned an error for non-existent file")
	}
}

func TestLoadConfig(t *testing.T) {
	// Create a temporary config file
	configContent := `
type: payload
endpoint: http://example.com
fields:
  - field1
  - field2
wordlists:
  - wordlist1.txt
staticValues:
  - static1
cookies:
  - cookie1=value1
validateType: status
sizeDefault: 100
codeDefault: 404
rateLimit: 10
timeout: 30
shardIndex: 0
numShards: 2
`
	tempConfigFile := "test_config.yaml"
	err := os.WriteFile(tempConfigFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}
	defer os.Remove(tempConfigFile)

	// Test LoadConfig function
	config, err := LoadConfig(tempConfigFile)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Verify loaded config
	expectedConfig := &YamlConfig{
		Type:         "payload",
		Endpoint:     "http://example.com",
		Fields:       []string{"field1", "field2"},
		Wordlists:    []string{"wordlist1.txt"},
		StaticValues: []string{"static1"},
		Cookies:      []string{"cookie1=value1"},
		ValidateType: "status",
		SizeDefault:  100,
		CodeDefault:  404,
		RateLimit:    10,
		Timeout:      30,
		ShardIndex:   0,
		NumShards:    2,
	}

	if !reflect.DeepEqual(config, expectedConfig) {
		t.Errorf("LoadConfig returned unexpected result. Got %+v, want %+v", config, expectedConfig)
	}

	// Test with non-existent file
	_, err = LoadConfig("non_existent_config.yaml")
	if err == nil {
		t.Error("LoadConfig should have returned an error for non-existent file")
	}

	// Test with invalid config
	invalidConfigContent := `
type: payload
endpoint: http://example.com
fields:
  - field1
wordlists:
  - wordlist1.txt
  - wordlist2.txt
`
	err = os.WriteFile(tempConfigFile, []byte(invalidConfigContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create invalid test config file: %v", err)
	}

	_, err = LoadConfig(tempConfigFile)
	if err == nil {
		t.Error("LoadConfig should have returned an error for invalid config")
	}
}

func TestYamlConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  YamlConfig
		wantErr bool
	}{
		{
			name: "Valid config",
			config: YamlConfig{
				Type:         "payload",
				Endpoint:     "http://example.com",
				Fields:       []string{"field1", "field2"},
				Wordlists:    []string{"wordlist1.txt"},
				StaticValues: []string{"static1"},
			},
			wantErr: false,
		},
		{
			name: "Missing endpoint",
			config: YamlConfig{
				Type:         "payload",
				Fields:       []string{"field1"},
				Wordlists:    []string{"wordlist1.txt"},
				StaticValues: []string{},
			},
			wantErr: true,
		},
		{
			name: "Mismatched fields and wordlists+staticValues",
			config: YamlConfig{
				Type:         "payload",
				Endpoint:     "http://example.com",
				Fields:       []string{"field1", "field2"},
				Wordlists:    []string{"wordlist1.txt"},
				StaticValues: []string{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("YamlConfig.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestYamlConfig_SetDefaults(t *testing.T) {
	config := &YamlConfig{}
	config.SetDefaults()

	if config.CodeDefault != 404 {
		t.Errorf("SetDefaults() CodeDefault = %v, want 404", config.CodeDefault)
	}
	if config.RateLimit != math.MaxFloat64 {
		t.Errorf("SetDefaults() RateLimit = %v, want %v", config.RateLimit, math.MaxFloat64)
	}
	if config.Timeout != 5 {
		t.Errorf("SetDefaults() Timeout = %v, want 5", config.Timeout)
	}
	if config.NumShards != 1 {
		t.Errorf("SetDefaults() NumShards = %v, want 1", config.NumShards)
	}
}
