package util

import (
	"bytes"
	"crypto/md5"
	"encoding/gob"
	"fmt"
	"io/ioutil"
)

func Store(e interface{}) error {
	buf := bytes.Buffer{}
	en := gob.NewEncoder(&buf)
	if err := en.Encode(e); err != nil {
		return err
	}
	if err := ioutil.WriteFile("./snap", buf.Bytes(), 0666); err != nil {
		return err
	}

	h := md5.New()
	if _, err := h.Write(buf.Bytes()); err != nil {
		return err
	}
	if err := ioutil.WriteFile("./md5", h.Sum(nil), 0666); err != nil {
		return err
	}
	return nil
}

func Load(e interface{}) error {
	data, err := ioutil.ReadFile("./snap")
	if err != nil {
		return err
	}

	data2, err := ioutil.ReadFile("./md5")
	if err != nil {
		return err
	}

	h := md5.New()
	if _, err := h.Write(data); err != nil {
		return err
	}

	if bytes.Compare(h.Sum(nil), data2) != 0 {
		return fmt.Errorf("md5 not correct")
	}

	de := gob.NewDecoder(bytes.NewBuffer(data))
	return de.Decode(e)
}
