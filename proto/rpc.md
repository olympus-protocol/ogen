# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [chain.proto](#chain.proto)
    - [AccountInfo](#.AccountInfo)
    - [ChainInfo](#.ChainInfo)
    - [SubscribeValidatorRequest](#.SubscribeValidatorRequest)
  
    - [Chain](#.Chain)
  
- [common.proto](#common.proto)
    - [Account](#.Account)
    - [Block](#.Block)
    - [BlockHeader](#.BlockHeader)
    - [Empty](#.Empty)
    - [Hash](#.Hash)
    - [KeyPair](#.KeyPair)
    - [KeyPairs](#.KeyPairs)
    - [Number](#.Number)
    - [RawData](#.RawData)
    - [Success](#.Success)
    - [TransferMulti](#.TransferMulti)
    - [TransferSingle](#.TransferSingle)
    - [Tx](#.Tx)
    - [ValidatorRegistry](#.ValidatorRegistry)
    - [ValidatorsInfo](#.ValidatorsInfo)
    - [ValidatorsRegistry](#.ValidatorsRegistry)
  
- [network.proto](#network.proto)
    - [IP](#.IP)
    - [NetworkInfo](#.NetworkInfo)
    - [Peer](#.Peer)
    - [Peers](#.Peers)
  
    - [Network](#.Network)
  
- [utils.proto](#utils.proto)
    - [Utils](#.Utils)
  
- [validators.proto](#validators.proto)
    - [Validators](#.Validators)
  
- [wallet.proto](#wallet.proto)
    - [Balance](#.Balance)
    - [ImportWalletData](#.ImportWalletData)
    - [Name](#.Name)
    - [SendTransactionInfo](#.SendTransactionInfo)
    - [Wallets](#.Wallets)
  
    - [Wallet](#.Wallet)
  
- [Scalar Value Types](#scalar-value-types)



<a name="chain.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## chain.proto



<a name=".AccountInfo"></a>

### AccountInfo



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| account | [string](#string) |  |  |
| balance | [uint64](#uint64) |  |  |
| nonce | [uint64](#uint64) |  |  |
| txs | [string](#string) | repeated |  |






<a name=".ChainInfo"></a>

### ChainInfo



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| block_hash | [string](#string) |  |  |
| block_height | [uint64](#uint64) |  |  |
| validators | [ValidatorsInfo](#ValidatorsInfo) |  |  |






<a name=".SubscribeValidatorRequest"></a>

### SubscribeValidatorRequest



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| public_key | [bytes](#bytes) | repeated |  |





 

 

 


<a name=".Chain"></a>

### Chain


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| GetChainInfo | [.Empty](#Empty) | [.ChainInfo](#ChainInfo) |  |
| GetRawBlock | [.Hash](#Hash) | [.Block](#Block) |  |
| GetBlock | [.Hash](#Hash) | [.Block](#Block) |  |
| GetBlockHash | [.Number](#Number) | [.Hash](#Hash) |  |
| GetAccountInfo | [.Account](#Account) | [.AccountInfo](#AccountInfo) |  |
| GetTransaction | [.Hash](#Hash) | [.Tx](#Tx) |  |
| Sync | [.Hash](#Hash) | [.RawData](#RawData) stream |  |
| SubscribeBlocks | [.Empty](#Empty) | [.RawData](#RawData) stream |  |
| SubscribeTransactions | [.KeyPairs](#KeyPairs) | [.RawData](#RawData) stream |  |
| SubscribeValidatorTransactions | [.KeyPairs](#KeyPairs) | [.RawData](#RawData) stream |  |

 



<a name="common.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## common.proto



<a name=".Account"></a>

### Account



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| account | [string](#string) |  |  |






<a name=".Block"></a>

### Block



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| hash | [string](#string) |  |  |
| raw_block | [string](#string) |  |  |
| header | [BlockHeader](#BlockHeader) |  |  |
| txs | [string](#string) | repeated |  |
| signature | [string](#string) |  |  |
| randao_signature | [string](#string) |  |  |






<a name=".BlockHeader"></a>

### BlockHeader



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| version | [int32](#int32) |  |  |
| nonce | [int32](#int32) |  |  |
| tx_merkle_root | [string](#string) |  |  |
| vote_merkle_root | [string](#string) |  |  |
| deposit_merkle_root | [string](#string) |  |  |
| exit_merkle_root | [string](#string) |  |  |
| vote_slashing_merkle_root | [string](#string) |  |  |
| randao_slashing_merkle_root | [string](#string) |  |  |
| proposer_slashing_merkle_root | [string](#string) |  |  |
| prev_block_hash | [string](#string) |  |  |
| timestamp | [int64](#int64) |  |  |
| slot | [uint64](#uint64) |  |  |
| state_root | [string](#string) |  |  |
| fee_address | [string](#string) |  |  |






<a name=".Empty"></a>

### Empty







<a name=".Hash"></a>

### Hash



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| hash | [string](#string) |  |  |






<a name=".KeyPair"></a>

### KeyPair



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| public | [string](#string) |  |  |
| private | [string](#string) |  |  |






<a name=".KeyPairs"></a>

### KeyPairs



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| keys | [string](#string) | repeated |  |






<a name=".Number"></a>

### Number



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| number | [uint64](#uint64) |  |  |






<a name=".RawData"></a>

### RawData



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| data | [string](#string) |  |  |
| type | [string](#string) |  |  |






<a name=".Success"></a>

### Success



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| success | [bool](#bool) |  |  |
| error | [string](#string) |  |  |
| data | [string](#string) |  |  |






<a name=".TransferMulti"></a>

### TransferMulti



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| to | [string](#string) |  |  |
| amount | [uint64](#uint64) |  |  |
| nonce | [uint64](#uint64) |  |  |
| fee | [uint64](#uint64) |  |  |
| signature | [string](#string) |  |  |






<a name=".TransferSingle"></a>

### TransferSingle



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| to | [string](#string) |  |  |
| from_public_key | [string](#string) |  |  |
| amount | [uint64](#uint64) |  |  |
| nonce | [uint64](#uint64) |  |  |
| fee | [uint64](#uint64) |  |  |
| signature | [string](#string) |  |  |






<a name=".Tx"></a>

### Tx



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| hash | [string](#string) |  |  |
| type | [int32](#int32) |  |  |
| version | [int32](#int32) |  |  |
| transfer_single_payload | [TransferSingle](#TransferSingle) |  |  |
| transfer_multi_payload | [TransferMulti](#TransferMulti) |  |  |






<a name=".ValidatorRegistry"></a>

### ValidatorRegistry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| balance | [string](#string) |  |  |
| public_key | [string](#string) |  |  |
| payee_address | [string](#string) |  |  |
| status | [string](#string) |  |  |
| first_active_epoch | [int64](#int64) |  |  |
| last_active_epoch | [int64](#int64) |  |  |






<a name=".ValidatorsInfo"></a>

### ValidatorsInfo



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| active | [int64](#int64) |  |  |
| pending_exit | [int64](#int64) |  |  |
| penalty_exit | [int64](#int64) |  |  |
| exited | [int64](#int64) |  |  |
| starting | [int64](#int64) |  |  |






<a name=".ValidatorsRegistry"></a>

### ValidatorsRegistry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| info | [ValidatorsInfo](#ValidatorsInfo) |  |  |
| validators | [ValidatorRegistry](#ValidatorRegistry) | repeated |  |





 

 

 

 



<a name="network.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## network.proto



<a name=".IP"></a>

### IP



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| host | [string](#string) |  |  |






<a name=".NetworkInfo"></a>

### NetworkInfo



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| peers | [int32](#int32) |  |  |
| protocol | [int32](#int32) |  |  |
| version | [string](#string) |  |  |






<a name=".Peer"></a>

### Peer



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  |  |
| host | [IP](#IP) |  |  |






<a name=".Peers"></a>

### Peers



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| peers | [Peer](#Peer) | repeated |  |





 

 

 


<a name=".Network"></a>

### Network


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| GetNetworkInfo | [.Empty](#Empty) | [.NetworkInfo](#NetworkInfo) |  |
| GetPeersInfo | [.Empty](#Empty) | [.Peers](#Peers) |  |
| AddPeer | [.IP](#IP) | [.Success](#Success) |  |

 



<a name="utils.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## utils.proto


 

 

 


<a name=".Utils"></a>

### Utils


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| GenValidatorKey | [.Number](#Number) | [.KeyPairs](#KeyPairs) | Method: GenValidatorKey Input: message Number Response: message KeyPairs Description: Returns private keys generated for validators start. |
| SubmitRawData | [.RawData](#RawData) | [.Success](#Success) | Method: SubmitRawData Input: message RawData Response: message Success Description: Broadcast a raw elements of different transactions. |
| DecodeRawTransaction | [.RawData](#RawData) | [.Tx](#Tx) | Method: DecodeRawTransaction Input: message RawData Response: message Tx Description: Returns a raw transaction on human readable format. |
| DecodeRawBlock | [.RawData](#RawData) | [.Block](#Block) | Method: DecodeRawBlock Input: message RawData Response: message Block Description: Returns a raw block on human readable format. |

 



<a name="validators.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## validators.proto


 

 

 


<a name=".Validators"></a>

### Validators


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| GetValidatorsList | [.Empty](#Empty) | [.ValidatorsRegistry](#ValidatorsRegistry) |  |
| GetAccountValidators | [.Account](#Account) | [.ValidatorsRegistry](#ValidatorsRegistry) |  |

 



<a name="wallet.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## wallet.proto



<a name=".Balance"></a>

### Balance



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| confirmed | [string](#string) |  |  |
| unconfirmed | [string](#string) |  |  |
| locked | [string](#string) |  |  |
| total | [string](#string) |  |  |






<a name=".ImportWalletData"></a>

### ImportWalletData



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| key | [KeyPair](#KeyPair) |  |  |






<a name=".Name"></a>

### Name



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |






<a name=".SendTransactionInfo"></a>

### SendTransactionInfo



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| account | [string](#string) |  |  |
| amount | [string](#string) |  |  |






<a name=".Wallets"></a>

### Wallets



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| wallets | [string](#string) | repeated |  |





 

 

 


<a name=".Wallet"></a>

### Wallet


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| ListWallets | [.Empty](#Empty) | [.Wallets](#Wallets) |  |
| CreateWallet | [.Name](#Name) | [.KeyPair](#KeyPair) |  |
| OpenWallet | [.Name](#Name) | [.Success](#Success) |  |
| ImportWallet | [.ImportWalletData](#ImportWalletData) | [.KeyPair](#KeyPair) |  |
| DumpWallet | [.Empty](#Empty) | [.KeyPair](#KeyPair) |  |
| CloseWallet | [.Empty](#Empty) | [.Success](#Success) |  |
| GetBalance | [.Empty](#Empty) | [.Balance](#Balance) |  |
| GetValidators | [.Empty](#Empty) | [.ValidatorsRegistry](#ValidatorsRegistry) |  |
| GetAccount | [.Empty](#Empty) | [.KeyPair](#KeyPair) |  |
| SendTransaction | [.SendTransactionInfo](#SendTransactionInfo) | [.Hash](#Hash) |  |
| StartValidator | [.KeyPair](#KeyPair) | [.Success](#Success) |  |
| StartValidatorBulk | [.KeyPairs](#KeyPairs) | [.Success](#Success) |  |
| ExitValidator | [.KeyPair](#KeyPair) | [.Success](#Success) |  |
| ExitValidatorBulk | [.KeyPairs](#KeyPairs) | [.Success](#Success) |  |

 



## Scalar Value Types

| .proto Type | Notes | C++ | Java | Python | Go | C# | PHP | Ruby |
| ----------- | ----- | --- | ---- | ------ | -- | -- | --- | ---- |
| <a name="double" /> double |  | double | double | float | float64 | double | float | Float |
| <a name="float" /> float |  | float | float | float | float32 | float | float | Float |
| <a name="int32" /> int32 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint32 instead. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="int64" /> int64 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint64 instead. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="uint32" /> uint32 | Uses variable-length encoding. | uint32 | int | int/long | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="uint64" /> uint64 | Uses variable-length encoding. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum or Fixnum (as required) |
| <a name="sint32" /> sint32 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int32s. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sint64" /> sint64 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int64s. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="fixed32" /> fixed32 | Always four bytes. More efficient than uint32 if values are often greater than 2^28. | uint32 | int | int | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="fixed64" /> fixed64 | Always eight bytes. More efficient than uint64 if values are often greater than 2^56. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum |
| <a name="sfixed32" /> sfixed32 | Always four bytes. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sfixed64" /> sfixed64 | Always eight bytes. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="bool" /> bool |  | bool | boolean | boolean | bool | bool | boolean | TrueClass/FalseClass |
| <a name="string" /> string | A string must always contain UTF-8 encoded or 7-bit ASCII text. | string | String | str/unicode | string | string | string | String (UTF-8) |
| <a name="bytes" /> bytes | May contain any arbitrary sequence of bytes. | string | ByteString | str | []byte | ByteString | string | String (ASCII-8BIT) |

