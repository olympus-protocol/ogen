package filedb

import (
	"bytes"
	"github.com/olympus-protocol/ogen/utils/serializer"
	"io"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

var filedb *FileDB

type TestElement struct {
	Name    string
	Payload []byte
}

const NumElementsToAdd = 1000

func (te *TestElement) serialize(w io.Writer) (err error) {
	err = serializer.WriteVarString(w, te.Name)
	if err != nil {
		return err
	}
	err = serializer.WriteVarBytes(w, te.Payload)
	if err != nil {
		return err
	}
	return nil
}

func (te *TestElement) deserialize(r io.Reader) (err error) {
	te.Name, err = serializer.ReadVarString(r)
	if err != nil {
		return err
	}
	te.Payload, err = serializer.ReadVarBytes(r)
	if err != nil {
		return err
	}
	return nil
}

var testArr []TestElement

func init() {
	// Delete mock file if exists
	_ = os.Remove("./datbase_test.dat")
	// Create mock file
	db, err := NewFileDB("./datbase_test.dat", metaData)
	if err != nil {
		panic(err)
	}
	filedb = db
	// Create the test map with mock data
	for i := 0; i < NumElementsToAdd; i++ {
		newElement := TestElement{
			Name:    "name-" + strconv.Itoa(i),
			Payload: bytes.Repeat([]byte{byte(1), byte(2), byte(3), byte(4), byte(5)}, 30),
		}
		testArr = append(testArr, newElement)
	}

	// Fill the database file with the test map
	for _, v := range testArr {
		buf := bytes.NewBuffer([]byte{})
		err := v.serialize(buf)
		if err != nil {
			panic(err)
		}
		err = filedb.Add(buf.Bytes())
		if err != nil {
			panic(err)
		}
	}

}

func TestFileDB_GetMeta(t *testing.T) {
	testMetaData, err := filedb.GetMeta()
	if err != nil {
		t.Errorf("Test GetMeta error: %v", err.Error())
		return
	}
	if testMetaData.Version != metaData.Version {
		t.Errorf("Test GetMeta error: Version doesn't match")
	}
	if testMetaData.Name != metaData.Name {
		t.Errorf("Test GetMeta error: Name doesn't match")
	}
	if testMetaData.MaxElemSize != metaData.MaxElemSize {
		t.Errorf("Test GetMeta error: element size doesn't match")
	}
}

func TestFileDB_UpdateMeta(t *testing.T) {
	newMeta := MetaData{
		Version:     1010101,
		Timestamp:   time.Now().Unix(),
		Name:        "new-meta-for-testing-updat",
		MaxElemSize: 250,
	}
	err := filedb.UpdateMeta(newMeta)
	if err != nil {
		t.Errorf("UpdateMeta error: %v", err.Error())
	}
	testMetaUpdated, err := filedb.GetMeta()
	if err != nil {
		t.Errorf("UpdateMeta error: %v", err.Error())
		return
	}
	if testMetaUpdated.Name != newMeta.Name {
		t.Errorf("UpdateMeta meta not updating correctly")
	}
	if testMetaUpdated.MaxElemSize != newMeta.MaxElemSize {
		t.Errorf("UpdateMeta max element size doesn't match")
	}
}

func TestFileDB_GetByIndex(t *testing.T) {
	index := 1
	originData := testArr[index]
	var element TestElement
	dataByte, err := filedb.GetByIndex(1)
	if err != nil {
		t.Errorf("Test GetByIndex error: %v", err.Error())
	}
	buf := bytes.NewBuffer(dataByte)
	err = element.deserialize(buf)
	if err != nil {
		t.Errorf("Test GetByIndex error: %v", err.Error())
	}
	if element.Name != originData.Name {
		t.Errorf("Test GetByIndex error: name doesn't match")
	}
	compare := bytes.Compare(element.Payload, originData.Payload)
	if compare != 0 {
		t.Errorf("Test GetByIndex error: payload doesn't match")
	}
}

func TestFileDB_GetMultiByIndex(t *testing.T) {
	dataByteArray, err := filedb.GetMultiByIndex(1, 10, 200, 500)
	if err != nil {
		t.Errorf("Test GetMultiByIndex error: %v", err.Error())
	}
	for k, v := range dataByteArray {
		var element TestElement
		buf := bytes.NewBuffer(v)
		err := element.deserialize(buf)
		if err != nil {
			t.Errorf("Test GetMultiByIndex error: %v", err.Error())
		}
		originData := testArr[k]
		if element.Name != originData.Name {
			t.Errorf("Test GetByIndex error: name doesn't match")
		}
		compare := bytes.Compare(element.Payload, originData.Payload)
		if compare != 0 {
			t.Errorf("Test GetByIndex error: payload doesn't match")
		}
	}
}

func TestFileDB_GetAll(t *testing.T) {
	dataByteArray, err := filedb.GetAll()
	if err != nil {
		t.Errorf("Test GetMultiByIndex error: %v", err.Error())
	}
	if len(dataByteArray) != len(testArr) {
		t.Errorf("Test GetMultiByIndex error: data size doesn't match")
	}
	for _, byteArray := range dataByteArray {
		var element TestElement
		buf := bytes.NewBuffer(byteArray)
		err := element.deserialize(buf)
		if err != nil {
			t.Errorf("Test GetMultiByIndex error: %v", err.Error())
		}
		// Since information returns non-sorted we must extract dbindex from deserialized data to compare
		nameStr := strings.Split(element.Name, "-")
		index, _ := strconv.Atoi(nameStr[1])
		originData := testArr[index]
		if element.Name != originData.Name {
			t.Errorf("Test GetByIndex error: name doesn't match")
		}
		compare := bytes.Compare(element.Payload, originData.Payload)
		if compare != 0 {
			t.Errorf("Test GetByIndex error: payload doesn't match")
		}
	}

}

func TestFileDB_Add(t *testing.T) {
	newElement := TestElement{
		Name:    "newTestElement",
		Payload: bytes.Repeat([]byte{byte(1), byte(2), byte(3), byte(4), byte(5)}, 30),
	}
	buf := bytes.NewBuffer([]byte{})
	err := newElement.serialize(buf)
	if err != nil {
		t.Errorf("Test Add error: %v", err.Error())
	}
	err = filedb.Add(buf.Bytes())
	if err != nil {
		t.Errorf("Test Add error: %v", err.Error())
	}
	storedElement, err := filedb.GetByIndex(NumElementsToAdd)
	bufStored := bytes.NewBuffer(storedElement)
	var element TestElement
	err = element.deserialize(bufStored)
	if err != nil {
		t.Errorf("Test Add error: %v", err.Error())
	}
	if element.Name != newElement.Name {
		t.Errorf("Test GetByIndex error: name doesn't match")
	}
	compare := bytes.Compare(element.Payload, newElement.Payload)
	if compare != 0 {
		t.Errorf("Test GetByIndex error: payload doesn't match")
	}
	finish()
}

func finish() {
	_ = os.Remove("./datbase_test.dat")
}
