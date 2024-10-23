package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type Config struct {
	Alphabet    string `json:"alphabet"`
	Epsilon     string `json:"epsilon"`
	LearnerMode string `json:"learner_mode"`
	ServerAddr  string `json:"server_address"`
	ServerPort  string `json:"server_port"`
	MatMode     string `json:"mat_mode"`
}

func LoadConfig() (*Config, error) {
	file, err := os.Open("/home/alexandr/BMSTU_git/IU9-ToFL/lab2/config.json")
	// file, err := os.Open("E:/BMSTU_git/IU9-ToFL/lab2/config.json")
	if err != nil {
		return nil, fmt.Errorf("ошибка при открытии файла конфигурации: %v", err)
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("ошибка при чтении файла конфигурации: %v", err)
	}

	var config Config
	err = json.Unmarshal(bytes, &config)
	if err != nil {
		return nil, fmt.Errorf("ошибка при разборе конфигурации: %v", err)
	}

	return &config, nil
}
