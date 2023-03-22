# 1sat-server


## Ordinal Indexing
*Not yet functioning.*

Ordinal Indexing is a means of walking back the blockchain to determin a unique serial number (ordinal) assigned to a single satoshi. Details of ordinals can be found here: [https://docs.ordinals.com/]

Indexing Ordinals can be a heavy process without some specialized indexing of the blockchain. These indexes will be built in the near future

## 1Sat Origin Indexing
The BSV blockchain is unique amount blockchains which support ordinals, in that BSV supports single satoshi outputs. This allows us to take some short-cuts in indexing until a full ordinal indexer can be built efficiently. 

Since ordinals are a unique serial number for each satoshi, an `origin` can be defined as the first outpoint where a satoshi exists alone, in a one satoshi output. Each subsequent spend of that satoshi will be crawled back only to the first ancestor where the origin has already been identified, or until it's ancestor is an output which contains more than one satoshi.

If a satoshi is subsequently packaged up in an output of more than one satoshi, the origin is no longer carried forward. If the satoshi is later spent into another one satoshi output, a new origin will be created. Both of these origins would be the same ordinal, and when the ordinal indexer is complete, both those origins will be identified as both being the same ordinal.

### Environment Variables
- POSTGRES=`<postgres connection string>`
- LISTEN=`<IP>`:`<PORT>` # defaults to 0.0.0.0:8080

## Run Web Server
```
go build
./server
```

## APIs

#### Get UTXOs
GET `/api/utxos/address/:address`
GET `/api/utxos/lock/:lock`
```
[
    {
        "txid": "13899501db55c2c0d9a79b6fe0a84eac9d68a8e1b9971b05bfab8511608bd009",
        "vout": 0,
        "satoshis": 1,
        "acc_sats": 1,
        "lock": "dab3a9eecb41663021b01755fa924332a922e026b3669823b40e05e8689a7005",
        "origin": "13899501db55c2c0d9a79b6fe0a84eac9d68a8e1b9971b05bfab8511608bd009_0",
        "ordinal": 0
    }
]
```

#### Get Inscriptions
GET `/api/inscriptions/:origin`
```
[
    {
        "id": 1318,
        "txid": "13899501db55c2c0d9a79b6fe0a84eac9d68a8e1b9971b05bfab8511608bd009",
        "vout": 0,
        "file": {
            "hash": "30c6198aeb8e94eeb7ac3aeb13bccd6735a0b6384cac19f4a88921913be0ba15",
            "size": 10133,
            "type": "image/png"
        },
        "origin": "13899501db55c2c0d9a79b6fe0a84eac9d68a8e1b9971b05bfab8511608bd009_0",
        "ordinal": 0,
        "height": 783968,
        "idx": 3434,
        "lock": "12P+UJi7rzav7yt4NHK/nleJU2YYmjoE4TG3wQRZ5HQ="
    }
]
```



#### GET `/api/files/inscriptions/:origin`
- Returns inscribed file






