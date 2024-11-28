package wechat

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"math/rand"
	"time"
)

const (
	blockSize = 32
	blockMask = blockSize - 1

	randomSize     = 16
	contentLenSize = 4
)

func init() {
	rand.New(rand.NewSource(time.Now().UnixNano()))
}

func EncryptMsg(appID string, msg string, encodingAesKey string) (string, error) {
	aesKey, err := base64.StdEncoding.DecodeString(encodingAesKey + "=")
	if err != nil {
		return "", err
	}

	if len(aesKey) != 32 {
		return "", errors.New("invalid aes key length")
	}

	appIDOffset := randomSize + contentLenSize + len(msg)
	contentLen := appIDOffset + len(appID)
	amountToPad := blockSize - contentLen&blockMask
	plaintextLen := contentLen + amountToPad
	plaintext := make([]byte, plaintextLen)

	copy(plaintext[:randomSize], randString(randomSize))
	binary.BigEndian.PutUint32(plaintext[randomSize:randomSize+contentLenSize], uint32(len(msg)))
	copy(plaintext[randomSize+contentLenSize:], msg)
	copy(plaintext[appIDOffset:], appID)

	for i := contentLen; i < plaintextLen; i++ {
		plaintext[i] = byte(amountToPad)
	}

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return "", err
	}

	mode := cipher.NewCBCEncrypter(block, aesKey[:aes.BlockSize])
	mode.CryptBlocks(plaintext, plaintext)

	return base64.StdEncoding.EncodeToString(plaintext), nil
}

func DecryptMsg(appID, encryptedMsg, encodingAesKey string) (string, error) {
	aesKey, err := base64.StdEncoding.DecodeString(encodingAesKey + "=")
	if err != nil {
		return "", err
	}

	if len(aesKey) != 32 {
		return "", errors.New("invalid aes key length")
	}

	cipherText, err := base64.StdEncoding.DecodeString(encryptedMsg)
	if err != nil {
		return "", fmt.Errorf("base64 decode error: %v", err)
	}

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return "", fmt.Errorf("failed to create aes cipher: %v", err)
	}

	plainBytes := make([]byte, len(cipherText))
	mode := cipher.NewCBCDecrypter(block, aesKey[:aes.BlockSize])
	mode.CryptBlocks(plainBytes, cipherText)

	pad := int(plainBytes[len(plainBytes)-1])
	if pad < 1 || pad > blockSize {
		return "", errors.New("invalid padding byte")
	}
	plainBytes = plainBytes[:len(plainBytes)-pad]

	content := plainBytes[randomSize:]
	if len(content) < contentLenSize {
		return "", errors.New("invalid content length")
	}

	contentLen := binary.BigEndian.Uint32(content[:contentLenSize])
	if len(content) < int(contentLenSize+contentLen) {
		return "", errors.New("invalid content length")
	}

	rawContent := content[contentLenSize : contentLenSize+contentLen]
	fromAppID := string(content[contentLenSize+contentLen:])

	if fromAppID != appID {
		return "", errors.New("app id mismatch")
	}

	return string(rawContent), nil
}

func randString(n int) string {
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
