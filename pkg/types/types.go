// Copyright IBM Corp. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0
//

package types

import (
	"crypto/sha256"
	"encoding/asn1"
	"encoding/hex"
	"fmt"
	"sync"

	"github.com/SmartBFT-Go/consensus/smartbftprotos"
)

type Proposal struct {
	Payload              []byte
	Header               []byte
	Metadata             []byte
	VerificationSequence int64 // int64 for asn1 marshaling
}

type Signature struct {
	ID    uint64
	Value []byte
	Msg   []byte
}

type Decision struct {
	Proposal   Proposal
	Signatures []Signature
}

type ViewAndSeq struct {
	View uint64
	Seq  uint64
}

type RequestInfo struct {
	ClientID string
	ID       string
}

func (r *RequestInfo) String() string {
	return r.ClientID + ":" + r.ID
}

func (p Proposal) Digest() string {
	rawBytes, err := asn1.Marshal(Proposal{
		VerificationSequence: p.VerificationSequence,
		Metadata:             p.Metadata,
		Payload:              p.Payload,
		Header:               p.Header,
	})
	if err != nil {
		panic(fmt.Sprintf("failed marshaling proposal: %v", err))
	}

	return computeDigest(rawBytes)
}

func computeDigest(rawBytes []byte) string {
	h := sha256.New()
	h.Write(rawBytes)
	digest := h.Sum(nil)
	return hex.EncodeToString(digest)
}

type Checkpoint struct {
	lock       sync.RWMutex
	proposal   Proposal
	signatures []Signature
}

func (c *Checkpoint) Get() (*smartbftprotos.Proposal, []*smartbftprotos.Signature) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	p := &smartbftprotos.Proposal{
		Header:               c.proposal.Header,
		Payload:              c.proposal.Payload,
		Metadata:             c.proposal.Metadata,
		VerificationSequence: uint64(c.proposal.VerificationSequence),
	}

	signatures := make([]*smartbftprotos.Signature, 0, len(c.signatures))
	for _, sig := range c.signatures {
		signatures = append(signatures, &smartbftprotos.Signature{
			Msg:    sig.Msg,
			Value:  sig.Value,
			Signer: sig.ID,
		})
	}
	return p, signatures
}

func (c *Checkpoint) Set(proposal Proposal, signatures []Signature) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.proposal = proposal
	c.signatures = signatures
}

type Reconfig struct {
	InLatestDecision bool
	CurrentNodes     []uint64
	CurrentConfig    Configuration
}

type SyncResponse struct {
	Latest   Decision
	Reconfig ReconfigSync
}

type ReconfigSync struct {
	InReplicatedDecisions bool
	CurrentNodes          []uint64
	CurrentConfig         Configuration
}

type Version struct {
	Q        int      //投票个数
	F        int      //容错个数
	N        int      //节点个数
	LeaderID uint64   //领导者id
	Quorum   []uint64 //投票节点
	Nodes    []uint64 //所有节点

	Type int //视图类型
}

const (
	OPTIMISTIC = iota
	NORMAL
)

// EqualVersions 比较两个 Version 是否相等
func EqualVersions(v1, v2 Version) bool {
	// 逐一比较结构体字段
	return v1.Q == v2.Q &&
		v1.F == v2.F &&
		v1.N == v2.N &&
		v1.LeaderID == v2.LeaderID &&
		equalSlices(v1.Quorum, v2.Quorum) &&
		equalSlices(v1.Nodes, v2.Nodes) &&
		v1.Type == v2.Type
}

// equalUint64Slices 比较两个 uint64 切片是否相等
func equalSlices(slice1, slice2 []uint64) bool {
	if len(slice1) != len(slice2) {
		return false
	}
	for i, v := range slice1 {
		if v != slice2[i] {
			return false
		}
	}
	return true
}
