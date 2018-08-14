package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type walletConfig struct {
	Path     string `json:"Path"`
	Password string `json:"Password"`
}

type TxReplayConfig struct {
	Wallets []walletConfig `json:"Wallets"`
}

func (this *TxReplayConfig) loadConfig(fileName string) error {
	data, err := this.readFile(fileName)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, this)
	if err != nil {
		return fmt.Errorf("json.Unmarshal TxReplayConfig:%s error:%s", data, err)
	}
	return nil
}

func (this *TxReplayConfig) readFile(fileName string) ([]byte, error) {
	file, err := os.OpenFile(fileName, os.O_RDONLY, 0666)
	if err != nil {
		return nil, fmt.Errorf("OpenFile %s error %s", fileName, err)
	}
	defer func() {
		err := file.Close()
		if err != nil {
			fmt.Errorf("File %s close error %s", fileName, err)
		}
	}()
	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("ioutil.ReadAll %s error %s", fileName, err)
	}
	return data, nil
}
