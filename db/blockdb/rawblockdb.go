package blockdb

import (
	"errors"
	"github.com/grupokindynos/ogen/p2p"
	"github.com/grupokindynos/ogen/primitives"
	"github.com/grupokindynos/ogen/utils/chainhash"
	"github.com/grupokindynos/ogen/utils/serializer"
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
	fileNum     uint32
	blockOffset uint32
	blockSize   uint32
}

func (bl *BlockLocation) GetSize() uint32 {
	return bl.blockSize
}

func (bl *BlockLocation) Serialize(w io.Writer) error {
	err := serializer.WriteElements(w, bl.fileNum, bl.blockOffset, bl.blockSize)
	if err != nil {
		return err
	}
	return nil
}

func (bl *BlockLocation) Deserialize(r io.Reader) error {
	err := serializer.ReadElements(r, &bl.fileNum, &bl.blockOffset, &bl.blockSize)
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

func (rawbdb *RawBlockDB) ConnectBlock(block *primitives.Block) (BlockLocation, error) {
	// Add to RawBlockDB
	locator, err := rawbdb.write(block.Bytes)
	if err != nil {
		return BlockLocation{}, err
	}
	return locator, nil
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
		fileNum:     uint32(rawbdb.lastFileNumber),
		blockOffset: uint32(fileInfo.Size()),
		blockSize:   uint32(len(block)),
	}
	return blockLocator, nil
}

func (rawbdb *RawBlockDB) read(hash chainhash.Hash, locator BlockLocation) ([]byte, error) {
	var selectedFile *os.File
	curFileNum, err := strconv.Atoi(strings.Split(rawbdb.currFile.Name(), "-")[1])
	if locator.fileNum == uint32(curFileNum) {
		selectedFile = rawbdb.currFile
	} else {
		file, err := os.OpenFile(rawbdb.path+"/"+FilePrefix+strconv.Itoa(int(locator.fileNum))+".dat", os.O_RDWR, os.ModePerm)
		if err != nil {
			return nil, err
		}
		selectedFile = file
	}
	blockData := make([]byte, locator.blockSize)
	_, err = selectedFile.ReadAt(blockData, int64(locator.blockOffset))
	if err != nil {
		return nil, err
	}
	if chainhash.DoubleHashH(blockData[0:p2p.MaxBlockHeaderBytes]) != hash {
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
