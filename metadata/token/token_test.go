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
	"encoding/json"
	"testing"
)

type testTokenData struct {
	Request int `json:"request"`
}

func (d *testTokenData) MarshalJSON() ([]byte, error) {
	type X testTokenData
	x := X(*d)
	return json.Marshal(x)
}

func (d *testTokenData) UnmarshalJSON(data []byte) error {
	type X *testTokenData
	x := X(d)
	return json.Unmarshal(data, &x)
}

func TestToken(t *testing.T) {
	c := NewTokenCodec(1, "example key 1234")
	data := testTokenData{
		Request: 1,
	}
	expire := 100
	token := NewToken(&data, expire)
	res, err := c.EncodeToken(token)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("Token =", res)

	decodedData := testTokenData{}
	token, err = c.DecodeToken(res, &data)
	if err != nil {
		t.Fatal(err)
	}

	if token.Expires != expire {
		t.Errorf("expected token.Expires to be %d, got %d", expire, token.Expires)
	}

	if data.Request != token.Data.(*testTokenData).Request {
		t.Errorf("expected token.Request to be %d, got %d", data.Request, decodedData.Request)
	}
}
