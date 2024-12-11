package types

import (
	"database/sql"
	"time"
)

type MonitorObj struct {
	//Moniker         string `db:"moniker"`
	OperatorAddr    string `db:"operator_addr"`
	OperatorAddrHex string `db:"operator_addr_hex"`
	//SelfStakeAddr   string `db:"self_stake_addr"`
}

type DatabaseConfig struct {
	Username string
	Password string
	Name     string
	Host     string
	Port     string
}
type CaredData struct {
	ValSigns       []*ValSign
	ValSignMisseds []*ValSignMissed
	InactiveVal    []*MonitorObj
}
type ValIsActive struct {
	OperatorAddr string `db:"operator_addr"`
	Status       int32  `db:"status"`
}

type ValStatus struct {
	Jsonrpc string          `json:"jsonrpc"`
	Id      int             `json:"id"`
	Result  ValStatusResult `json:"result"`
}

type ValStatusResult struct {
	BlockHeight string      `json:"block_height"`
	Validators  []Validator `json:"validators"`
	Count       string      `json:"count"`
	Total       string      `json:"total"`
}

type Validator struct {
	Address          string `json:"address"`
	PubKey           PubKey `json:"pub_key"`
	VotingPower      string `json:"voting_power"`
	ProposerPriority string `json:"proposer_priority"`
}
type ValSignNum struct {
	OperatorAddr string `db:"operator_addr"`
	SignNum      int    `db:"sign_num"`
}
type MaxBlockHeight struct {
	MaxBlockHeightSign   sql.NullInt64 `db:"max_block_height_sign"`
	MaxBlockHeightMissed sql.NullInt64 `db:"max_block_height_missed"`
}

type ValSign struct {
	OperatorAddr string `db:"operator_addr"`
	BlockHeight  int64  `db:"block_height"`
	Status       int    `db:"status"`
	ChildTable   int64  `db:"child_table"`
}

type ValSignMissed struct {
	OperatorAddr string `db:"operator_addr"`
	BlockHeight  int64  `db:"block_height"`
}

type CommitBlock struct {
	Jsonrpc string      `json:"jsonrpc"`
	Id      int         `json:"id"`
	Result  BlockResult `json:"result"`
}

type BlockResult struct {
	SignedHeader BlockSignedHeader `json:"signed_header"`
	Canonical    bool              `json:"canonical"`
}

type BlockSignedHeader struct {
	Header BlockHeader `json:"header"`
	Commit BlockCommit `json:"commit"`
}

type BlockHeader struct {
	Version            Version   `json:"version"`
	ChainId            string    `json:"chain_id"`
	Height             string    `json:"height"`
	Time               time.Time `json:"time"`
	LastBlockId        BlockID   `json:"last_block_id"`
	LastCommitHash     string    `json:"last_commit_hash"`
	DataHash           string    `json:"data_hash"`
	ValidatorsHash     string    `json:"validators_hash"`
	NextValidatorsHash string    `json:"next_validators_hash"`
	ConsensusHash      string    `json:"consensus_hash"`
	AppHash            string    `json:"app_hash"`
	LastResultsHash    string    `json:"last_results_hash"`
	EvidenceHash       string    `json:"evidence_hash"`
	ProposerAddress    string    `json:"proposer_address"`
}

type Version struct {
	Block string `json:"block"`
}

type BlockID struct {
	Hash  string     `json:"hash"`
	Parts BlockParts `json:"parts"`
}

type BlockParts struct {
	Total int    `json:"total"`
	Hash  string `json:"hash"`
}

type BlockCommit struct {
	Height     string           `json:"height"`
	Round      int              `json:"round"`
	BlockID    BlockID          `json:"block_id"`
	Signatures []BlockSignature `json:"signatures"`
}

type BlockSignature struct {
	BlockIdFlag      int       `json:"block_id_flag"`
	ValidatorAddress string    `json:"validator_address"`
	Timestamp        time.Time `json:"timestamp"`
	Signature        string    `json:"signature"`
}

type ChainStatus struct {
	Jsonrpc string            `json:"jsonrpc"`
	Id      int               `json:"id"`
	Result  ChainStatusResult `json:"result"`
}

