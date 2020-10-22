// Copyright 2020 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package bdls_engine

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/binary"
	"errors"
	"math/big"
	"sort"

	"github.com/Sperax/SperaxChain/common"
	"github.com/Sperax/SperaxChain/common/hexutil"
	"github.com/Sperax/SperaxChain/core/state"
	"github.com/Sperax/SperaxChain/core/types"
	"github.com/Sperax/SperaxChain/crypto"
	"github.com/Sperax/SperaxChain/log"
	"github.com/Sperax/SperaxChain/params"
	"github.com/Sperax/SperaxChain/rlp"
	"github.com/Sperax/bdls"
	"golang.org/x/crypto/sha3"
)

var (
	// Base Quorum is the quorum to make sure blockchain can generate new blocks
	// while no other validators are running.
	BaseQuorum = []common.Address{
		common.HexToAddress("f2580391fe8a83366ed550de4e45af1714d74b8d"),
		common.HexToAddress("066aaff9e575302365b7862dcebd4a5a65f75f5f"),
		common.HexToAddress("3f80e8718d8e17a1768b467f193a6fbeaa6236e3"),
		common.HexToAddress("29d3fbe3e7983a41d0e6d984c480ceedb3c251fd"),
	}

	BaseQuorumR = common.BytesToHash(hexutil.MustDecode("0x053706572617844656661756c744e6f646"))
)

var (
	CommonCoin = []byte("Sperax")
	// block 0 common random number
	W0 = crypto.Keccak256Hash(hexutil.MustDecode("0x03243F6A8885A308D313198A2E037073"))
	// potential propser expectation
	E1 = big.NewInt(5)
	// BFT committee expectationA
	E2 = big.NewInt(50)
	// unit of staking SPA
	StakingUnit = new(big.Int).Mul(big.NewInt(100000), big.NewInt(params.Ether))
	// transfering tokens to this address will be specially treated
	StakingAddress = common.HexToAddress("0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee")

	// max unsigned 256-bit integer
	MaxUint256 = big.NewFloat(0).SetInt(big.NewInt(0).SetBytes([]byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}))
)

var (
	ErrStakingRequest       = errors.New("already staked")
	ErrStakingMinimumTokens = errors.New("staking has less than minimum tokens")
	ErrStakingInvalidPeriod = errors.New("invalid staking period")
	ErrRedeemRequest        = errors.New("not staked")
)

// types of staking related operation
type StakingOp byte

// Staking Operations
const (
	Staking = StakingOp(0x00)
	Redeem  = StakingOp(0xFF)
)

// StakingRequest will be sent along in transaction.payload
type StakingRequest struct {
	// Staking or Redeem operation
	StakingOp StakingOp

	// The begining height to participate in consensus
	StakingFrom uint64

	// The ending  height to participate in consensus
	StakingTo uint64

	// The staker's hash at the height - StakingFrom
	StakingHash common.Hash
}

// Staker & StakingObject are the structures stored in
// StakingAddress's Account.Code for staking related information
// A single Staker
type Staker struct {
	// the Staker's address
	Address common.Address
	// the 1st block expected to participant in validator and proposer
	StakingFrom uint64
	// the last block to participant in validator and proposer, the tokens will be refunded
	// to participants' addresses after this block has mined
	StakingTo uint64
	// StakingHash is the last hash in hashchain,  random nubmers(R) in futureBlock
	// will be hashed for (futureBlock - stakingFrom) times to match with StakingHash.
	StakingHash common.Hash
	// records the number of tokens staked
	StakedValue *big.Int
}

// The object to be stored in StakingAddress's Account.Code
type StakingObject struct {
	Stakers []Staker // staker's, expired stakers will automatically be removed
}

// GetStakingObject returns the stakingObject at some state
func (e *BDLSEngine) GetStakingObject(state *state.StateDB) (*StakingObject, error) {
	var stakingObject StakingObject
	// retrieve staking object from account.Code
	code := state.GetCode(StakingAddress)
	if code != nil {
		err := rlp.DecodeBytes(code, &stakingObject)
		if err != nil {
			return nil, err
		}
	}
	return &stakingObject, nil
}

// GetW calculates random number W based on block information
// W0 = H(U0)
// Wj = H(Pj-1,Wj-1) for 0<j<=r,
func (e *BDLSEngine) deriveW(header *types.Header) common.Hash {
	if header.Number.Uint64() == 0 {
		return W0
	}

	hasher := sha3.NewLegacyKeccak256()

	// derive Wj from Pj-1 & Wj-1
	hasher.Write(header.Coinbase.Bytes())
	hasher.Write(header.W.Bytes())
	return common.BytesToHash(hasher.Sum(nil))
}

