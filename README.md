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

    root@DS2-V2-36:/home/ubuntu# ./txreplay tximport --importtxsfile txs-20180703 --ip polaris2.ont.io -h
	NAME:
	   txreplay tximport - Import txs from a file
	
	USAGE:
	   txreplay tximport [command options] [arguments...]
	
	OPTIONS:
	   --ip value             node's ip address (default: "localhost")
	   --rpcport value        Json rpc server listening port (default: 20336)
	   --importtxsfile value  Path of import txs file (default: "./txs.dat")
	   --routinenum value     concurrent routine number (default: 1)
	   --constanttimer value  constant timer delay (ms) (default: 1)
	
	root@DS2-V2-36:/home/ubuntu# ./txreplay tximport --importtxsfile txs-20180703 --ip polaris2.ont.io --routinenum 4
    Tue Jul  3 05:14:22 UTC 2018 Start import Txs...
	Tue Jul  3 05:14:22 UTC 2018 Sent tx count 0, errNum 0
	Tue Jul  3 05:14:23 UTC 2018 Sent tx count 7, errNum 0
	Tue Jul  3 05:14:23 UTC 2018 Sent tx count 11, errNum 0
	Tue Jul  3 05:14:23 UTC 2018 Sent tx count 12, errNum 0
	Tue Jul  3 05:14:23 UTC 2018 Sent tx count 14, errNum 0
    ...
    Tue Jul  3 05:24:21 UTC 2018 Sent tx count 24540, errNum 0
	Tue Jul  3 05:24:22 UTC 2018 Import Txs complete, total txs 24540 sent txs 24540 errNum 0
	root@DS2-V2-36:/home/ubuntu#



