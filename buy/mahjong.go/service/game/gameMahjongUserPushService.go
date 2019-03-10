package game

import (
	"sync"

	"mahjong.go/mi/protocal"
)

// MSeqContainer 有序消息容器
type MSeqContainer struct {
	list []*SeqOperation
	seq  int // 最大消息编号
	mux  *sync.Mutex
}

// SeqOperation 牌局中发给用户的消息
type SeqOperation struct {
	seq        int // 消息编号
	wOperation []*Operation
	uOperation *UserOperation
	cOperation *ClientOperation
}

// NewMSeqContainer 生成有序消息容器
func NewMSeqContainer() *MSeqContainer {
	return &MSeqContainer{
		list: make([]*SeqOperation, 0),
		seq:  0,
		mux:  &sync.Mutex{},
	}
}

// GetSeq 获取最后的消息编号
func (msc *MSeqContainer) GetSeq() int {
	return msc.seq
}

// GetList 获取消息列表
func (msc *MSeqContainer) GetList(seq int) []*SeqOperation {
	msc.mux.Lock()
	defer msc.mux.Unlock()
	if seq >= msc.seq {
		return nil
	}
	return msc.list[seq:]
}

// AddWOperation 添加operationPush
func (msc *MSeqContainer) AddWOperation(opList []*Operation) *SeqOperation {
	msc.mux.Lock()
	defer msc.mux.Unlock()
	msc.seq++
	seqOperation := &SeqOperation{
		seq:        msc.seq,
		wOperation: opList,
	}
	msc.list = append(msc.list, seqOperation)
	return seqOperation
}

// AddUOperation 添加UserOperation
func (msc *MSeqContainer) AddUOperation(operation *UserOperation) *SeqOperation {
	msc.mux.Lock()
	defer msc.mux.Unlock()
	msc.seq++
	seqOperation := &SeqOperation{
		seq:        msc.seq,
		uOperation: operation,
	}
	msc.list = append(msc.list, seqOperation)
	return seqOperation

}

// AddCOperation 添加ClientOpetion
func (msc *MSeqContainer) AddCOperation(operation *ClientOperation) *SeqOperation {
	msc.mux.Lock()
	defer msc.mux.Unlock()
	msc.seq++
	seqOperation := &SeqOperation{
		seq:        msc.seq,
		cOperation: operation,
	}
	msc.list = append(msc.list, seqOperation)
	return seqOperation
}

// ToImPacket 打包, 直接生成可发送的impacket
// 根据传入参数，决定打包时是否带上seq
func (so *SeqOperation) ToImPacket(enableSeq bool) *protocal.ImPacket {
	var seq int
	if enableSeq {
		seq = so.seq
	}
	if so.wOperation != nil {
		return OperationPushWithSeq(so.wOperation, seq)
	} else if so.uOperation != nil {
		return UserOperationPushWithSeq(so.uOperation, seq)
	}
	return ClientOperationPushWithSeq(so.cOperation, seq)
}
