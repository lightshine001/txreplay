# txreplay

txreplay to export transactions from ontology db to a file and import transactions from a file to ontology.

    [aaron@localhost txreplay]$ ./txreplay -h
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

Export transactions to a file


    [aaron@localhost txreplay]$ ./txreplay txexport -h
	NAME:
	   txreplay txexport - Export txs in DB to a file
	
	USAGE:
	   txreplay txexport [command options] [arguments...]
	
	OPTIONS:
	   --ip value       node's ip address (default: "localhost")
	   --rpcport value  Json rpc server listening port (default: 20336)
	   --file value     Path of export file (default: "./txs.dat")
	   --height value   Using to specifies the beginning of the block to be exported. (default: 0)
	   
	[aaron@localhost txreplay]$ ./txreplay txexport --ip 139.219.128.213 --rpcport 20336 --file txs-20180626 --height 20
	Start export...
	Remaining Block 0 [====================================================================] 100%   30s
	Export txs successfully.
	Total txs:76642 from block 20 to block 550
	Export file:txs-20180626


Import transactions from a file to Ontology Chain

    [aaron@localhost txreplay]$ ./txreplay tximport -h
	NAME:
	   txreplay tximport - Import txs from a file
	
	USAGE:
	   txreplay tximport [command options] [arguments...]
	
	OPTIONS:
	   --ip value             node's ip address (default: "localhost")
	   --rpcport value        Json rpc server listening port (default: 20336)
	   --importtxsfile value  Path of import txs file (default: "./txs.dat")
	   
	[aaron@localhost txreplay]$ 
	[aaron@localhost txreplay]$ ./txreplay tximport --ip 139.219.128.213 --rpcport 20336 --importtxsfile txs-20180626
	Tue Jun 26 09:09:00 UTC 2018 Start import Txs...
	Tue Jun 26 09:09:07 UTC 2018 Sent tx count 792, errNum 0
	...
	Tue Jun 26 09:20:04 UTC 2018 Sent tx count 75514, errNum 0
	Tue Jun 26 09:20:11 UTC 2018 Sent tx count 76281, errNum 0
	Tue Jun 26 09:20:14 UTC 2018 Sent tx count 76642, errNum 0
	Tue Jun 26 09:20:14 UTC 2018 Import Txs complete, total txs 76642 sent txs 76642 errNum 0