// H(r;0;Ri,r,0;Wr) > max{0;1 i-aip}
func (e *BDLSEngine) IsProposer(header *types.Header, stakingObject *StakingObject) bool {
	// empty block is valid proposer
	if header.Coinbase == StakingAddress {
		// make sure transactions are emtpy for staking miner
		if header.TxHash != types.EmptyRootHash {
			log.Debug("verifyProposerField for empty transactions", "header.TxHash", header.TxHash)
			return false
		} else if header.ReceiptHash != types.EmptyRootHash {
			log.Debug("verifyProposerField for empty receipts", "header.ReceiptHash", header.ReceiptHash)
			return false
		}
		return true
	}

	// addresses in base quorum are permanent proposers
	for k := range BaseQuorum {
		if header.Coinbase == BaseQuorum[k] {
			return true
		}
	}

	// non-empty blocks
	numStaked := big.NewFloat(0)
	totalStaked := big.NewFloat(0) // effective stakings

	// lookup the staker's information
	for k := range stakingObject.Stakers {
		staker := stakingObject.Stakers[k]
		// count effective stakings
		if header.Number.Uint64() > staker.StakingFrom || header.Number.Uint64() <= staker.StakingTo {
			totalStaked.Add(totalStaked, big.NewFloat(0).SetInt(staker.StakedValue))
		}

		if staker.Address == header.Coinbase {
			if header.Number.Uint64() <= staker.StakingFrom || header.Number.Uint64() > staker.StakingTo {
				log.Debug("invalid staking period")
				return false
			} else if common.BytesToHash(e.hashChain(staker.StakingHash.Bytes(), header.Number.Uint64()-staker.StakingFrom)) != header.R {
				log.Debug("hashchain verification failed for header.R")
				return false
			} else {
				numStaked = big.NewFloat(0).SetInt(staker.StakedValue)
			}
		}
	}

	// if there's staking
	if totalStaked.Sign() == 1 {
		// compute p
		p := big.NewFloat(0).SetInt(E1)
		p.Mul(p, big.NewFloat(0).SetInt(StakingUnit))
		p.Quo(p, totalStaked)

		// max{0, 1 - ai*p}
		max := p.Sub(big.NewFloat(1), p.Mul(numStaked, p))
		if max.Cmp(big.NewFloat(0)) != 1 {
			max = big.NewFloat(0)
		}

		// compute proposer hash
		proposerHash := e.proposerHash(header)

		// calculate H/MaxUint256
		h := big.NewFloat(0).SetInt(big.NewInt(0).SetBytes(proposerHash.Bytes()))
		h.Quo(h, MaxUint256)

		// prob compare
		if h.Cmp(max) == 1 {
			return true
		}
	}

	return false
}

// ValidatorVotes counts the number of votes for a validator
func (e *BDLSEngine) ValidatorVotes(header *types.Header, staker *Staker, totalStaked *big.Int) uint64 {
	numStaked := staker.StakedValue
	validatorR := staker.StakingHash

	// compute p'
	// p' = E2* numStaked /totalStaked
	p := big.NewFloat(0).SetInt(E2)
	p.Mul(p, big.NewFloat(0).SetInt(StakingUnit))
	p.Quo(p, big.NewFloat(0).SetInt(totalStaked))

	maxVotes := numStaked.Uint64() / StakingUnit.Uint64()

	// compute validator's hash
	validatorHash := e.validatorHash(header.Coinbase, header.Number.Uint64(), validatorR, header.W)

	// calculate H/MaxUint256
	h := big.NewFloat(0).SetInt(big.NewInt(0).SetBytes(validatorHash.Bytes()))
	h.Quo(h, MaxUint256)

	// find the minium possible votes
	var votes uint64
	binominal := big.NewInt(0)
	for i := uint64(0); i <= maxVotes; i++ {
		// computes binomial
		sum := big.NewFloat(0)
		for j := uint64(0); j <= i; j++ {
			coefficient := big.NewFloat(float64(binominal.Binomial(int64(maxVotes), int64(j)).Uint64()))
			a := Pow(p, j)
			b := Pow(big.NewFloat(0).Sub(big.NewFloat(1), p), maxVotes-j)
			r := big.NewFloat(0).Mul(a, b)
			r.Mul(r, coefficient)
			sum.Add(sum, r)
		}

		// effective vote
		if sum.Cmp(h) == 1 {
			votes = i
		}
	}

	return votes
}

type orderedValidator struct {
	identity bdls.Identity
	hash     common.Hash
}

type SortableValidators []orderedValidator

func (s SortableValidators) Len() int { return len(s) }
func (s SortableValidators) Less(i, j int) bool {
	return bytes.Compare(s[i].hash.Bytes(), s[j].hash.Bytes()) == -1
}
func (s SortableValidators) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

// validatorHash computes a hash for validator's random number
func (ov SortableValidators) Hash(height uint64, R common.Hash, W common.Hash) common.Hash {
	hasher := sha3.New256()
	binary.Write(hasher, binary.LittleEndian, height)
	binary.Write(hasher, binary.LittleEndian, 1)
	hasher.Write(R.Bytes())
	hasher.Write(CommonCoin)
	hasher.Write(W.Bytes())

	return common.BytesToHash(hasher.Sum(nil))
}

