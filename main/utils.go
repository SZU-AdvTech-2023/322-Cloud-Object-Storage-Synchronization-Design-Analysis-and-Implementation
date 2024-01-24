package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
)

func getHashValueBySHA256(file *os.File) string {
	hash := sha256.New()
	buffer, err := io.ReadAll(file)
	if err != nil {
		fmt.Println("Failed to read the file")
		return ""
	}
	hash.Write([]byte(buffer))
	return hex.EncodeToString(hash.Sum(nil))
}

func getLocalFileHash(filePath string) string {
	/*
		Get the hash value of local file by using SHA-256 algorithm
		param : filePath is the path of local file
		return: Hash value
	*/
	file, err := os.OpenFile(filePath, os.O_RDONLY, 0666)
	if err != nil {
		fmt.Println("Cannot open the file：" + filePath)
		return ""
	}
	defer file.Close()
	return getHashValueBySHA256(file)
}

func getBufferHash(buffer []byte) string {
	hash := sha256.New()
	hash.Write(buffer)
	hashBytes := hash.Sum(nil)
	hashValue := strings.TrimSpace(hex.EncodeToString(hashBytes))
	return hashValue
}

func getCloudFileHash(cloudPath string) string {
	/*You need to download the file to a tempPath
	 */
	tempPath := home + "/tempCloudFile/"
	_, err := os.Stat(tempPath)
	if os.IsNotExist(err) {
		err = os.Mkdir(tempPath, 0777)
		if err != nil {
			panic(err)
		}
	}
	cfs.download(cloudPath, tempPath+path.Base(cloudPath))
	return getLocalFileHash(tempPath + path.Base(cloudPath))
}
