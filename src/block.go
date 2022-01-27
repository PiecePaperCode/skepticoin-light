package main

type Block struct {
	Header          BlockHeader
	LenTransactions byte
	Transactions    [1]Transaction
}
type BlockHeader struct {
	Version           byte
	Height            uint32 // Variable l q
	PreviousBlockHash [32]byte
	MerkleRootHash    [32]byte
	Timestamp         uint32
	Target            [32]byte
	Nonce             uint32
	BlockHash         [32]byte
	ChainSample       [32]byte
	SummaryHash       [32]byte
}
type Transaction struct {
	Version    byte
	LenInputs  byte // Variable l q
	Inputs     [1]Input
	LenOutputs byte // Variable l q
	Outputs    [1]Output
}
type Input struct {
	Hash      [32]byte
	Index     uint32
	Signature CoinbaseSignature
}
type Output struct {
	Value     uint64
	PublicKey [64]byte
}
type Signature struct {
	Type byte
}
type CoinbaseSignature struct {
	Type    byte
	Height  uint32
	Len     byte
	Message [0]byte
}
type SECP256k1Signature struct {
	Type      byte
	Signature [64]byte
}
