package md5f

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"io"
	"log"
	"os"
)

var (
	ErrFileCorrupt = errors.New("Check md5 is failed. File is corrupt")
)

type Md5App struct {
}

func NewMd5App() *Md5App {
	return &Md5App{}
}

//ComputeMd5 Md5 Calculates and returns a array of byte
func (c *Md5App) ComputeMd5(filePath string) ([]byte, error) {
	var result []byte
	file, err := os.Open(filePath)
	if err != nil {
		return result, err
	}
	defer func() {
		if e := file.Close(); e != nil {
			log.Println(e)
		}
	}()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {

		return result, err
	}
	return hash.Sum(result), nil
}

//CheckMd5 Calculates the md5 hash and checks by comparing it with bMd5
func (c *Md5App) CheckMd5(pFile, bMd5 string) (res bool, err error) {
	var dst string
	hash, err := c.ComputeMd5(pFile)
	if err != nil {
		return false, err
	}
	dst = hex.EncodeToString(hash[:])

	if dst != bMd5 {
		return false, ErrFileCorrupt
	}
	return true, nil
}

//ComputeMd5String Md5 Calculates and returns a string
func (c *Md5App) ComputeMd5String(filePath string) (s string, err error) {
	hash, err := c.ComputeMd5(filePath)
	if err != nil {
		return "", err
	}
	s = hex.EncodeToString(hash[:])
	return s, nil

}
