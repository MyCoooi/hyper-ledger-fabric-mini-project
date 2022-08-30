# hyper-ledger-fabric-mini-project

-READY
1. Network: fabric-samples/test-network
(*Github address: https://github.com/hyperledger/fabric-samples/tree/main/test-network)

cd ~
git clone https://github.com/hyperledger/fabric-samples/tree/main/test-network

2. node js install

3. golang install





-START

cd ~/test-network
./network.sh up createChannel -ca -c mychannel
./network.sh deployCC -ccn review_event -ccl go -ccv 1.0 -ccp ~/hyper-ledger-fabric-mini-project/chaincode
cd ~/hyper-ledger-fabric-mini-project/server
node server.js





-End

cd ~/test-network
./network.sh down
