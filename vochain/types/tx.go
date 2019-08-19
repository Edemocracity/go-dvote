package vochain

import (
	"errors"
	"fmt"
	"strconv"

	tmtypes "github.com/tendermint/tendermint/types"
)

// ________________________ TX ________________________

// Tx represents a raw Tx that has a method and an args keys
type Tx struct {
	Method string            `json:"method"`
	Args   map[string]string `json:"args"`
}

// ValidateMethod returns true if the method is defined in the TxMethod enum
func (tx Tx) ValidateMethod() TxMethod {
	m := tx.Method
	switch m {
	case "newProcessTx":
		return NewProcessTx
	case "voteTx":
		return VoteTx
	case "addTrustedOracleTx":
		return AddTrustedOracleTx
	case "removeTrustedOracleTx":
		return RemoveTrustedOracleTx
	case "addValidatorTx":
		return AddValidatorTx
	case "removeValidatorTx":
		return RemoveValidatorTx
	default:
		return ""
	}
}

// ________________________ VALID TX ________________________

// ValidTx represents a Tx with a valid method and valid args for the method
type ValidTx struct {
	Method TxMethod `json:"method"`
	Args   TxArgs   `json:"args"`
}

// String converets a ValidTx struct to a human easy readable string format
func (vtx ValidTx) String() string {
	return fmt.Sprintf(`{ "method": %s, "args": %s }`, vtx.Method.String(), vtx.Args)
}

// ________________________ TX METHODS ________________________

// TxMethod is a string representing the allowed methods in the Vochain paradigm
type TxMethod string

const (
	// NewProcessTx is the method name for init a new process
	NewProcessTx TxMethod = "newProcessTx"
	// VoteTx is the method name for casting a vote
	VoteTx TxMethod = "voteTx"
	// AddTrustedOracleTx is the method name for adding a new Census Manager
	AddTrustedOracleTx TxMethod = "addTrustedOracleTx"
	// RemoveTrustedOracleTx is the method name for removing an existing Census Manager
	RemoveTrustedOracleTx TxMethod = "removeTrustedOracleTx"
	// AddValidatorTx is the method name for adding a new validator address in the consensusParams validator list
	AddValidatorTx TxMethod = "addvalidatortx"
	// RemoveValidatorTx is the method name for removing an existing validator address in the consensusParams validator list
	RemoveValidatorTx TxMethod = "removeValidatorTx"
)

// String returns the CurrentProcessState as string
func (m TxMethod) String() string {
	return fmt.Sprintf("%s", string(m))
}

// ________________________ TX ARGS ________________________

// TxArgs generic interface to address valid method args
type TxArgs interface{}

// NewProcessTxArgs represents the data required in order to start a new process
type NewProcessTxArgs struct {
	// EntityID the process belongs to
	EntityID       string `json:"entityid"`
	EntityResolver string `json:"entityresolver"`
	// MetadataHash hash of the entity metadata
	MetadataHash string `json:"metadatahash"`
	// MkRoot merkle root of all the census in the process
	MkRoot string `json:"mkroot"`
	// NumberOfBlocks represents the tendermint block where the process goes from active to finished
	NumberOfBlocks int64 `json:"numberofblocks"`
	// InitBlock represents the tendermint block where the process goes from scheduled to active
	InitBlock int64 `json:"initblock"`
	// EncryptionKeys are the keys required to encrypt the votes
	//EncryptionKeys []string `json:"encryptionkeys"`
	EncryptionKeys string `json:"encryptionkeys"`
}

func (n NewProcessTxArgs) String() string {
	last := fmt.Sprintf(`"mkroot": %v, "initblock": %v, "numberofblocks": %v, "encryptionkeys": %v  }`, n.MkRoot, n.InitBlock, n.NumberOfBlocks, n.EncryptionKeys)
	return fmt.Sprintf(`{ "entityid": %v, "entityresolver": %v, "metadatahash": %v, %v`, n.EntityID, n.EntityResolver, n.MetadataHash, last)
}

// VoteTxArgs represents the data required in order to cast a vote
type VoteTxArgs struct {
	Nullifier   string `json:"nullifier"`
	Payload     string `json:"payload"`
	CensusProof string `json:"censusproof"`
}

func (n VoteTxArgs) String() string {
	return fmt.Sprintf(`{ "nullifier": %v, "payload": %v, "censusProof": %v }`, n.Nullifier, n.Payload, n.CensusProof)
}

// AddTrustedOracleTxArgs represents the data required in order to add a new trusted oracle
type AddTrustedOracleTxArgs struct {
	Address tmtypes.Address `json:"address"`
}

func (n AddTrustedOracleTxArgs) String() string {
	return fmt.Sprintf(`{ "address": %v }`, n.Address)
}

// RemoveTrustedOracleTxArgs represents the data required in order to remove an existing trusted oracle
type RemoveTrustedOracleTxArgs struct {
	Address tmtypes.Address `json:"address"`
}

func (n RemoveTrustedOracleTxArgs) String() string {
	return fmt.Sprintf(`{ "address": %v }`, n.Address)
}

