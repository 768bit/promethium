package common

import (
	"debug/elf"
	"encoding/json"
	"io/ioutil"
)

type KernelMeta struct {
	Machine  string `json:"machine" yaml:"machine"`
	Platform string `json:"platform" yaml:"platform"`
	Version  string `json:"version" yaml:"version"`
	From     string `json:"from" yaml:"from"`
}

func LoadKernelMeta(path string) (*KernelMeta, error) {
	var km KernelMeta
	ba, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(ba, &km)
	if err != nil {
		return nil, err
	}
	return &km, nil
}

func LoadKernelElf(path string) (*elf.FileHeader, error) {
	kf, err := elf.Open(path)
	if err != nil {
		return nil, err
	}
	return &kf.FileHeader, nil
}
