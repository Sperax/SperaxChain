package node

import (
	"context"
	"encoding/hex"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/Sperax/SperaxChain/consensus/bdls_engine"
	"github.com/Sperax/SperaxChain/core"
	"github.com/Sperax/SperaxChain/core/rawdb"
	"github.com/Sperax/SperaxChain/core/types"
	"github.com/Sperax/SperaxChain/core/vm"
	"github.com/Sperax/SperaxChain/p2p"
	"github.com/Sperax/SperaxChain/worker"
	"github.com/Sperax/bdls"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rlp"
	proto "github.com/golang/protobuf/proto"
	lru "github.com/hashicorp/golang-lru"
	libp2p_pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/minio/blake2b-simd"

	"github.com/Sperax/bdls/timer"
)

const (
	freezerDir = "/freezer"
	chainDBDir = "/chaindb"
	namespace  = "sperax/db/chaindata/"
)

const (
	p2pGenericTopic = "sperax/transactions/1.0.0"
)

// Node represents a Sperax node on it's network
type Node struct {
	// the p2p host for messaging
	host *p2p.Host

	// the consensus p2p interface for broadcasting
	p2pEntry *p2p.BDLSEntry

	// consensus related
	consensusConfig *bdls.Config            // the configuration for BDLS consensus algorithm
	consensus       *bdls.Consensus         // the current working consensus object
	consensusEngine *bdls_engine.BDLSEngine // the current working consensus engine(for verification)
	consensusLock   sync.Mutex              // consensus related lock
	// consensus in-progress blocks
	unconfirmedBlocks *lru.Cache
	proposedBlock     *types.Block

	// generic topic to exchange transactions, blocks
	speraxTopic *libp2p_pubsub.Topic

	// worker to assemble new block to propose
	worker *worker.Worker

	// transactions pool for local & remote transactions
	txPool *core.TxPool

	// the blockchain
	blockchain *core.BlockChain

	// closing signal
	die     chan struct{} // closing signal
	dieOnce sync.Once
}

// New creates a new node.
func New(host *p2p.Host, consensusConfig *bdls.Config, config *Config) (*Node, error) {
	node := new(Node)
	node.host = host
	node.die = make(chan struct{})
	node.consensusConfig = consensusConfig
	topic, err := host.GetOrJoin(p2pGenericTopic)
	if err != nil {
		return nil, err
	}
	node.speraxTopic = topic
	cache, err := lru.New(128) // TODO: config
	if err != nil {
		panic(err)
	}

	node.unconfirmedBlocks = cache
	// consensus network entry
	entry, err := p2p.NewBDLSPeerAdapter(node.host)
	if err != nil {
		panic(err)
	}
	node.p2pEntry = entry

	// init chaindb
	chainDb, err := rawdb.NewLevelDBDatabaseWithFreezer(config.DatabaseDir+chainDBDir, config.DatabaseCache, config.DatabaseHandles, config.DatabaseDir+freezerDir, namespace)
	if err != nil {
		log.Println("new leveldb:", chainDb, err)
		return nil, err
	}
	chainConfig, genesisHash, genesisErr := core.SetupGenesisBlock(chainDb, config.Genesis)
	if _, ok := genesisErr.(*params.ConfigCompatError); genesisErr != nil && !ok {
		return nil, genesisErr
	}
	log.Println("Initialised chain configuration", "config", chainConfig, genesisHash)

	// init basic consensus to init blockchain
	basicConsensus, err := bdls.NewConsensus(consensusConfig)
	if err != nil {
		panic(err)
	}
	basicConsensus.SetLatency(200 * time.Millisecond)
	engine := bdls_engine.NewBDLSEngine()
	engine.SetConsensus(basicConsensus)
	node.consensusEngine = engine
	node.consensus = basicConsensus

	// cache config
	cacheConfig := &core.CacheConfig{
		TrieCleanLimit:      config.TrieCleanCache,
		TrieCleanNoPrefetch: config.NoPrefetch,
		TrieDirtyLimit:      config.TrieDirtyCache,
		TrieDirtyDisabled:   config.NoPruning,
		TrieTimeLimit:       config.TrieTimeout,
		SnapshotLimit:       config.SnapshotCache,
	}

	// vm config
	vmConfig := vm.Config{
		EnablePreimageRecording: config.EnablePreimageRecording,
		EWASMInterpreter:        config.EWASMInterpreter,
		EVMInterpreter:          config.EVMInterpreter,
	}

	// init blockchain
	node.blockchain, err = core.NewBlockChain(chainDb, cacheConfig, chainConfig, engine, vmConfig, nil, &config.TxLookupLimit)
	if err != nil {
		log.Println("new blockchain:", err)
		return nil, err
	}

	// init txpool
	txPoolConfig := core.DefaultTxPoolConfig
	node.txPool = core.NewTxPool(txPoolConfig, chainConfig, node.blockchain)

	// init worker
	node.worker = worker.New(config.Genesis.Config, node.blockchain, engine)

	// kick off consensus updater
	node.consensusUpdater()
	// start core messaging loop
	go node.messenger()
	// start consensus messaging loop
	go node.consensusMessenger()
	return node, nil
}

