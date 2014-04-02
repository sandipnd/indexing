// Code generated by protoc-gen-go.
// source: mutation.proto
// DO NOT EDIT!

package protobuf

import proto "code.google.com/p/goprotobuf/proto"
import json "encoding/json"
import math "math"

// Reference proto, json, and math imports to suppress error if they are not otherwise used.
var _ = proto.Marshal
var _ = &json.SyntaxError{}
var _ = math.Inf

// List of possible mutation commands.
type Command int32

const (
	Command_Upsert         Command = 1
	Command_Deletion       Command = 2
	Command_UpsertDeletion Command = 3
	Command_Sync           Command = 4
	Command_DropData       Command = 5
	Command_StreamBegin    Command = 6
	Command_StreamEnd      Command = 7
)

var Command_name = map[int32]string{
	1: "Upsert",
	2: "Deletion",
	3: "UpsertDeletion",
	4: "Sync",
	5: "DropData",
	6: "StreamBegin",
	7: "StreamEnd",
}
var Command_value = map[string]int32{
	"Upsert":         1,
	"Deletion":       2,
	"UpsertDeletion": 3,
	"Sync":           4,
	"DropData":       5,
	"StreamBegin":    6,
	"StreamEnd":      7,
}

func (x Command) Enum() *Command {
	p := new(Command)
	*p = x
	return p
}
func (x Command) String() string {
	return proto.EnumName(Command_name, int32(x))
}
func (x *Command) UnmarshalJSON(data []byte) error {
	value, err := proto.UnmarshalJSONEnum(Command_value, data, "Command")
	if err != nil {
		return err
	}
	*x = Command(value)
	return nil
}

// A single mutation message that will framed and transported by router.
// For efficiency mutations from mutiple vbuckets can be packed into the same
// message.
type Mutation struct {
	Version *uint32 `protobuf:"varint,1,req,name=version" json:"version,omitempty"`
	// -- Following fields are mutually exclusive --
	Vbmutations      []*VbucketMutation `protobuf:"bytes,2,rep,name=vbmutations" json:"vbmutations,omitempty"`
	Vbuckets         *VbConnectionMap   `protobuf:"bytes,3,opt,name=vbuckets" json:"vbuckets,omitempty"`
	XXX_unrecognized []byte             `json:"-"`
}

func (m *Mutation) Reset()         { *m = Mutation{} }
func (m *Mutation) String() string { return proto.CompactTextString(m) }
func (*Mutation) ProtoMessage()    {}

func (m *Mutation) GetVersion() uint32 {
	if m != nil && m.Version != nil {
		return *m.Version
	}
	return 0
}

func (m *Mutation) GetVbmutations() []*VbucketMutation {
	if m != nil {
		return m.Vbmutations
	}
	return nil
}

func (m *Mutation) GetVbuckets() *VbConnectionMap {
	if m != nil {
		return m.Vbuckets
	}
	return nil
}

// Mutation per vbucket, mutations are broadly divided into data and
// control messages. The division is based on the commands.
//
// Interpreting seq.no:
// 1. For Upsert, Deletion, UpsertDeletion messages, sequence number corresponds
//    to kv mutation.
// 2. For Sync message, it is the latest kv mutation sequence-no. received for
//   a vbucket.
// 3. For DropData message, it is the first kv mutation that was dropped due
//    to buffer overflow.
// 4. For StreamBegin, it is the first kv mutation received after opening a
//    vbucket stream with kv.
// 5. For StreamEnd, it is the last kv mutation received before ending a vbucket
//    stream with kv.
//
// fields `docid`, `indexids`, `keys`, `oldkeys` are valid only for
// Upsert, Deletion, UpsertDeletion messages.
type VbucketMutation struct {
	Command          *uint32  `protobuf:"varint,1,req,name=command" json:"command,omitempty"`
	Vbucket          *uint32  `protobuf:"varint,2,req,name=vbucket" json:"vbucket,omitempty"`
	Seqno            *uint64  `protobuf:"varint,5,req,name=seqno" json:"seqno,omitempty"`
	Vbuuid           *uint64  `protobuf:"varint,3,req,name=vbuuid" json:"vbuuid,omitempty"`
	Docid            []byte   `protobuf:"bytes,4,opt,name=docid" json:"docid,omitempty"`
	Indexids         []uint32 `protobuf:"varint,6,rep,name=indexids" json:"indexids,omitempty"`
	Keys             [][]byte `protobuf:"bytes,7,rep,name=keys" json:"keys,omitempty"`
	Oldkeys          [][]byte `protobuf:"bytes,8,rep,name=oldkeys" json:"oldkeys,omitempty"`
	XXX_unrecognized []byte   `json:"-"`
}

func (m *VbucketMutation) Reset()         { *m = VbucketMutation{} }
func (m *VbucketMutation) String() string { return proto.CompactTextString(m) }
func (*VbucketMutation) ProtoMessage()    {}

func (m *VbucketMutation) GetCommand() uint32 {
	if m != nil && m.Command != nil {
		return *m.Command
	}
	return 0
}

func (m *VbucketMutation) GetVbucket() uint32 {
	if m != nil && m.Vbucket != nil {
		return *m.Vbucket
	}
	return 0
}

func (m *VbucketMutation) GetSeqno() uint64 {
	if m != nil && m.Seqno != nil {
		return *m.Seqno
	}
	return 0
}

func (m *VbucketMutation) GetVbuuid() uint64 {
	if m != nil && m.Vbuuid != nil {
		return *m.Vbuuid
	}
	return 0
}

func (m *VbucketMutation) GetDocid() []byte {
	if m != nil {
		return m.Docid
	}
	return nil
}

func (m *VbucketMutation) GetIndexids() []uint32 {
	if m != nil {
		return m.Indexids
	}
	return nil
}

func (m *VbucketMutation) GetKeys() [][]byte {
	if m != nil {
		return m.Keys
	}
	return nil
}

func (m *VbucketMutation) GetOldkeys() [][]byte {
	if m != nil {
		return m.Oldkeys
	}
	return nil
}

// List of vbuckets that will be streamed via a newly opened connection.
type VbConnectionMap struct {
	Vbno             []uint32 `protobuf:"varint,1,rep,name=vbno" json:"vbno,omitempty"`
	Vbuuid           []uint64 `protobuf:"varint,2,rep,name=vbuuid" json:"vbuuid,omitempty"`
	XXX_unrecognized []byte   `json:"-"`
}

func (m *VbConnectionMap) Reset()         { *m = VbConnectionMap{} }
func (m *VbConnectionMap) String() string { return proto.CompactTextString(m) }
func (*VbConnectionMap) ProtoMessage()    {}

func (m *VbConnectionMap) GetVbno() []uint32 {
	if m != nil {
		return m.Vbno
	}
	return nil
}

func (m *VbConnectionMap) GetVbuuid() []uint64 {
	if m != nil {
		return m.Vbuuid
	}
	return nil
}

func init() {
	proto.RegisterEnum("protobuf.Command", Command_name, Command_value)
}
