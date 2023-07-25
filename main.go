package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

type Config struct {
	BaseURL        string   `yaml:"baseUrl"`
	SealedKeys     []string `yaml:"sealedKeys"`
	UnsealInterval int      `yaml:"unsealInterval"`
}

type Response struct {
	Sealed   bool `json:"sealed"`
	Progress int  `json:"progress"`
}

func main() {
	config, err := loadConfig("config.yaml")
	if err != nil {
		fmt.Println("Error loading config:", err)
		return
	}

	unseal(config)
}

func loadConfig(filename string) (*Config, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	config := &Config{}
	err = yaml.Unmarshal(data, config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func unseal(config *Config) {
	client := &http.Client{}

	for {
		fmt.Println("Checking vault status...")
		resp, err := client.Get(config.BaseURL + "/v1/sys/seal-status")
		if err != nil {
			fmt.Println("Error while sending request:", err)
			time.Sleep(5 * time.Second)
			continue
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Error reading response:", err)
			time.Sleep(5 * time.Second)
			continue
		}

		var result Response
		err = json.Unmarshal(body, &result)
		if err != nil {
			fmt.Println("Error parsing JSON:", err)
			time.Sleep(5 * time.Second)
			continue
		}

		if result.Sealed {
			for _, key := range config.SealedKeys {
				payload := strings.NewReader(fmt.Sprintf("{\"key\":\"%v\"}", key))
				req, err := http.NewRequest("PUT", config.BaseURL+"/v1/sys/unseal", payload)
				if err != nil {
					fmt.Println("Error creating request:", err)
					continue
				}

				res, err := client.Do(req)
				if err != nil {
					fmt.Println("Error sending request:", err)
					continue
				}
				defer res.Body.Close()
			}
			fmt.Println("Vault was unsealed.")
		}
		fmt.Println("Vault is OK.")

		time.Sleep(time.Duration(config.UnsealInterval) * time.Second)
	}
}