// Close this node
func (node *Node) Close() {
	node.dieOnce.Do(func() {
		close(node.die)
	})
}

// messenger is a goroutine to receive all messages required for transactions & blocks
func (node *Node) messenger() {
	sub, err := node.speraxTopic.Subscribe()
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	for {
		msg, err := sub.Next(ctx)
		if err != nil {
			log.Println(err)
			continue
		}

		// Unmarshal message
		message := new(SperaxMessage)
		proto.Unmarshal(msg.Data, message)

		// handle tx & blocks
		switch message.Type {
		case MessageType_Transaction:
			tx := new(types.Transaction)
			err := rlp.DecodeBytes(message.Message, tx)
			if err != nil {
				log.Println(err)
				continue
			}
			node.AddRemoteTransaction(tx)
		case MessageType_Block:
			block := new(types.Block)
			err := rlp.DecodeBytes(message.Message, block)
			if err != nil {
				log.Println(err)
				continue
			}

			if block.Decision() == nil { // unconfirmed block, awaiting consensus
				// validators  may broadcast unconfirmed blocks ahead of consensus.roundchange
				node.unconfirmedBlocks.Add(block.Hash(), block)
			} else {
				// confirmed block, store to blockchain
				height := uint64(node.blockchain.CurrentHeader().Number.Int64())
				node.AddBlock(block)
				newHeight := uint64(node.blockchain.CurrentHeader().Number.Int64())

				// as validator we should propose new block at new height
				if newHeight > height {
					newBlock, err := node.proposeNewBlock()
					if err != nil {
						panic(err)
					}
					node.proposedBlock = newBlock
					// start consensus
					log.Println("current height:", newHeight)
					node.beginConsensus(newBlock, newHeight+1)
				}
			}
		}
	}
}

// consensusMessenger is a goroutine to receive all messages required for BDLS consensus
func (node *Node) consensusMessenger() {
	newBlock, err := node.proposeNewBlock()
	if err != nil {
		panic(err)
	}
	node.proposedBlock = newBlock

	log.Println("current height:", uint64(node.blockchain.CurrentHeader().Number.Int64()))
	node.beginConsensus(newBlock, uint64(node.blockchain.CurrentHeader().Number.Int64())+1)

	// subscribe & handle messages
	sub, err := node.p2pEntry.Topic().Subscribe()
	ctx := context.Background()
	for {
		msg, err := sub.Next(ctx)
		if err != nil {
			log.Println(err)
			continue
		}

		node.consensusLock.Lock()
		currentHeight, _, _ := node.consensus.CurrentState()
		// handle consensus messages
		node.consensus.ReceiveMessage(msg.Data, time.Now())
		newHeight, newRound, newState := node.consensus.CurrentState()
		node.consensusLock.Unlock()

		// new height, broadcast confirmed block
		if newHeight > currentHeight {
			h := blake2b.Sum256(newState)
			log.Printf("<decide> at height:%v round:%v hash:%v", newHeight, newRound, hex.EncodeToString(h[:]))

			blkHash := common.BytesToHash(newState)
			value, ok := node.unconfirmedBlocks.Get(blkHash)
			if !ok {
				panic("no block")
			}

			// seal the block with proof
			header := value.(*types.Block).Header()
			header.Decision = msg.Data // store the the proof in block header
			finalized := value.(*types.Block).WithSeal(header)

			// broadcast this block
			node.broadcastBlock(finalized)
		}
	}
}

