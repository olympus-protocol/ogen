package filedb

import (
	"bytes"
	"errors"
	"github.com/grupokindynos/ogen/utils/serializer"
	"os"
	"time"
)

type FileDB struct {
	File *os.File
}

func (db *FileDB) GetMultiByIndex(values ...int) (map[int][]byte, error) {
	valueMap := make(map[int][]byte)
	for _, i := range values {
		data, err := db.GetByIndex(i)
		if err != nil {
			return nil, err
		}
		valueMap[i] = data
	}
	return valueMap, nil
}

func (db *FileDB) GetByIndex(n int) ([]byte, error) {
	metaData, err := db.GetMeta()
	if err != nil {
		return nil, err
	}
	element := make([]byte, metaData.MaxElemSize)
	offset := metaData.MaxElemSize*int64(n) + int64(metaData.serializedSize())
	_, err = db.File.ReadAt(element[:], offset)
	if err != nil {
		return nil, err
	}
	return element[:], nil
}

func (db *FileDB) GetAll() ([][]byte, error) {
	var byteArray [][]byte
	fInfo, err := db.File.Stat()
	if err != nil {
		return nil, err
	}
	metaData, err := db.GetMeta()
	if err != nil {
		return nil, err
	}
	dataWithoutMeta := make([]byte, fInfo.Size()-int64(metaData.serializedSize()))
	_, err = db.File.ReadAt(dataWithoutMeta[:], int64(metaData.serializedSize()))
	if err != nil {
		return nil, err
	}
	dataSize := len(dataWithoutMeta)
	if dataSize == 0 {
		return byteArray, nil
	}
	elements := int64(dataSize) / metaData.MaxElemSize
	for i := int64(0); i < elements; i++ {
		byteArray = append(byteArray, dataWithoutMeta[i*metaData.MaxElemSize:(i+1)*metaData.MaxElemSize])
	}
	return byteArray, nil
}

// Update value methods

func (db *FileDB) UpdateMulti(data map[int][]byte) error {
	for k, v := range data {
		err := db.UpdateByIndex(k, v)
		if err != nil {
			return err
		}
	}
	return nil
}

func (db *FileDB) UpdateByIndex(index int, data []byte) error {
	metaData, err := db.GetMeta()
	if err != nil {
		return err
	}
	if int64(len(data)) > metaData.MaxElemSize {
		return errors.New("unable to store element, size exceed Max Element Size")
	}
	dataElement := make([]byte, metaData.MaxElemSize)
	buf := bytes.NewBuffer(dataElement[:0])
	buf.Write(data)
	offset := int64(index)*metaData.MaxElemSize + int64(metaData.serializedSize())
	_, err = db.File.WriteAt(dataElement[:], offset)
	if err != nil {
		return err
	}
	return nil
}

// Add value methods

func (db *FileDB) Add(data []byte) error {
	fInfo, err := db.File.Stat()
	if err != nil {
		return err
	}
	metaData, err := db.GetMeta()
	if err != nil {
		return err
	}
	if int64(len(data)) > metaData.MaxElemSize {
		return errors.New("unable to store element, size exceed Max Element Size")
	}
	dataByte := make([]byte, metaData.MaxElemSize)
	buf := bytes.NewBuffer(dataByte[:0])
	buf.Write(data)
	_, err = db.File.WriteAt(dataByte[:], fInfo.Size())
	if err != nil {
		return err
	}
	return nil
}

// MetaData Methods

func (db *FileDB) UpdateMeta(newMeta MetaData) error {
	newMeta.Timestamp = time.Now().Unix()
	metaBytes := make([]byte, newMeta.serializedSize())
	buf := bytes.NewBuffer(metaBytes[:0])
	err := newMeta.serialize(buf)
	if err != nil {
		return err
	}
	_, err = db.File.WriteAt(metaBytes, 0)
	if err != nil {
		return err
	}
	return nil
}

func (db *FileDB) GetMeta() (MetaData, error) {
	var metaDataSizeBytes [4]byte
	_, err := db.File.ReadAt(metaDataSizeBytes[:], 0)
	if err != nil {
		return MetaData{}, err
	}
	buf := bytes.NewBuffer(metaDataSizeBytes[:])
	var metaDataSize int32
	err = serializer.ReadElements(buf, &metaDataSize)
	if err != nil {
		return MetaData{}, err
	}
	metaDataBytes := make([]byte, metaDataSize)
	var metaData MetaData
	_, err = db.File.ReadAt(metaDataBytes[:], 0)
	if err != nil {
		return MetaData{}, err
	}
	newBuf := bytes.NewBuffer(metaDataBytes[:])
	err = metaData.deserialize(newBuf)
	if err != nil {
		return MetaData{}, err
	}
	return metaData, nil
}

func NewFileDB(path string, metaParams MetaData) (*FileDB, error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0777)
	if err != nil {
		return nil, err
	}
	filedb := FileDB{
		File: file,
	}
	err = filedb.UpdateMeta(metaParams)
	if err != nil {
		return nil, err
	}
	return &filedb, nil
}
