package main

import (
	"encoding/json"
	"fmt"
	"io"
)

type OLSConfig struct {
	Database      string `json:"database"` //database connection string
	ListenAddress string `json:"listen"`   //listen address(optional) and/or port(optional)
}

func ReadConfig(r io.Reader) (OLSConfig, error) {
	var cfg OLSConfig
	dec := json.NewDecoder(r)
	err := dec.Decode(&cfg)
	if err != nil {
		return cfg, fmt.Errorf("config: ошибка json %v", err)
	}
	if cfg.Database == "" {
		return cfg, fmt.Errorf("config: строка подключения к БД пуста")
	}
	if cfg.ListenAddress == "" {
		return cfg, fmt.Errorf("config: адрес/порт прослушивания не определены")
	}
	return cfg, nil
}
