package api

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log"
)

// adapted from:
// https://www.melvinvivas.com/how-to-encrypt-and-decrypt-data-using-aes/

func encrypt(input, salt string) string {
	//var key []byte
	//hex.Encode(key, []byte(salt))
	block, err := aes.NewCipher([]byte(salt))
	if err != nil {
		log.Printf("could not create cipher: %v", err)
		return ""
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		log.Printf("could not create GCM: %v", err)
		return ""
	}
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		log.Printf("could not generate nonce: %v", err)
		return "" // todo: probably should handle error differently
	}
	output := aesGCM.Seal(nonce, nonce, []byte(input), nil)
	return fmt.Sprintf("%x", output)
}

func decrypt(input, salt string) string {
	enc, _ := hex.DecodeString(input)
	block, err := aes.NewCipher([]byte(salt))
	if err != nil {
		log.Printf("could not create cipher: %v", err)
		return ""
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		log.Printf("could not create GCM: %v", err)
		return ""
	}
	nonceSize := aesGCM.NonceSize()
	nonce, cipherText := enc[:nonceSize], enc[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, cipherText, nil)
	if err != nil {
		log.Printf("could not decrypt cipher: %v", err)
	}
	return fmt.Sprintf("%s", plaintext)
}