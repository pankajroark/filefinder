package ds

import (
	"bytes"
	"encoding/binary"
)

/*
This is a hash table where key are uint32 but values are any byte array.
The value is stored in a write ahead log and offset stored in a hashtable, thus
providing random access to value while keeping memory usage organized in a
serialization friendly way.
*/

type IntKeyHashTable struct {
	innerHT IntHashTable
	// We could use a file here but let's start with a buffer for now
	buf *bytes.Buffer
}

func CreateIntKeyHashTable() *IntKeyHashTable {
	ht := IntKeyHashTable{}
	ht.innerHT = CreateIntHashTable(1024)
	ht.buf = bytes.NewBuffer(make([]byte, 1024*4))
	return &ht
}

func (ht *IntKeyHashTable) Put(k uint32, v []byte) {
	// but in buffer get offset, store offset in inner ht
	offset := uint32(ht.buf.Len())
	// write length then write bytes
	valueLen := uint32(len(v))
	valueLenBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(valueLenBytes, valueLen)
	ht.buf.Write(valueLenBytes)
	ht.buf.Write(v)
	ht.innerHT.Put(k, offset)
}

func (ht *IntKeyHashTable) Get(k uint32) ([]byte, bool) {
	offset, found := ht.innerHT.Get(k)
	if !found {
		return nil, false
	}
	valueLenBytes := make([]byte, 4)
	bufAtOffset := bytes.NewBuffer(ht.buf.Bytes()[offset:])
	bufAtOffset.Read(valueLenBytes) // todo: handle error
	valueLen := binary.LittleEndian.Uint32(valueLenBytes)
	valueBytes := make([]byte, valueLen)
	bufAtOffset.Read(valueBytes) // todo handle error
	return valueBytes, true
}
