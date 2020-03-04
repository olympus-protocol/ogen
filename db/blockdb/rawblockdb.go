package blockdb

import (
	"bytes"
	"errors"
	"github.com/olympus-protocol/ogen/primitives"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"github.com/olympus-protocol/ogen/utils/serializer"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

const (
	FilePrefix      = "blk-"
	MaxBytesforFile = 1024 * 1024 * 100 // 150 MB
)

type BlockLocation struct {
	FileNum     uint32
	BlockOffset uint32
	BlockSize   uint32
}

func (bl *BlockLocation) GetSize() uint32 {
	return bl.BlockSize
}

func (bl *BlockLocation) Serialize(w io.Writer) error {
	err := serializer.WriteElements(w, bl.FileNum, bl.BlockOffset, bl.BlockSize)
	if err != nil {
		return err
	}
	return nil
}

func (bl *BlockLocation) Deserialize(r io.Reader) error {
	err := serializer.ReadElements(r, &bl.FileNum, &bl.BlockOffset, &bl.BlockSize)
	if err != nil {
		return err
	}
	return nil
}

type RawBlockDB struct {
	path           string
	lastFileNumber int
	filesInfo      []os.FileInfo
	currFile       *os.File
	currFileLock   sync.RWMutex
}

func (rawbdb *RawBlockDB) AddBlock(block *primitives.Block) (*BlockLocation, error) {
	// Add to RawBlockDB
	buf := bytes.NewBuffer([]byte{})
	_ = block.Encode(buf)
	locator, err := rawbdb.write(buf.Bytes())
	if err != nil {
		return nil, err
	}
	return &locator, nil
}

func (rawbdb *RawBlockDB) write(block []byte) (BlockLocation, error) {
	// Before write, check if there is a file loaded
	var wg sync.WaitGroup
	if rawbdb.currFile == nil {
		wg.Add(1)
		err := rawbdb.newFile(&wg)
		if err != nil {
			return BlockLocation{}, err
		}
		wg.Wait()
	}
info:
	// Get current file stats
	fileInfo, err := rawbdb.currFile.Stat()
	if err != nil {
		return BlockLocation{}, err
	}
	// If file size is bigger than 150 MB rotate the file
	if fileInfo.Size() > MaxBytesforFile {
		wg.Add(1)
		err := rawbdb.newFile(&wg)
		if err != nil {
			return BlockLocation{}, err
		}
		wg.Wait()
		goto info
	}
	rawbdb.currFileLock.Lock()
	_, err = rawbdb.currFile.WriteAt(block, fileInfo.Size())
	if err != nil {
		return BlockLocation{}, err
	}
	rawbdb.currFileLock.Unlock()
	blockLocator := BlockLocation{
		FileNum:     uint32(rawbdb.lastFileNumber),
		BlockOffset: uint32(fileInfo.Size()),
		BlockSize:   uint32(len(block)),
	}
	return blockLocator, nil
}

func (rawbdb *RawBlockDB) read(hash chainhash.Hash, locator BlockLocation) ([]byte, error) {
	var selectedFile *os.File
	curFileNum, err := strconv.Atoi(strings.Split(rawbdb.currFile.Name(), "-")[1])
	if locator.FileNum == uint32(curFileNum) {
		selectedFile = rawbdb.currFile
	} else {
		file, err := os.OpenFile(rawbdb.path+"/"+FilePrefix+strconv.Itoa(int(locator.FileNum))+".dat", os.O_RDWR, os.ModePerm)
		if err != nil {
			return nil, err
		}
		selectedFile = file
	}
	blockData := make([]byte, locator.BlockSize)
	_, err = selectedFile.ReadAt(blockData, int64(locator.BlockOffset))
	if err != nil {
		return nil, err
	}
	if chainhash.DoubleHashH(blockData[0:primitives.MaxBlockHeaderBytes]) != hash {
		return nil, errors.New("read block error: hashes doesn't match")
	}
	return blockData, nil
}

func (rawbdb *RawBlockDB) FullDataSize() int64 {
	var size int64
	_ = filepath.Walk(rawbdb.path, func(path string, info os.FileInfo, err error) error {
		if strings.Split(info.Name(), "-")[0] == "blk" {
			size += info.Size()
		}
		return nil
	})
	return size
}

func (rawbdb *RawBlockDB) newFile(wg *sync.WaitGroup) error {
	defer wg.Done()
	if rawbdb.currFile != nil {
		err := rawbdb.currFile.Close()
		if err != nil {
			return err
		}
	}
	file, err := os.OpenFile(rawbdb.path+"/"+FilePrefix+strconv.Itoa(rawbdb.lastFileNumber+1)+".dat", os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		return err
	}
	rawbdb.lastFileNumber += 1
	rawbdb.currFile = file
	return nil
}

func NewRawBlockDB(path string) (*RawBlockDB, error) {
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return nil, err
	}
	var lastFileNum int
	_ = filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if strings.Split(info.Name(), "-")[0] == "blk" {
			fileNum, err := strconv.Atoi(strings.Split(info.Name(), "-")[1])
			if err != nil {
				return err
			}
			if fileNum > lastFileNum {
				lastFileNum = fileNum
			}
		}
		return nil
	})
	currFile, err := os.OpenFile(path+"/"+FilePrefix+strconv.Itoa(lastFileNum)+".dat", os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		return nil, err
	}
	rbdb := &RawBlockDB{
		path:           path,
		lastFileNumber: lastFileNum,
		currFile:       currFile,
	}
	return rbdb, nil
}