// CreateValidators creates an ordered list for all qualified validators with weights
func (e *BDLSEngine) CreateValidators(header *types.Header, stakingObject *StakingObject) []bdls.Identity {
	var orderedValidators []orderedValidator

	// count effective stakings
	var totalStaked *big.Int
	for k := range stakingObject.Stakers {
		staker := stakingObject.Stakers[k]
		// count effective stakings
		if header.Number.Uint64() > staker.StakingFrom || header.Number.Uint64() <= staker.StakingTo {
			totalStaked.Add(totalStaked, staker.StakedValue)
		}
	}

	// setup validators
	for k := range stakingObject.Stakers {
		staker := stakingObject.Stakers[k]
		if header.Number.Uint64() <= staker.StakingFrom || header.Number.Uint64() > staker.StakingTo {
			continue
		} else {
			n := e.ValidatorVotes(header, &staker, totalStaked)
			for i := uint64(0); i < n; i++ { // a validator has N slots to be a leader
				var validator orderedValidator
				copy(validator.identity[:], staker.Address.Bytes())
				validator.hash = e.validatorSortingHash(staker.Address, staker.StakingHash, header.W, i)
				orderedValidators = append(orderedValidators, validator)
			}
		}
	}

	// sort by the validators based on the sorting hash
	sort.Stable(SortableValidators(orderedValidators))
	var sortedValidators []bdls.Identity
	for i := 0; i < len(orderedValidators); i++ {
		sortedValidators = append(sortedValidators, orderedValidators[i].identity)
	}

	// always append based quorum to then end of the validators
	for k := range BaseQuorum {
		var id bdls.Identity
		copy(id[:], BaseQuorum[k][:])
		sortedValidators = append(sortedValidators, id)
	}

	return sortedValidators
}

// Pow calculates a^e
func Pow(a *big.Float, e uint64) *big.Float {
	result := big.NewFloat(0.0).Copy(a)
	for i := uint64(0); i < e-1; i++ {
		result = big.NewFloat(0.0).Mul(result, a)
	}
	return result
}

// proposerHash computes a hash for proposer's random number
func (e *BDLSEngine) proposerHash(header *types.Header) common.Hash {
	hasher := sha3.New256()
	hasher.Write(header.Coinbase.Bytes())
	binary.Write(hasher, binary.LittleEndian, header.Number.Uint64())
	binary.Write(hasher, binary.LittleEndian, 0)
	hasher.Write(header.R.Bytes())
	hasher.Write(CommonCoin)
	hasher.Write(header.W.Bytes())

	return common.BytesToHash(hasher.Sum(nil))
}

// validatorHash computes a hash for validator's random number
func (e *BDLSEngine) validatorHash(coinbase common.Address, height uint64, R common.Hash, W common.Hash) common.Hash {
	hasher := sha3.New256()
	hasher.Write(coinbase.Bytes())
	binary.Write(hasher, binary.LittleEndian, height)
	binary.Write(hasher, binary.LittleEndian, 1)
	hasher.Write(R.Bytes())
	hasher.Write(CommonCoin)
	hasher.Write(W.Bytes())

	return common.BytesToHash(hasher.Sum(nil))
}

// validatorSortHash computes a hash for validator's sorting hashing
func (e *BDLSEngine) validatorSortingHash(address common.Address, R common.Hash, W common.Hash, votes uint64) common.Hash {
	hasher := sha3.New256()
	hasher.Write(address.Bytes())
	binary.Write(hasher, binary.LittleEndian, votes)
	hasher.Write(R.Bytes())
	hasher.Write(CommonCoin)
	hasher.Write(W.Bytes())

	return common.BytesToHash(hasher.Sum(nil))
}

// deriveStakingSeed deterministically derives the pseudo-random number with height and private key
// seed := H(H(privatekey,stakingFrom) *G)
func (e *BDLSEngine) deriveStakingSeed(priv *ecdsa.PrivateKey, stakingFrom uint64) []byte {
	// H(privatekey + stakingFrom)
	hasher := sha3.New256()
	hasher.Write(priv.D.Bytes())
	binary.Write(hasher, binary.LittleEndian, stakingFrom)

	// H(privatekey + lastHeight) *G
	x, y := crypto.S256().ScalarBaseMult(hasher.Sum(nil))

	// H(H(privatekey + lastHeight) *G)
	hasher = sha3.New256()
	hasher.Write(x.Bytes())
	hasher.Write(y.Bytes())
	return hasher.Sum(nil)
}

// compute hash recursively for n(n>=0) times
func (e *BDLSEngine) hashChain(hash []byte, n uint64) []byte {
	lastHash := hash
	hasher := sha3.New256()
	for i := uint64(0); i < n; i++ {
		hasher.Write(lastHash)
		lastHash = hasher.Sum(nil)
	}
	return lastHash
}
