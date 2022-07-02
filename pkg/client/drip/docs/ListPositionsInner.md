# ListPositionsInner

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Pubkey** | **string** |  | 
**Vault** | **string** |  | 
**Authority** | **string** |  | 
**DepositedTokenAAmount** | **int32** |  | 
**WithdrawnTokenBAmount** | **int32** |  | 
**DepositTimestamp** | **string** |  | 
**DcaPeriodIdBeforeDeposit** | **int32** |  | 
**NumberOfSwaps** | **int32** |  | 
**PeriodicDripAmount** | **int32** |  | 
**IsClosed** | **bool** |  | 

## Methods

### NewListPositionsInner

`func NewListPositionsInner(pubkey string, vault string, authority string, depositedTokenAAmount int32, withdrawnTokenBAmount int32, depositTimestamp string, dcaPeriodIdBeforeDeposit int32, numberOfSwaps int32, periodicDripAmount int32, isClosed bool, ) *ListPositionsInner`

NewListPositionsInner instantiates a new ListPositionsInner object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewListPositionsInnerWithDefaults

`func NewListPositionsInnerWithDefaults() *ListPositionsInner`

NewListPositionsInnerWithDefaults instantiates a new ListPositionsInner object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetPubkey

`func (o *ListPositionsInner) GetPubkey() string`

GetPubkey returns the Pubkey field if non-nil, zero value otherwise.

### GetPubkeyOk

`func (o *ListPositionsInner) GetPubkeyOk() (*string, bool)`

GetPubkeyOk returns a tuple with the Pubkey field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPubkey

`func (o *ListPositionsInner) SetPubkey(v string)`

SetPubkey sets Pubkey field to given value.


### GetVault

`func (o *ListPositionsInner) GetVault() string`

GetVault returns the Vault field if non-nil, zero value otherwise.

### GetVaultOk

`func (o *ListPositionsInner) GetVaultOk() (*string, bool)`

GetVaultOk returns a tuple with the Vault field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetVault

`func (o *ListPositionsInner) SetVault(v string)`

SetVault sets Vault field to given value.


### GetAuthority

`func (o *ListPositionsInner) GetAuthority() string`

GetAuthority returns the Authority field if non-nil, zero value otherwise.

### GetAuthorityOk

`func (o *ListPositionsInner) GetAuthorityOk() (*string, bool)`

GetAuthorityOk returns a tuple with the Authority field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAuthority

`func (o *ListPositionsInner) SetAuthority(v string)`

SetAuthority sets Authority field to given value.


### GetDepositedTokenAAmount

`func (o *ListPositionsInner) GetDepositedTokenAAmount() int32`

GetDepositedTokenAAmount returns the DepositedTokenAAmount field if non-nil, zero value otherwise.

### GetDepositedTokenAAmountOk

`func (o *ListPositionsInner) GetDepositedTokenAAmountOk() (*int32, bool)`

GetDepositedTokenAAmountOk returns a tuple with the DepositedTokenAAmount field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDepositedTokenAAmount

`func (o *ListPositionsInner) SetDepositedTokenAAmount(v int32)`

SetDepositedTokenAAmount sets DepositedTokenAAmount field to given value.


### GetWithdrawnTokenBAmount

`func (o *ListPositionsInner) GetWithdrawnTokenBAmount() int32`

GetWithdrawnTokenBAmount returns the WithdrawnTokenBAmount field if non-nil, zero value otherwise.

### GetWithdrawnTokenBAmountOk

`func (o *ListPositionsInner) GetWithdrawnTokenBAmountOk() (*int32, bool)`

GetWithdrawnTokenBAmountOk returns a tuple with the WithdrawnTokenBAmount field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWithdrawnTokenBAmount

`func (o *ListPositionsInner) SetWithdrawnTokenBAmount(v int32)`

SetWithdrawnTokenBAmount sets WithdrawnTokenBAmount field to given value.


### GetDepositTimestamp

`func (o *ListPositionsInner) GetDepositTimestamp() string`

GetDepositTimestamp returns the DepositTimestamp field if non-nil, zero value otherwise.

### GetDepositTimestampOk

`func (o *ListPositionsInner) GetDepositTimestampOk() (*string, bool)`

GetDepositTimestampOk returns a tuple with the DepositTimestamp field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDepositTimestamp

`func (o *ListPositionsInner) SetDepositTimestamp(v string)`

SetDepositTimestamp sets DepositTimestamp field to given value.


### GetDcaPeriodIdBeforeDeposit

`func (o *ListPositionsInner) GetDcaPeriodIdBeforeDeposit() int32`

GetDcaPeriodIdBeforeDeposit returns the DcaPeriodIdBeforeDeposit field if non-nil, zero value otherwise.

### GetDcaPeriodIdBeforeDepositOk

`func (o *ListPositionsInner) GetDcaPeriodIdBeforeDepositOk() (*int32, bool)`

GetDcaPeriodIdBeforeDepositOk returns a tuple with the DcaPeriodIdBeforeDeposit field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDcaPeriodIdBeforeDeposit

`func (o *ListPositionsInner) SetDcaPeriodIdBeforeDeposit(v int32)`

SetDcaPeriodIdBeforeDeposit sets DcaPeriodIdBeforeDeposit field to given value.


### GetNumberOfSwaps

`func (o *ListPositionsInner) GetNumberOfSwaps() int32`

GetNumberOfSwaps returns the NumberOfSwaps field if non-nil, zero value otherwise.

### GetNumberOfSwapsOk

`func (o *ListPositionsInner) GetNumberOfSwapsOk() (*int32, bool)`

GetNumberOfSwapsOk returns a tuple with the NumberOfSwaps field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNumberOfSwaps

`func (o *ListPositionsInner) SetNumberOfSwaps(v int32)`

SetNumberOfSwaps sets NumberOfSwaps field to given value.


### GetPeriodicDripAmount

`func (o *ListPositionsInner) GetPeriodicDripAmount() int32`

GetPeriodicDripAmount returns the PeriodicDripAmount field if non-nil, zero value otherwise.

### GetPeriodicDripAmountOk

`func (o *ListPositionsInner) GetPeriodicDripAmountOk() (*int32, bool)`

GetPeriodicDripAmountOk returns a tuple with the PeriodicDripAmount field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPeriodicDripAmount

`func (o *ListPositionsInner) SetPeriodicDripAmount(v int32)`

SetPeriodicDripAmount sets PeriodicDripAmount field to given value.


### GetIsClosed

`func (o *ListPositionsInner) GetIsClosed() bool`

GetIsClosed returns the IsClosed field if non-nil, zero value otherwise.

### GetIsClosedOk

`func (o *ListPositionsInner) GetIsClosedOk() (*bool, bool)`

GetIsClosedOk returns a tuple with the IsClosed field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIsClosed

`func (o *ListPositionsInner) SetIsClosed(v bool)`

SetIsClosed sets IsClosed field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


