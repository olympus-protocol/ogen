package db

import (
	"bytes"
	"database/sql"
	"encoding/hex"
	"github.com/olympus-protocol/ogen/pkg/bls"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"math"
	"os"
	"testing"
)

func TestByteInsertion(t *testing.T) {
	tempFilename := TempFilename(t)
	defer os.Remove(tempFilename)
	db, err := sql.Open("sqlite3", tempFilename)
	if err != nil {
		t.Fatal("Failed to open database:", err)
	}
	defer db.Close()

	_, err = db.Exec("drop table foo")
	_, err = db.Exec("create table foo (hash text)")
	if err != nil {
		t.Fatal("Failed to create table:", err)
	}

	randKey := bls.RandKey()
	pubkey, err := randKey.PublicKey().Hash()
	if err != nil {
		t.Fatal("Failed to create pubKey:", err)
	}
	prepQuery := "insert into foo(hash) values('" + hex.EncodeToString(pubkey[:]) + "');"
	res, err := db.Exec(prepQuery)
	if err != nil {
		t.Fatal("Failed to insert record:", err)
	}
	affected, _ := res.RowsAffected()
	if affected != 1 {
		t.Fatalf("Expected %d for affected rows, but %d:", 1, affected)
	}

	rows, err := db.Query("select hash from foo")
	if err != nil {
		t.Fatal("Failed to select records:", err)
	}
	defer rows.Close()

	rows.Next()

	var result string
	rows.Scan(&result)
	decodedKey, err := hex.DecodeString(result)
	assert.NoError(t, err)
	comparison := bytes.Compare(decodedKey, pubkey[:])
	assert.Equal(t, 0, comparison)

}

func TestUInt64Insertion(t *testing.T) {
	tempFilename := TempFilename(t)
	defer os.Remove(tempFilename)
	db, err := sql.Open("sqlite3", tempFilename)
	if err != nil {
		t.Fatal("Failed to open database:", err)
	}
	defer db.Close()

	_, err = db.Exec("drop table foo")
	_, err = db.Exec("create table foo (numb integer)")
	if err != nil {
		t.Fatal("Failed to create table:", err)
	}
	var num uint64
	num = math.MaxUint64

	prepQuery := "insert into foo(numb) values(?);"
	stmt, err := db.Prepare(prepQuery)
	if err != nil {
		t.Fatal("Failed to insert record:", err)
	}
	res, err := stmt.Exec(int(num))
	if err != nil {
		t.Fatal("Failed to insert record:", err)
	}
	affected, _ := res.RowsAffected()
	if affected != 1 {
		t.Fatalf("Expected %d for affected rows, but %d:", 1, affected)
	}

	rows, err := db.Query("select numb from foo")
	if err != nil {
		t.Fatal("Failed to select records:", err)
	}
	defer rows.Close()

	rows.Next()

	var result int
	rows.Scan(&result)
	res64 := uint64(result)

	assert.Equal(t, num, res64)

}

func TempFilename(t *testing.T) string {
	f, err := ioutil.TempFile("", "go-sqlite3-test-")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	return f.Name()
}
