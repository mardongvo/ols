package main

import (
	"bytes"
	"testing"
)

func TestWrongCfg(t *testing.T) {
	CASES := [][]byte{[]byte(`{"database":"1", "listen":""}`),
		[]byte(`{"database":"", "listen":"1"}`),
		[]byte(`{"database":"", "listen":"1"`),
	}
	for _, v := range CASES {
		buf := bytes.NewBuffer(v)
		_, err := ReadConfig(buf)
		if err == nil {
			t.Errorf("Конфиг %v - должна быть ошибка, но ее нет", v)
		}
	}
}

func TestRightsCfg(t *testing.T) {
	CASES := [][]byte{[]byte(`{"database":"1", "listen":"1"}`),
		[]byte(`{"database":"1", "listen":"1", "something":"1"}`),
	}
	for _, v := range CASES {
		buf := bytes.NewBuffer(v)
		_, err := ReadConfig(buf)
		if err != nil {
			t.Errorf("Конфиг %v - без ошибок, но есть: %v", v, err)
		}
	}
}