// AddValidatorTxArgs represents the data required in order to add a new validator node
type AddValidatorTxArgs struct {
	Address tmtypes.Address `json:"address"`
	Power   uint64          `json:"power"`
}

func (n AddValidatorTxArgs) String() string {
	return fmt.Sprintf(`{ "address": %v, "power": %v }`, n.Address, n.Power)
}

// RemoveValidatorTxArgs represents the data required in order to remove an existing validator node
type RemoveValidatorTxArgs struct {
	Address tmtypes.Address `json:"address"`
}

func (n RemoveValidatorTxArgs) String() string {
	return fmt.Sprintf(`{ "address": %v }`, n.Address)
}

// ValidateArgs does a sanity check onto the arguments passed to a valid TxMethod
func (tx Tx) ValidateArgs() (TxArgs, error) {
	var t TxArgs
	switch tx.Method {
	case "newProcessTx":
		// invalid length
		if len(tx.Args) != 7 {
			return nil, errors.New("Invalid args number")
		}
		// invalid args
		if _, ok := tx.Args["entityId"]; ok {
			if _, ok = tx.Args["entityResolver"]; ok {
				if _, ok = tx.Args["metadataHash"]; ok {
					if _, ok = tx.Args["mkRoot"]; ok {
						if _, ok = tx.Args["encryptionKeys"]; ok {
							nblocks, err := strconv.ParseInt(tx.Args["numberOfBlocks"], 10, 64)
							if err == nil {
								iblock, err := strconv.ParseInt(tx.Args["initBlock"], 10, 64)
								if err == nil {
									t = NewProcessTxArgs{
										EntityID:       tx.Args["entityId"],
										EntityResolver: tx.Args["entityResolver"],
										MetadataHash:   tx.Args["metadataHash"],
										MkRoot:         tx.Args["mkRoot"],
										NumberOfBlocks: nblocks,
										InitBlock:      iblock,
										EncryptionKeys: tx.Args["encryptionKeys"],
									}
								} else {
									return nil, errors.New("initBlock arg not found")
								}
							} else {
								return nil, errors.New("numberOfBlocks arg not found")
							}
						} else {
							return nil, errors.New("encryptionKeys arg not found")
						}
					} else {
						return nil, errors.New("mkRoot arg not found")
					}
				} else {
					return nil, errors.New("metadataHash arg not found")
				}
			} else {
				return nil, errors.New("entityResolver arg not found")
			}
		} else {
			return nil, errors.New("entityId arg not found")
		}
		// sanity check done
		return t, nil
	case "voteTx":
		// invalid length
		if len(tx.Args) != 3 {
			return nil, errors.New("Invalid args number")
		}
		// invalid args
		if _, ok := tx.Args["nullifier"]; ok {
			if _, ok = tx.Args["payload"]; ok {
				if _, ok = tx.Args["censusProof"]; ok {
					// VoteTxArgs can be created
					t = VoteTxArgs{
						Nullifier:   tx.Args["nullifier"],
						Payload:     tx.Args["payload"],
						CensusProof: tx.Args["censusProof"],
					}
				} else {
					return nil, errors.New("censusProof arg not found")
				}
			} else {
				return nil, errors.New("payload arg not found")
			}
		} else {
			return nil, errors.New("nullifier arg not found")
		}
		// sanity check done
		return t, nil
	case "addTrustedOracleTx":
		// invalid length
		if len(tx.Args) != 1 {
			return nil, errors.New("Invalid args number")
		}
		// invalid args
		if _, ok := tx.Args["address"]; ok {
			// AddTrustedOracleTxArgs can be created
			t = AddTrustedOracleTxArgs{
				Address: []byte(tx.Args["address"]),
			}
		} else {
			return nil, errors.New("address arg not found")
		}
		// sanity check done
		return t, nil
	case "removeTrustedOracleTx":
		// invalid length
		if len(tx.Args) != 1 {
			return nil, errors.New("Invalid args number")
		}
		// invalid args
		if _, ok := tx.Args["address"]; ok {
			// RemoveTrustedOracleTxArgs can be created
			t = RemoveTrustedOracleTxArgs{
				Address: []byte(tx.Args["address"]),
			}
		} else {
			return nil, errors.New("address arg not found")
		}
		// sanity check done
		return t, nil
	case "addValidatorTx":
		// invalid length
		if len(tx.Args) != 1 {
			return nil, errors.New("Invalid args number")
		}
		// invalid args
		if _, ok := tx.Args["address"]; ok {
			// AddvalidatorTxArgs can be created
			t = AddValidatorTxArgs{
				Address: []byte(tx.Args["address"]),
			}
		} else {
			return nil, errors.New("address arg not found")
		}
		// sanity check done
		return t, nil
	case "removeValidatorTx":
		// invalid length
		if len(tx.Args) != 1 {
			return nil, errors.New("Invalid args number")
		}
		// invalid args
		if _, ok := tx.Args["address"]; ok {
			// RemoveValidatorTxArgs can be created
			t = RemoveValidatorTxArgs{
				Address: []byte(tx.Args["address"]),
			}
		} else {
			return nil, errors.New("address arg not found")
		}
		// sanity check done
		return t, nil
	default:
		return nil, errors.New("Cannot validate args")
	}
}