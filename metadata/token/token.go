package token

/**
 * Copyright (C) 2018 Preetam Jinka
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"strconv"
	"strings"
)

var (
	ErrInvalidToken = errors.New("link token invalid")
)

type TokenCodec struct {
	key        string
	keyVersion int
}

func NewTokenCodec(keyVersion int, key string) *TokenCodec {
	return &TokenCodec{
		key:        key,
		keyVersion: keyVersion,
	}
}

type Token struct {
	Data    TokenData `json:"data"`
	Expires int       `json:"expires"`
}

type TokenData interface {
	json.Marshaler
	json.Unmarshaler
}

type UserTokenData struct {
	User string `json:"user"`
}

func (d *UserTokenData) MarshalJSON() ([]byte, error) {
	type X UserTokenData
	x := X(*d)
	return json.Marshal(x)
}

func (d *UserTokenData) UnmarshalJSON(data []byte) error {
	type X *UserTokenData
	x := X(d)
	return json.Unmarshal(data, &x)
}

func NewToken(data TokenData, expires int) *Token {
	return &Token{
		Data:    data,
		Expires: expires,
	}
}

func (c *TokenCodec) EncodeToken(token *Token) (string, error) {
	buf := bytes.NewBuffer(nil)
	err := json.NewEncoder(buf).Encode(token)

	block, err := aes.NewCipher([]byte(c.key))
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nil, nonce, buf.Bytes(), nil)

	nonceStr := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(nonce)
	dataStr := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(ciphertext)
	versionStr := strconv.FormatInt(int64(c.keyVersion), 10)
	return versionStr + "." + nonceStr + "." + dataStr, nil
}

func (c *TokenCodec) DecodeToken(tokenString string, tokenData TokenData) (*Token, error) {
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidToken
	}

	versionStr := parts[0]
	nonce, err := base64.URLEncoding.WithPadding(base64.NoPadding).DecodeString(parts[1])
	if err != nil {
		return nil, err
	}
	ciphertext, err := base64.URLEncoding.WithPadding(base64.NoPadding).DecodeString(parts[2])
	if err != nil {
		return nil, err
	}

	version, err := strconv.ParseInt(versionStr, 10, 64)
	if err != nil {
		return nil, err
	}

	if c.keyVersion != int(version) {
		return nil, ErrInvalidToken
	}

	block, err := aes.NewCipher([]byte(c.key))
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	data, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	linkToken := Token{
		Data: tokenData,
	}
	err = json.NewDecoder(bytes.NewReader(data)).Decode(&linkToken)
	if err != nil {
		return nil, err
	}
	tokenData = linkToken.Data
	return &linkToken, nil
}
