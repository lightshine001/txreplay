# txreplay

txreplay to export transactions from ontology db to a file and import transactions from a file to ontology.

    root@DS2-V2-36:/home/ubuntu# ./txreplay -h
	NAME:
	   txreplay - Ontology tx replay
	
	USAGE:
	   txreplay [global options] command [command options] [arguments...]
	
	COMMANDS:
	     txexport  Export txs in DB to a file
	     tximport  Import txs from a file
	     help, h   Shows a list of commands or help for one command
	
	GLOBAL OPTIONS:
	   --help, -h     show help
	   --version, -v  print the version
	
	COPYRIGHT:
	   Copyright in 2018 The Ontology Authors
	root@DS2-V2-36:/home/ubuntu#


Export transactions to a file


	root@DS2-V2-36:/home/ubuntu# ./txreplay txexport -h
	NAME:
	   txreplay txexport - Export txs in DB to a file
	
	USAGE:
	   txreplay txexport [command options] [arguments...]
	
	OPTIONS:
	   --ip value       node's ip address (default: "localhost")
	   --rpcport value  Json rpc server listening port (default: 20336)
	   --file value     Path of export file (default: "./txs.dat")
	   --height value   Using to specifies the beginning of the block to be exported. (default: 0)
	
	root@DS2-V2-36:/home/ubuntu# ./txreplay txexport --ip polaris2.ont.io --file txs-20180703 --rpcport 20336 --height 20
	Start export...
	Remaining Block 0 [====================================================================] 100% 3m58ssss
	Export txs successfully.
	Total txs:24540 from block 20 to block 2421
	Export file:txs-20180703
	root@DS2-V2-36:/home/ubuntu#



Import transactions from a file to Ontology Chain

    1. Copy the consensus wallets on the target chain net to local
	-rw-rw-r-- 1 ubuntu ubuntu      511 Jul  5 05:48 wallet1.dat
	-rw-rw-r-- 1 ubuntu ubuntu      511 Jul  5 05:48 wallet2.dat
	-rw-rw-r-- 1 ubuntu ubuntu      511 Jul  5 05:48 wallet3.dat
	-rw-rw-r-- 1 ubuntu ubuntu      511 Jul  5 05:48 wallet4.dat
	-rw-rw-r-- 1 ubuntu ubuntu      511 Jul  5 05:48 wallet5.dat
	-rw-rw-r-- 1 ubuntu ubuntu      511 Jul  5 05:48 wallet6.dat
	-rw-rw-r-- 1 ubuntu ubuntu      511 Jul  5 05:48 wallet7.dat

    2. Create wallet config. Please refer to the sample(wallets.json)
        {
                "Path": "wallet1.dat",
            "Password": "1"
        },
    3. Copy the Chain db on the target chain net to local
    4. Rebuild blocks with the exported txs and append those blocks to the above Chain db, meanwhile, export the blocks on local Chain after finish rebuild.
	    Sample:
	    root@DS2-V2-35:/home/ubuntu/test# ./txreplay tximport -h
		NAME:
		   txreplay tximport - Import txs from a file
		
		USAGE:
		   txreplay tximport [command options] [arguments...]
		
		OPTIONS:
		   --importtxsfile value  Path of import txs file (default: "./txs.dat")
		   --networkid value      Using to specify the network ID. Different networkids cannot connect to the blockchain network. 1=ontology main net, 2=polaris test net, 3=testmode, and other for custom network (default: 1)
		   --constanttimer value  constant timer delay (ms) (default: 1)
	     root@DS2-V2-35:/home/ubuntu/test# ./txreplay tximport --networkid 2 --importtxsfile txs-20180705
            ...
			Thu Jul  5 06:49:07 UTC 2018 packed tx count 38237 errNum 10,  current block height 4215  block hash 61a69de4c303c2175625bf4b5999b42cd60aae4db0947f03778b1993696b1e4a
			Thu Jul  5 06:49:07 UTC 2018 Import Txs complete, total txs 38247 packed txs 38237 errNum 10
			Start export block.
			Block(4215/4215) [====================================================================] 100%    8s
			Export blocks successfully.
			Total blocks:4215
			Export file:block.dat
			root@DS2-V2-35:/home/ubuntu/test#

     5. Clean the target chain db and copy the generated block.dat to use block import function to start chain net.  
        root@DS2-V2-35:/opt/gopath/test# ./ontology  --import --importfile block.dat
     




    



