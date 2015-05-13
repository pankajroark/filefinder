package lib

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"hash/fnv"
	"os"
)

// Note that max size of string is 2bytes

type Node struct {
	offset uint32
	next   *Node
}

type OffsetTable struct {
	offsetTable []*Node
	size        uint32
}

func NewOffsetTable(capacity uint32) *OffsetTable {
	ot := make([]*Node, capacity)
	o := OffsetTable{offsetTable: ot, size: 0}
	return &o
}

func (t *OffsetTable) capacity() uint32 {
	return uint32(len(t.offsetTable))
}

func (t *OffsetTable) put(hash, offset uint32) {
	slot := hash % uint32(len(t.offsetTable))
	t.offsetTable[slot] = &Node{offset: offset, next: t.offsetTable[slot]}
	t.size += 1
}

func (t *OffsetTable) get(hash uint32) *Node {
	slot := hash % uint32(len(t.offsetTable))
	return t.offsetTable[slot]
}

func (t *OffsetTable) rehashNeeded() bool {
	return float64(t.size) > 0.9*float64(len(t.offsetTable))
}

func (t *OffsetTable) forAll(f func(uint32)) {
	var i uint32
	for i = 0; i < t.capacity(); i++ {
		node := t.offsetTable[i]
		for node != nil {
			f(node.offset)
			node = node.next
		}
	}
}

type Stringids struct {
	indexPath   string
	wal         *os.File
	walSize     uint32
	offsetTable *OffsetTable
}

func NewStringids(path string) *Stringids {
	wal, _ := os.OpenFile(path, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0660)
	fi, e := wal.Stat()
	if e != nil {
		panic(e)
	}
	walSize := uint32(fi.Size())
	offsetTable := NewOffsetTable(1024)
	strids := &Stringids{indexPath: path, wal: wal, walSize: walSize, offsetTable: offsetTable}
	// todo read wal here
	strids.loadOffsetTableFromWal()
	return strids
}

func (s *Stringids) loadOffsetTableFromWal() {
	fmt.Println("Reading WAL...")
	var offset uint32
	offset = 0
	for {
		str, e := s.StrAtOffset(offset)
		if e != nil {
			break
		}
		s.storeOffset(str, offset)
		offset += uint32(len(str) + 2)
	}
	fmt.Println("Finished reading WAL...")
}

func (s *Stringids) rehash() {
	fmt.Println("rehashing...")
	nt := NewOffsetTable(2 * s.offsetTable.capacity())
	anon := func(offset uint32) {
		str, _ := s.StrAtOffset(offset)
		h := s.hash(str)
		nt.put(h, offset)
	}
	s.offsetTable.forAll(anon)
	s.offsetTable = nt
}

func (s *Stringids) reset() {
	s.wal, _ = os.OpenFile(s.indexPath, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0660)
	fi, e := s.wal.Stat()
	if e != nil {
		panic(e)
	}
	s.walSize = uint32(fi.Size())
	s.offsetTable = NewOffsetTable(s.offsetTable.capacity())
}

func (s *Stringids) hash(str string) uint32 {
	hash := fnv.New32()
	_, e := hash.Write([]byte(str))
	if e != nil {
		panic(e)
	}
	return hash.Sum32()
}

func (s *Stringids) writeToWal(str string) uint32 {
	offset := s.walSize
	binary.Write(s.wal, binary.LittleEndian, uint16(len(str)))
	n, err := s.wal.Write([]byte(str))
	if err != nil {
		panic(err)
	}
	s.walSize += uint32(n) + 2 // 2 bytes for size
	return offset
}

func (s *Stringids) storeOffset(str string, offset uint32) {
	s.offsetTable.put(s.hash(str), offset)
	if s.offsetTable.rehashNeeded() {
		s.rehash()
	}
}

func (s *Stringids) Add(str string) uint32 {
	offset, err := s.GetOffset(str)
	if err != nil {
		offset = s.writeToWal(str)
		s.storeOffset(str, offset)
	}
	return offset
}

func (s *Stringids) StrAtOffset(offset uint32) (string, error) {
	var size uint16
	ba := make([]byte, 2)
	_, e := s.wal.ReadAt(ba, int64(offset))
	if e != nil {
		return "", e
	}
	reader := bytes.NewReader(ba)
	binary.Read(reader, binary.LittleEndian, &size)
	ba = make([]byte, size)
	offset += 2
	s.wal.ReadAt(ba, int64(offset))
	return string(ba), nil
}

func (s *Stringids) GetOffset(str string) (uint32, error) {
	node := s.offsetTable.get(s.hash(str))
	for node != nil {
		tstr, _ := s.StrAtOffset(node.offset)
		if tstr == str {
			return node.offset, nil
		}
		node = node.next
	}
	return 0, errors.New("not found")
}

func (s *Stringids) Clear() {
	err := os.Remove(s.indexPath)
	if err != nil {
		panic(err)
	}
	s.reset()
}