type ChainStatusResult struct {
	NodeInfo      NodeInfo      `json:"node_info"`
	SyncInfo      SyncInfo      `json:"sync_info"`
	ValidatorInfo ValidatorInfo `json:"validator_info"`
}

type NodeInfo struct {
	ProtocolVersion ProtocolVersion `json:"protocol_version"`
	Id              string          `json:"id"`
	ListenAddr      string          `json:"listen_addr"`
	Network         string          `json:"network"`
	Version         string          `json:"version"`
	Channels        string          `json:"channels"`
	Moniker         string          `json:"moniker"`
	Other           Other           `json:"other"`
}

type ProtocolVersion struct {
	P2P   string `json:"p2p"`
	Block string `json:"block"`
	App   string `json:"app"`
}

type Other struct {
	TxIndex    string `json:"tx_index"`
	RpcAddress string `json:"rpc_address"`
}

type SyncInfo struct {
	LatestBlockHash     string    `json:"latest_block_hash"`
	LatestAppHash       string    `json:"latest_app_hash"`
	LatestBlockHeight   string    `json:"latest_block_height"`
	LatestBlockTime     time.Time `json:"latest_block_time"`
	EarliestBlockHash   string    `json:"earliest_block_hash"`
	EarliestAppHash     string    `json:"earliest_app_hash"`
	EarliestBlockHeight string    `json:"earliest_block_height"`
	EarliestBlockTime   time.Time `json:"earliest_block_time"`
	CatchingUp          bool      `json:"catching_up"`
}

type ValidatorInfo struct {
	Address     string `json:"address"`
	PubKey      PubKey `json:"pub_key"`
	VotingPower string `json:"voting_power"`
}

type PubKey struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

/*type Block struct {
	Jsonrpc string      `json:"jsonrpc"`
	ID      int         `json:"id"`
	Result  interface{} `json:"result"`
}*/

/*type Block struct {
	Jsonrpc string `json:"jsonrpc"`
	ID      int    `json:"id"`
	Result  struct {
		BlockID struct {
			Hash  string `json:"hash"`
			Parts struct {
				Total int    `json:"total"`
				Hash  string `json:"hash"`
			} `json:"parts"`
		} `json:"block_id"`
		Block struct {
			Header struct {
				Version struct {
					Block string `json:"block"`
				} `json:"version"`
				ChainID     string    `json:"chain_id"`
				Height      string    `json:"height"`
				Time        time.Time `json:"time"`
				LastBlockID struct {
					Hash  string `json:"hash"`
					Parts struct {
						Total int    `json:"total"`
						Hash  string `json:"hash"`
					} `json:"parts"`
				} `json:"last_block_id"`
				LastCommitHash     string `json:"last_commit_hash"`
				DataHash           string `json:"data_hash"`
				ValidatorsHash     string `json:"validators_hash"`
				NextValidatorsHash string `json:"next_validators_hash"`
				ConsensusHash      string `json:"consensus_hash"`
				AppHash            string `json:"app_hash"`
				LastResultsHash    string `json:"last_results_hash"`
				EvidenceHash       string `json:"evidence_hash"`
				ProposerAddress    string `json:"proposer_address"`
			} `json:"header"`
			Data struct {
				Txs []string `json:"txs"`
			} `json:"data"`
			Evidence struct {
				Evidence []interface{} `json:"evidence"`
			} `json:"evidence"`
			LastCommit struct {
				Height  string `json:"height"`
				Round   int    `json:"round"`
				BlockID struct {
					Hash  string `json:"hash"`
					Parts struct {
						Total int    `json:"total"`
						Hash  string `json:"hash"`
					} `json:"parts"`
				} `json:"block_id"`
				Signatures []struct {
					BlockIDFlag      int       `json:"block_id_flag"`
					ValidatorAddress string    `json:"validator_address"`
					Timestamp        time.Time `json:"timestamp"`
					Signature        string    `json:"signature"`
				} `json:"signatures"`
			} `json:"last_commit"`
		} `json:"block"`
	} `json:"result"`
}*/