// broadcast a given block
func (node *Node) broadcastBlock(block *types.Block) error {
	message := new(SperaxMessage)
	message.Type = MessageType_Block
	bts, err := rlp.EncodeToBytes(block)
	if err != nil {
		return err
	}
	message.Message = bts

	// marshal to SperaxMessage
	bts, err = proto.Marshal(message)
	if err != nil {
		return err
	}

	// wire
	return node.speraxTopic.Publish(context.Background(), bts)
}

//  begin Consensus on new height
func (node *Node) beginConsensus(block *types.Block, height uint64) error {
	node.consensusLock.Lock()
	defer node.consensusLock.Unlock()

	// calculate block hash(with Decision field setting to nil)
	blockHash := block.Hash()

	// initiated new consensus object for new height with new config
	newConfig := new(bdls.Config)
	*newConfig = *node.consensusConfig
	newConfig.CurrentHeight = height
	newConfig.StateValidate = func(s bdls.State) bool {
		h := common.BytesToHash(s)
		// check if it's the local proposed block
		if node.proposedBlock.Hash() == h {
			return true
		}
		// check if it's the remote proposed block
		if _, ok := node.unconfirmedBlocks.Get(h); ok {
			log.Println("state validate true")
			return true
		}
		log.Println("state validate false")
		return false
	}

	// we register a consensus message watcher here, to send data along with consensus
	newConfig.MessageCallback = func(m *bdls.Message, sp *bdls.SignedProto) {
		if m.Type == bdls.MessageType_RoundChange {
			message := new(SperaxMessage)
			message.Type = MessageType_Block
			bts, err := rlp.EncodeToBytes(block)
			if err != nil {
				log.Println(err)
			}
			message.Message = bts

			// marshal to SperaxMessage
			bts, err = proto.Marshal(message)
			if err != nil {
				log.Println(err)
			}

			// wire
			err = node.speraxTopic.Publish(context.Background(), bts)
			if err != nil {
				log.Println(err)
			}
		}
	}

	// replace current working consensus object with newer
	node.consensus, _ = bdls.NewConsensus(newConfig)
	node.consensus.Join(node.p2pEntry)

	// also update the engine for verification
	node.consensusEngine.SetConsensus(node.consensus)

	// purge all unconfirmed blocks
	node.unconfirmedBlocks.Purge()

	// propose the block hash to consensus
	node.consensus.Propose(blockHash.Bytes())
	return nil
}

// consensusUpdater is a self-sustaining function to call consensus.Update periodically
// with the help of bdls.timer
func (node *Node) consensusUpdater() {
	node.consensusLock.Lock()
	if node.consensus != nil {
		node.consensus.Update(time.Now())
	}
	node.consensusLock.Unlock()
	timer.SystemTimedSched.Put(node.consensusUpdater, time.Now().Add(20*time.Millisecond))
}

// proposeNewBlock collects transactions from txpool and seal a new block to propose to
// consensus algorithm
func (node *Node) proposeNewBlock() (*types.Block, error) {
	// update current header & reset statsdb
	node.worker.UpdateCurrent()

	// fetch transactions from txpoll
	pending, err := node.txPool.Pending()
	if err != nil {
		return nil, errors.New("Failed to fetch pending transactions")
	}

	coinbase := common.Address{}
	if err := node.worker.CommitTransactions(pending, coinbase); err != nil {
		return nil, err
	}

	log.Println("proposed")
	return node.worker.FinalizeNewBlock()
}

// Add a remote transactions
func (node *Node) AddRemoteTransaction(tx *types.Transaction) error {
	err := node.txPool.AddRemote(tx)
	if err != nil {
		return err
	}
	pendingCount, queueCount := node.txPool.Stats()
	log.Println("addtx:", pendingCount, queueCount)
	return nil
}

// AddBlock
func (node *Node) AddBlock(block *types.Block) error {
	_, err := node.blockchain.InsertChain([]*types.Block{block})
	if err != nil {
		return err
	}

	log.Println("added new block:", block.NumberU64())
	return nil
}
