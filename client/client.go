package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type HTTPError struct {
	StatusCode int
	Status     string
	Body       []byte
}

func (h *HTTPError) Error() string {
	return fmt.Sprintf("(%s) %s", h.Status, string(h.Body))
}

type BlockHeader struct {
	Protocol                  string    `json:"protocol"`
	ChainID                   string    `json:"chain_id"`
	Hash                      string    `json:"hash"`
	Level                     int64     `json:"level"`
	Proto                     uint      `json:"proto"`
	Predecessor               string    `json:"predecessor"`
	Timestamp                 time.Time `json:"timestamp"`
	ValidationPass            uint      `json:"validation_pass"`
	OperationsHash            string    `json:"operations_hash"`
	Fitness                   []Bytes   `json:"fitness"`
	Context                   string    `json:"context"`
	Priority                  uint      `json:"priority"`
	ProofOfWorkNonce          Bytes     `json:"proof_of_work_nonce"`
	SeedNonceHash             string    `json:"seed_nonce_hash"`
	LiquidityBakingEscapeVote bool      `json:"liquidity_baking_escape_vote"`
	Signature                 string    `json:"signature"`
}

type BlockContextConstants struct {
	ProofOfWorkNonceSize              uint       `json:"proof_of_work_nonce_size"`
	NonceLength                       uint       `json:"nonce_length"`
	MaxAnonOpsPerBlock                uint       `json:"max_anon_ops_per_block"`
	MaxOperationDataLength            int64      `json:"max_operation_data_length"`
	MaxProposalsPerDelegate           uint       `json:"max_proposals_per_delegate"`
	PreservedCycles                   uint       `json:"preserved_cycles"`
	BlocksPerCycle                    int64      `json:"blocks_per_cycle"`
	BlocksPerCommitment               int64      `json:"blocks_per_commitment"`
	BlocksPerRollSnapshot             int64      `json:"blocks_per_roll_snapshot"`
	BlocksPerVotingPeriod             int64      `json:"blocks_per_voting_period"`
	TimeBetweenBlocks                 Int64Array `json:"time_between_blocks"`
	EndorsersPerBlock                 uint       `json:"endorsers_per_block"`
	HardGasLimitPerOperation          *BigInt    `json:"hard_gas_limit_per_operation"`
	HardGasLimitPerBlock              *BigInt    `json:"hard_gas_limit_per_block"`
	ProofOfWorkThreshold              int64      `json:"proof_of_work_threshold,string"`
	TokensPerRoll                     *BigInt    `json:"tokens_per_roll"`
	MichelsonMaximumTypeSize          uint       `json:"michelson_maximum_type_size"`
	SeedNonceRevelationTip            *BigInt    `json:"seed_nonce_revelation_tip"`
	OriginationSize                   int64      `json:"origination_size"`
	BlockSecurityDeposit              *BigInt    `json:"block_security_deposit"`
	EndorsementSecurityDeposit        *BigInt    `json:"endorsement_security_deposit"`
	BakingRewardPerEndorsement        []*BigInt  `json:"baking_reward_per_endorsement"`
	EndorsementReward                 []*BigInt  `json:"endorsement_reward"`
	CostPerByte                       *BigInt    `json:"cost_per_byte"`
	HardStorageLimitPerOperation      *BigInt    `json:"hard_storage_limit_per_operation"`
	QuorumMin                         int64      `json:"quorum_min"`
	QuorumMax                         int64      `json:"quorum_max"`
	MinProposalQuorum                 int64      `json:"min_proposal_quorum"`
	InitialEndorsers                  uint       `json:"initial_endorsers"`
	DelayPerMissingEndorsement        int64      `json:"delay_per_missing_endorsement,string"`
	MinimalBlockDelay                 int64      `json:"minimal_block_delay,string"`
	LiquidityBakingSubsidy            *BigInt    `json:"liquidity_baking_subsidy"`
	LiquidityBakingSunsetLevel        int64      `json:"liquidity_baking_sunset_level"`
	LiquidityBakingEscapeEMAThreshold int64      `json:"liquidity_baking_escape_ema_threshold"`
}

type Client struct {
	URL    string
	Chain  string
	Client *http.Client
}

func (c *Client) chain() string {
	if c.Chain != "" {
		return c.Chain
	}
	return "main"
}

func (c *Client) client() *http.Client {
	if c.Client != nil {
		return c.Client
	}
	return http.DefaultClient
}

func (c *Client) do(r *http.Request) (io.ReadCloser, error) {
	res, err := c.client().Do(r)
	if err != nil {
		return nil, err
	}

	if res.StatusCode/100 != 2 {
		defer res.Body.Close()
		body, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}
		return nil, &HTTPError{
			StatusCode: res.StatusCode,
			Status:     res.Status,
			Body:       body,
		}
	}

	return res.Body, nil
}

func (c *Client) NewGetBlockHeaderRequest(id string) (*http.Request, error) {
	url := fmt.Sprintf("%s/chains/%s/blocks/%s/header", c.URL, c.chain(), id)
	return http.NewRequest("GET", url, nil)
}

func (c *Client) GetBlockHeader(id string) (*BlockHeader, error) {
	req, err := c.NewGetBlockHeaderRequest(id)
	if err != nil {
		return nil, err
	}
	res, err := c.do(req)
	if err != nil {
		return nil, err
	}
	defer res.Close()

	var v BlockHeader
	if err := json.NewDecoder(res).Decode(&v); err != nil {
		return nil, err
	}
	return &v, nil
}

func (c *Client) NewGetBlockContextConstantsRequest(id string) (*http.Request, error) {
	url := fmt.Sprintf("%s/chains/%s/blocks/%s/context/constants", c.URL, c.chain(), id)
	return http.NewRequest("GET", url, nil)
}

func (c *Client) GetBlockContextConstants(id string) (*BlockContextConstants, error) {
	req, err := c.NewGetBlockContextConstantsRequest(id)
	if err != nil {
		return nil, err
	}
	res, err := c.do(req)
	if err != nil {
		return nil, err
	}
	defer res.Close()

	var v BlockContextConstants
	if err := json.NewDecoder(res).Decode(&v); err != nil {
		return nil, err
	}
	return &v, nil
}
