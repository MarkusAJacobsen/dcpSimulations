package main

import (
	"fmt"
	"github.com/DCP-DCT/DCP"
	"github.com/google/uuid"
	"math/rand"
	"os"
	"time"
)

func main() {
	//benchmarkEncryption(50)

	temp := os.Stdout
	os.Stdout = nil
	runSimulation(10)
	os.Stdout = temp
}

func createNodes(numberOfNodes int, config *DCP.CtNodeConfig) []*DCP.CtNode {
	var nodes []*DCP.CtNode
	rand.Seed(time.Now().UnixNano())

	for i := 0; i < numberOfNodes; i++ {
		node := &DCP.CtNode{
			Id: uuid.New(),
			Co: &DCP.CalculationObjectPaillier{
				TransactionId: uuid.New(),
				Counter:       0,
			},
			Ids:          GenerateIdTable(rand.Intn(25)),
			HandledCoIds: make(map[uuid.UUID]struct{}),
			TransportLayer: &DCP.ChannelTransport{
				DataCh:          make(chan *[]byte),
				StopCh:          make(chan struct{}),
				ReachableNodes:  make(map[chan *[]byte]chan struct{}),
				SuppressLogging: config.SuppressLogging,
			},
			Config: config,
		}
		e := node.Co.KeyGen()
		if e != nil {
			fmt.Println(e.Error())
			break
		}

		nodes = append(nodes, node)
	}

	return nodes
}

func benchmarkEncryption(numberOfNodes int) {
	config := &DCP.CtNodeConfig{
		NodeVisitDecryptThreshold: 5,
	}

	nodes := createNodes(numberOfNodes, config)

	for _, node := range nodes {
		node.Listen()
	}

	initialNode := nodes[0]
	EstablishNodeRelationShipAllInRange(nodes)

	e := DCP.InitRoutine(DCP.PrepareIdLenCalculation, initialNode)
	if e != nil {
		fmt.Println(e)
	}

	initialNode.Broadcast(nil)

	time.Sleep(10 * time.Second)
	msg := initialNode.Co.Decrypt(initialNode.Co.Cipher)

	fmt.Printf("Initial Node Counter %d, Node Cipher %s\n", initialNode.Co.Counter, msg.String())
}

func runSimulation(numberOfNodes int) {
	config := &DCP.CtNodeConfig{
		NodeVisitDecryptThreshold: 2,
		SuppressLogging:           true,
	}

	nodes := createNodes(numberOfNodes, config)
	EstablishNodeRelationShipAllInRange(nodes)

	for _, node := range nodes {
		node.Listen()
	}

	done := make(chan struct{})
	go LaunchMonitor(nodes, done)

	stop := make(chan struct{})
	for _, node := range nodes {
		go RandomCalculationProcessInitiator(node, stop)
	}

	for {
		select {
		case <-done:
			close(stop)
			return
		}
	}
}
