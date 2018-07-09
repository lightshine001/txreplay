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

package command

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gosuri/uiprogress"
	"github.com/urfave/cli"

	cutils "github.com/ontio/ontology/cmd/utils"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/common/serialization"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/txreplay/utils"
)

const (
	DEFAULT_CONFIG_FILE_NAME = "./config.json"
	DEFAULT_BLOCK_FILE_NAME  = "block.dat"
)

var TxExportCommand = cli.Command{
	Name:      "txexport",
	Usage:     "Export txs in DB to a file",
	ArgsUsage: "",
	Action:    exportTxs,
	Flags: []cli.Flag{
		HostIPFlag,
		RPCPortFlag,
		TxExportFileFlag,
		TxExportHeightFlag,
	},
	Description: "",
}

func exportTxs(ctx *cli.Context) error {
	ip := ctx.String(GetFlagName(HostIPFlag))
	port := ctx.Uint(GetFlagName(RPCPortFlag))
	utils.SetIPPort(ip, port)

	txFile := ctx.String(GetFlagName(TxExportFileFlag))
	if txFile == "" {
		fmt.Printf("Missing file argumen\n")
		cli.ShowSubcommandHelp(ctx)
		return nil
	}
	if common.FileExisted(txFile) {
		return fmt.Errorf("File:%s has already exist", txFile)
	}
	startHeight := ctx.Uint(GetFlagName(TxExportHeightFlag))
	blockCount, err := utils.GetBlockCount()
	if err != nil {
		return fmt.Errorf("GetBlockCount error:%s", err)
	}

	if startHeight >= uint(blockCount) {
		return fmt.Errorf("The specified height is over current height")
	}

	ef, err := os.OpenFile(txFile, os.O_RDWR|os.O_CREATE, 0664)
	if err != nil {
		return fmt.Errorf("Open file:%s error:%s", txFile, err)
	}
	defer ef.Close()
	fWriter := bufio.NewWriter(ef)

	totalBlocks := int(blockCount) - int(startHeight)
	uiprogress.Start()
	bar := uiprogress.AddBar(totalBlocks).
		AppendCompleted().
		AppendElapsed().
		PrependFunc(func(b *uiprogress.Bar) string {
			return fmt.Sprintf("Remaining Block %d", totalBlocks-b.Current())
		})

	fmt.Printf("Start export...\n")
	var count uint64
	for i := uint32(startHeight); i < uint32(blockCount); i++ {
		blockData, err := utils.GetBlockData(i)
		if err != nil {
			return fmt.Errorf("Get block:%d error:%s", i, err)
		}
		var block types.Block
		buf := bytes.NewBuffer(blockData)
		err = block.Deserialize(buf)
		if err != nil {
			return fmt.Errorf("failed to read block at height %d err %v", i, err)
		}
		fWriter.WriteString(fmt.Sprintf("Block %d num %d\n", i, len(block.Transactions)))
		for _, tx := range block.Transactions {
			txHex := hex.EncodeToString(tx.ToArray())
			_, err = fWriter.WriteString(fmt.Sprintf("%x %s\n", tx.Hash(), txHex))
			if err != nil {
				return fmt.Errorf("failed to write tx data %x at block height %d", tx.Hash(), i)
			}
			count++
		}
		bar.Incr()
	}
	uiprogress.Stop()

	err = fWriter.Flush()
	if err != nil {
		return fmt.Errorf("Export flush file error:%s", err)
	}
	fmt.Printf("Export txs successfully.\n")
	fmt.Printf("Total txs:%d from block %d to block %d\n", count, startHeight, blockCount)
	fmt.Printf("Export file:%s\n", txFile)
	return nil
}

var TxImportCommand = cli.Command{
	Name:      "tximport",
	Usage:     "Import txs from a file",
	ArgsUsage: "",
	Action:    importTxs,
	Flags: []cli.Flag{
		ImportTxFileFlag,
		NetworkIdFlag,
		TimerFlag,
	},
	Description: "",
}

