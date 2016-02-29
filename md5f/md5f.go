package md5f

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"io"
	"log"
	"os"
)

//ComputeMd5 Md5 Calculates and returns a array of byte
func ComputeMd5(filePath string) ([]byte, error) {
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
func CheckMd5(pFile, bMd5 string) (res bool, err error) {
	var dst string
	hash, err := ComputeMd5(pFile)
	if err != nil {
		log.Println("file:", pFile, "dst:", dst, "curhash:", bMd5)
		return false, err
	}
	dst = hex.EncodeToString(hash[:])

	if dst != bMd5 {
		log.Println("file:", pFile, "dst:", dst, "curhash:", bMd5)
		err := errors.New("Check md5 is failed. File is corrupt")
		return false, err
	}
	return true, nil
}

//ComputeMd5String Md5 Calculates and returns a string
func ComputeMd5String(filePath string) (s string, err error) {
	hash, err := ComputeMd5(filePath)
	if err != nil {
		return "", err
	}
	s = hex.EncodeToString(hash[:])
	return s, nil

}
