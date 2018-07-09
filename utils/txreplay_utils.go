/*
 * Copyright (C) 2018 The ontology Authors
 * This file is part of The ontology library.
 *
 * The ontology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ontology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology.  If not, see <http://www.gnu.org/licenses/>.
 */

package utils

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	//"os"
	"strings"
	"time"

	//"github.com/ontio/ontology-crypto/keypair"
	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/config"
	vbft "github.com/ontio/ontology/consensus/vbft"
	"github.com/ontio/ontology/consensus/vbft/config"
	"github.com/ontio/ontology/core/genesis"
	"github.com/ontio/ontology/core/ledger"
	"github.com/ontio/ontology/core/signature"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract/service/native/governance"
)

var (
	ip   string
	port uint
)

const JSON_RPC_VERSION = "2.0"

//JsonRpcRequest object in rpc
type JsonRpcRequest struct {
	Version string        `json:"jsonrpc"`
	Id      string        `json:"id"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

//JsonRpcResponse object response for JsonRpcRequest
type JsonRpcResponse struct {
	Error  int64           `json:"error"`
	Desc   string          `json:"desc"`
	Result json.RawMessage `json:"result"`
}

func SetIPPort(hostIP string, hostPort uint) {
	ip = hostIP
	port = hostPort
}

func rpcAddress() string {
	address := fmt.Sprintf("http://%s:%d", ip, port)
	return address
}

func sendRpcRequest(method string, params []interface{}) ([]byte, error) {
	rpcReq := &JsonRpcRequest{
		Version: JSON_RPC_VERSION,
		Id:      "cli",
		Method:  method,
		Params:  params,
	}
	data, err := json.Marshal(rpcReq)
	if err != nil {
		return nil, fmt.Errorf("JsonRpcRequest json.Marsha error:%s", err)
	}

	addr := rpcAddress()
	resp, err := http.Post(addr, "application/json", strings.NewReader(string(data)))
	if err != nil {
		return nil, fmt.Errorf("http post request:%s error:%s", data, err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read rpc response body error:%s", err)
	}
	rpcRsp := &JsonRpcResponse{}
	err = json.Unmarshal(body, rpcRsp)
	if err != nil {
		return nil, fmt.Errorf("json.Unmarshal JsonRpcResponse:%s error:%s", body, err)
	}
	if rpcRsp.Error != 0 {
		return nil, fmt.Errorf("error code:%d desc:%s", rpcRsp.Error, rpcRsp.Desc)
	}
	return rpcRsp.Result, nil
}

func GetBlockCount() (uint32, error) {
	data, err := sendRpcRequest("getblockcount", []interface{}{})
	if err != nil {
		return 0, err
	}
	num := uint32(0)
	err = json.Unmarshal(data, &num)
	if err != nil {
		return 0, fmt.Errorf("json.Unmarshal:%s ss error:%s", data, err)
	}
	return num, nil
}

func GetBlockData(hashOrHeight interface{}) ([]byte, error) {
	data, err := sendRpcRequest("getblock", []interface{}{hashOrHeight})
	if err != nil {
		return nil, err
	}
	hexStr := ""
	err = json.Unmarshal(data, &hexStr)
	if err != nil {
		return nil, fmt.Errorf("json.Unmarshal sff error:%s", err)
	}
	blockData, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil, fmt.Errorf("hex.DecodeString error:%s", err)
	}
	return blockData, nil
}

func SendRawTransaction(tx string) error {
	_, err := sendRpcRequest("sendrawtransaction", []interface{}{tx})
	return err
}

func initVbftBlock(block *types.Block) (*vbft.Block, error) {
	if block == nil {
		return nil, fmt.Errorf("nil block in initVbftBlock")
	}

	blkInfo := &vconfig.VbftBlockInfo{}
	if err := json.Unmarshal(block.Header.ConsensusPayload, blkInfo); err != nil {
		return nil, fmt.Errorf("unmarshal blockInfo: %s", err)
	}

	return &vbft.Block{
		Block: block,
		Info:  blkInfo,
	}, nil
}

func getConsensusPayload(blk *types.Block) ([]byte, error) {
	block, err := initVbftBlock(blk)
	if err != nil {
		return nil, err
	}
	lastConfigBlkNum := block.Info.LastConfigBlockNum
	if block.Info.NewChainConfig != nil {
		lastConfigBlkNum = block.Block.Header.Height
	}
	vbftBlkInfo := &vconfig.VbftBlockInfo{
		Proposer:           math.MaxUint32,
		LastConfigBlockNum: lastConfigBlkNum,
		NewChainConfig:     nil,
	}
	consensusPayload, err := json.Marshal(vbftBlkInfo)
	if err != nil {
		return nil, err
	}
	return consensusPayload, nil
}

func ConstructBlock(accounts []*account.Account, ldg *ledger.Ledger, blkNum uint32,
	preBlock *types.Block, txs []*types.Transaction) (*types.Block, error) {
	consensusPayload, err := getConsensusPayload(preBlock)
	if err != nil {
		return nil, fmt.Errorf("failed to get consensus payload %v", err)
	}
	blockTimestamp := uint32(time.Now().Unix())
	if preBlock.Header.Timestamp >= blockTimestamp {
		blockTimestamp = preBlock.Header.Timestamp + 1
	}

	txHash := []common.Uint256{}
	for _, t := range txs {
		txHash = append(txHash, t.Hash())
	}
	txRoot := common.ComputeMerkleRoot(txHash)
	blockRoot := ldg.GetBlockRootWithNewTxRoot(txRoot)

	blkHeader := &types.Header{
		PrevBlockHash:    preBlock.Hash(),
		TransactionsRoot: txRoot,
		BlockRoot:        blockRoot,
		Timestamp:        blockTimestamp,
		Height:           uint32(blkNum),
		ConsensusData:    common.GetNonce(),
		ConsensusPayload: consensusPayload,
	}
	blk := &types.Block{
		Header:       blkHeader,
		Transactions: txs,
	}
	blkHash := blk.Hash()
	for _, account := range accounts {
		sig, err := signature.Sign(account, blkHash[:])
		if err != nil {
			return nil, fmt.Errorf("sign block failed, block hash：%x, error: %s", blkHash, err)
		}
		blkHeader.Bookkeepers = append(blkHeader.Bookkeepers, account.PublicKey)
		blkHeader.SigData = append(blkHeader.SigData, sig)
	}

	return blk, nil
}

func setGenesis(genesisFile string, cfg *config.GenesisConfig) error {
	if !common.FileExisted(genesisFile) {
		return nil
	}

	fmt.Printf("Load genesis config:%s\n", genesisFile)
	data, err := ioutil.ReadFile(genesisFile)
	if err != nil {
		return fmt.Errorf("ioutil.ReadFile:%s error:%s", genesisFile, err)
	}
	// Remove the UTF-8 Byte Order Mark
	data = bytes.TrimPrefix(data, []byte("\xef\xbb\xbf"))

	cfg.Reset()
	err = json.Unmarshal(data, cfg)
	if err != nil {
		return fmt.Errorf("json.Unmarshal GenesisConfig:%s error:%s", data, err)
	}
	switch cfg.ConsensusType {
	case config.CONSENSUS_TYPE_DBFT:
		if len(cfg.DBFT.Bookkeepers) < config.DBFT_MIN_NODE_NUM {
			return fmt.Errorf("DBFT consensus at least need %d bookkeepers in config", config.DBFT_MIN_NODE_NUM)
		}
		if cfg.DBFT.GenBlockTime <= 0 {
			cfg.DBFT.GenBlockTime = config.DEFAULT_GEN_BLOCK_TIME
		}
	case config.CONSENSUS_TYPE_VBFT:
		err = governance.CheckVBFTConfig(cfg.VBFT)
		if err != nil {
			return fmt.Errorf("VBFT config error %v", err)
		}
		if len(cfg.VBFT.Peers) < config.VBFT_MIN_NODE_NUM {
			return fmt.Errorf("VBFT consensus at least need %d peers in config", config.VBFT_MIN_NODE_NUM)
		}
	default:
		return fmt.Errorf("Unknow consensus:%s", cfg.ConsensusType)
	}

	return nil
}

func InitConfig(genesisFile string, networkId int) (*config.OntologyConfig, error) {
	cfg := config.DefConfig
	switch networkId {
	case config.NETWORK_ID_MAIN_NET:
		cfg.Genesis = config.MainNetConfig
	case config.NETWORK_ID_POLARIS_NET:
		cfg.Genesis = config.PolarisConfig
	}
	if genesisFile == "" {
		return cfg, nil
	}

	err := setGenesis(genesisFile, cfg.Genesis)
	if err != nil {
		return nil, fmt.Errorf("setGenesis error:%s", err)
	}

	fmt.Println("Config init success")
	return cfg, nil
}

func getDefaultAccounts(config TxReplayConfig) ([]*account.Account, error) {
	accounts := make([]*account.Account, 0, len(config.Wallets))
	for _, wallet := range config.Wallets {
		if !common.FileExisted(wallet.Path) {
			return nil, fmt.Errorf("cannot find wallet file:%s", wallet.Path)
		}

		client, err := account.Open(wallet.Path)
		if err != nil {
			return nil, fmt.Errorf("failed to open wallet %s, err %v", wallet.Path, err)
		}
		user, err := client.GetDefaultAccount([]byte(wallet.Password))
		if err != nil {
			return nil, fmt.Errorf("failed to get default account err %v", err)
		}
		accounts = append(accounts, user)
	}
	return accounts, nil
}

func InitAccounts() ([]*account.Account, error) {
	walletCfg := TxReplayConfig{
		Wallets: make([]walletConfig, 0),
	}

	err := walletCfg.loadConfig("./wallets.json")
	if err != nil {

		return nil, err
	}

	return getDefaultAccounts(walletCfg)
}

func InitLedger(cfg *config.OntologyConfig, networkId int) (*ledger.Ledger, error) {
	var err error

	networkName := config.GetNetworkName(uint32(networkId))
	dbDir := fmt.Sprintf("./Chain/%s", networkName)

	ledger.DefLedger, err = ledger.NewLedger(dbDir)
	if err != nil {
		fmt.Println(err)
		return nil, fmt.Errorf("NewLedger error:%s", err)
	}
	bookKeepers, err := cfg.GetBookkeepers()
	if err != nil {
		return nil, fmt.Errorf("GetBookkeepers error:%s", err)
	}

	genesisBlock, err := genesis.BuildGenesisBlock(bookKeepers, cfg.Genesis)
	if err != nil {
		return nil, fmt.Errorf("genesisBlock error %s", err)
	}
	err = ledger.DefLedger.Init(bookKeepers, genesisBlock)
	if err != nil {
		return nil, fmt.Errorf("Init ledger error:%s", err)
	}

	fmt.Println("Ledger init success")
	return ledger.DefLedger, nil
}
