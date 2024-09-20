package config

import (
	"bufio"
	"fmt"
	"math"
	"os"

	"gopkg.in/yaml.v3"
)

type YamlConfig struct {
	Type         string   `yaml:"type"`
	Endpoint     string   `yaml:"endpoint"`
	Fields       []string `yaml:"fields"`
	Wordlists    []string `yaml:"wordlists"`
	StaticValues []string `yaml:"staticValues"`
	Cookies      []string `yaml:"cookies"`
	ValidateType string   `yaml:"validateType"`
	SizeDefault  int      `yaml:"sizeDefault"`
	CodeDefault  int      `yaml:"codeDefault"`
	RateLimit    float64  `yaml:"rateLimit"`
	Timeout      int      `yaml:"timeout"`
	ShardIndex   int      `yaml:"shardIndex"`
	NumShards    int      `yaml:"numShards"`
}

func LoadWordlists(filenames []string) ([][]string, error) {
	wordlists := make([][]string, len(filenames))
	for i, filename := range filenames {
		file, err := os.Open(filename)
		if err != nil {
			return nil, fmt.Errorf("error opening wordlist file %s: %w", filename, err)
		}
		defer file.Close()

		var words []string
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			words = append(words, scanner.Text())
		}

		if err := scanner.Err(); err != nil {
			return nil, fmt.Errorf("error reading wordlist file %s: %w", filename, err)
		}

		wordlists[i] = words
	}
	return wordlists, nil
}

func LoadConfig(filename string) (*YamlConfig, error) {
	file, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var config YamlConfig
	err = yaml.Unmarshal(file, &config)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	config.SetDefaults()

	return &config, nil
}

func (c *YamlConfig) Validate() error {
	if c.Endpoint == "" {
		return fmt.Errorf("endpoint is required")
	}
	fmt.Println(c.Fields)
	fmt.Println(c.Wordlists)
	fmt.Println(c.StaticValues)
	if c.Type == "payload" && len(c.Fields) != (len(c.Wordlists)+len(c.StaticValues)) {
		return fmt.Errorf("number of fields must equal number of wordlists + staticValues")
	}

	return nil
}

func (c *YamlConfig) SetDefaults() {
	if c.CodeDefault == 0 {
		c.CodeDefault = 404
	}
	if c.RateLimit == 0 {
		c.RateLimit = math.MaxFloat64
	}
	if c.Timeout == 0 {
		c.Timeout = 5
	}
	if c.NumShards == 0 {
		c.NumShards = 1
	}
}
