/*
Drip Backend

Drip backend service.

API version: 1.0.0
Contact: dcafmocha@protonmail.com
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package drip

import (
	"encoding/json"
)

// ListTokensInner struct for ListTokensInner
type ListTokensInner struct {
	Pubkey string `json:"pubkey"`
	Symbol string `json:"symbol"`
	Decimals int32 `json:"decimals"`
}

// NewListTokensInner instantiates a new ListTokensInner object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewListTokensInner(pubkey string, symbol string, decimals int32) *ListTokensInner {
	this := ListTokensInner{}
	this.Pubkey = pubkey
	this.Symbol = symbol
	this.Decimals = decimals
	return &this
}

// NewListTokensInnerWithDefaults instantiates a new ListTokensInner object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewListTokensInnerWithDefaults() *ListTokensInner {
	this := ListTokensInner{}
	return &this
}

// GetPubkey returns the Pubkey field value
func (o *ListTokensInner) GetPubkey() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Pubkey
}

// GetPubkeyOk returns a tuple with the Pubkey field value
// and a boolean to check if the value has been set.
func (o *ListTokensInner) GetPubkeyOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Pubkey, true
}

// SetPubkey sets field value
func (o *ListTokensInner) SetPubkey(v string) {
	o.Pubkey = v
}

// GetSymbol returns the Symbol field value
func (o *ListTokensInner) GetSymbol() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Symbol
}

// GetSymbolOk returns a tuple with the Symbol field value
// and a boolean to check if the value has been set.
func (o *ListTokensInner) GetSymbolOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Symbol, true
}

// SetSymbol sets field value
func (o *ListTokensInner) SetSymbol(v string) {
	o.Symbol = v
}

// GetDecimals returns the Decimals field value
func (o *ListTokensInner) GetDecimals() int32 {
	if o == nil {
		var ret int32
		return ret
	}

	return o.Decimals
}

// GetDecimalsOk returns a tuple with the Decimals field value
// and a boolean to check if the value has been set.
func (o *ListTokensInner) GetDecimalsOk() (*int32, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Decimals, true
}

// SetDecimals sets field value
func (o *ListTokensInner) SetDecimals(v int32) {
	o.Decimals = v
}

func (o ListTokensInner) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if true {
		toSerialize["pubkey"] = o.Pubkey
	}
	if true {
		toSerialize["symbol"] = o.Symbol
	}
	if true {
		toSerialize["decimals"] = o.Decimals
	}
	return json.Marshal(toSerialize)
}

type NullableListTokensInner struct {
	value *ListTokensInner
	isSet bool
}

func (v NullableListTokensInner) Get() *ListTokensInner {
	return v.value
}

func (v *NullableListTokensInner) Set(val *ListTokensInner) {
	v.value = val
	v.isSet = true
}

func (v NullableListTokensInner) IsSet() bool {
	return v.isSet
}

func (v *NullableListTokensInner) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableListTokensInner(val *ListTokensInner) *NullableListTokensInner {
	return &NullableListTokensInner{value: val, isSet: true}
}

func (v NullableListTokensInner) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableListTokensInner) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}