func importTxs(ctx *cli.Context) {
	log.Init(log.PATH, log.Stdout)
	networkId := ctx.Int(GetFlagName(NetworkIdFlag))
	cfg, err := utils.InitConfig("", networkId)
	if err != nil {
		fmt.Println("failed to init config %v", err)
		return
	}
	accounts, err := utils.InitAccounts()

	if err != nil {
		fmt.Println(err)
		return
	}

	ldg, err := utils.InitLedger(cfg, networkId)
	if err != nil {
		fmt.Println(err)
		return
	}

	txFile := ctx.String(GetFlagName(ImportTxFileFlag))
	if txFile == "" {
		fmt.Println("Missing file argument")
		cli.ShowSubcommandHelp(ctx)
		return
	}

	ifile, err := os.OpenFile(txFile, os.O_RDONLY, 0644)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer ifile.Close()

	fReader := bufio.NewReader(ifile)

	fmt.Printf("%s Start import Txs...\n",
		time.Now().UTC().Format(time.UnixDate))

	var txs []*types.Transaction
	count := 0
	summary := 0
	errNum := 0
	delay := ctx.Uint(GetFlagName(TimerFlag))
	rateLimiter := time.NewTicker(time.Millisecond * time.Duration(delay))

	for {
		line, err := fReader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if strings.HasPrefix(line, "Block ") {
			if len(txs) != 0 {
				blockHeight := ldg.GetCurrentBlockHeight()
				preBlock, err := ldg.GetBlockByHeight(blockHeight)
				if err != nil {
					fmt.Println(err)
					return
				}

				// build a block
				blk, err := utils.ConstructBlock(accounts, ldg, blockHeight+1, preBlock, txs)
				if err != nil {
					fmt.Println(err)
					return
				}
				err = ldg.AddBlock(blk)
				if err != nil {
					fmt.Errorf("add block height:%d error:%s", blockHeight, err)
					return
				}
				summary = summary + len(txs)

				fmt.Printf("%s packed tx count %d, errNum %d, current block height %d  block hash %x\n",
					time.Now().UTC().Format(time.UnixDate), summary, errNum, blk.Header.Height,
					blk.Hash())
			}
			<-rateLimiter.C

			index := strings.LastIndex(line, " ")
			num, _ := strconv.Atoi(line[index+1 : len(line)-1])
			txs = make([]*types.Transaction, 0, num)
		} else {
			index := strings.Index(line, " ")
			if index < 0 {
				fmt.Printf("failed to split tx %s\n", line)
				errNum++
				continue
			}
			txHex := line[index+1 : len(line)-1]
			hex, err := common.HexToBytes(txHex)
			if err != nil {
				fmt.Printf("failed to convert from hex to bytes %s\n", line)
				errNum++
				continue
			}
			tx := &types.Transaction{}
			if err := tx.Deserialize(bytes.NewReader(hex)); err != nil {
				fmt.Printf("failed to deserialize tx %x\n", hex)
				errNum++
				continue
			}
			count++

			exist, err := ldg.IsContainTransaction(tx.Hash())
			if err != nil {
				fmt.Printf("Unknown error tx %x\n", tx.Hash())
				errNum++
				continue
			} else if exist {
				fmt.Printf("Duplicated input tx %x\n", tx.Hash())
				errNum++
				continue
			}
			txs = append(txs, tx)
		}
	}

	// Last block
	if len(txs) != 0 {
		blockHeight := ldg.GetCurrentBlockHeight()
		preBlock, err := ldg.GetBlockByHeight(blockHeight)
		if err != nil {
			fmt.Println(err)
			return
		}
		// build a block
		blk, err := utils.ConstructBlock(accounts, ldg, blockHeight+1, preBlock, txs)
		if err != nil {
			fmt.Println(err)
			return
		}
		err = ldg.AddBlock(blk)
		if err != nil {
			fmt.Errorf("add block height:%d error:%s", blockHeight, err)
			return
		}
		summary = summary + len(txs)
		fmt.Printf("%s packed tx count %d errNum %d,  current block height %d  block hash %x\n",
			time.Now().UTC().Format(time.UnixDate), summary, errNum, blk.Header.Height,
			blk.Hash())
	}

	fmt.Printf("%s Import Txs complete, total txs %d packed txs %d errNum %d\n",
		time.Now().UTC().Format(time.UnixDate), count, summary, errNum)
	rateLimiter.Stop()

	oFile, err := os.OpenFile(DEFAULT_BLOCK_FILE_NAME, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		fmt.Printf("failed to open file %s, err %v\n", DEFAULT_BLOCK_FILE_NAME, err)
		return
	}

	defer oFile.Close()

	fWriter := bufio.NewWriter(oFile)
	blockHeight := ldg.GetCurrentBlockHeight()

	metadata := cutils.NewExportBlockMetadata()
	metadata.BlockHeight = blockHeight
	err = metadata.Serialize(fWriter)
	if err != nil {
		fmt.Printf("Write export metadata error:%s\n", err)
		return
	}

	//progress bar
	uiprogress.Start()
	bar := uiprogress.AddBar(int(blockHeight)).
		AppendCompleted().
		AppendElapsed().
		PrependFunc(func(b *uiprogress.Bar) string {
			return fmt.Sprintf("Block(%d/%d)", b.Current(), int(blockHeight))
		})

	fmt.Printf("Start export block.\n")
	for i := uint32(0); i <= blockHeight; i++ {
		block, err := ldg.GetBlockByHeight(i)
		if err != nil {
			fmt.Println(err)
			return
		}

		w := bytes.NewBuffer(nil)
		block.Serialize(w)

		data, err := cutils.CompressBlockData(w.Bytes(), metadata.CompressType)
		if err != nil {
			fmt.Printf("Compress block height:%d error:%s\n", i, err)
			return
		}
		err = serialization.WriteUint32(fWriter, uint32(len(data)))
		if err != nil {
			fmt.Errorf("write block data height:%d len:%d error:%s\n", i, uint32(len(data)), err)
			return
		}
		_, err = fWriter.Write(data)
		if err != nil {
			fmt.Errorf("write block data height:%d error:%s\n", i, err)
			return
		}

		bar.Incr()
	}
	uiprogress.Stop()

	err = fWriter.Flush()
	if err != nil {
		fmt.Errorf("Export flush file error:%s\n", err)
		return
	}
	fmt.Printf("Export blocks successfully.\n")
	fmt.Printf("Total blocks:%d\n", blockHeight)
	fmt.Printf("Export file:%s\n", DEFAULT_BLOCK_FILE_NAME)
}
