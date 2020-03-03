package users_txverifier

import (
	"bytes"
	"github.com/olympus-protocol/ogen/chain/index"
	"github.com/olympus-protocol/ogen/p2p"
	"github.com/olympus-protocol/ogen/params"
	users_txpayload "github.com/olympus-protocol/ogen/txs/txpayloads/users"
	"github.com/olympus-protocol/ogen/users"
	"github.com/olympus-protocol/ogen/utils/chainhash"
	"testing"
)

var userIndexMock = index.InitUsersIndex()

var user UsersTxVerifier

// 1: 39f02d7bf429b4c2da4e18587e60b8eca633c24d95a6fc6901f4a67f82222554
// 2: 64b2f2bd9dec35cb584fab86647c5f5a9f96529833a85b7ad735b2296b1314ac
// 3: 44e3c44315ca78c0d850989fc9736ac8acc1f390fce60fc146a36f69318243ae
// 4: 1ef3fec45cae4f90277111e3700ac45a46e68044d44435aabbf02f57777f730a
// 5: 4ebec5ce08d9f5efad027f92db80498dc49da19b7af03bf5a177dde8285b1f1d

var PubKeyUser1 = [48]byte{179, 67, 240, 201, 137, 190, 30, 1, 247, 45, 249, 144, 12, 10, 52, 11, 22, 230, 26, 175, 186, 115, 255, 94, 202, 85, 170, 43, 92, 178, 15, 82, 5, 11, 198, 222, 215, 196, 100, 25, 113, 38, 27, 63, 118, 106, 26, 134}
var PubKeyUser2 = [48]byte{140, 83, 161, 192, 242, 252, 248, 74, 211, 62, 63, 185, 235, 239, 61, 232, 46, 49, 205, 193, 162, 155, 112, 25, 44, 17, 161, 115, 31, 76, 53, 238, 127, 243, 89, 81, 223, 62, 223, 80, 172, 147, 138, 150, 107, 207, 86, 47}
var PubKeyUser3 = [48]byte{170, 202, 170, 223, 168, 218, 211, 199, 212, 172, 121, 71, 35, 40, 116, 145, 198, 77, 255, 68, 123, 43, 20, 252, 179, 177, 139, 54, 121, 181, 85, 36, 67, 75, 98, 159, 0, 185, 91, 11, 251, 215, 59, 148, 193, 52, 49, 175}
var PubKeyUser4 = [48]byte{170, 58, 157, 252, 139, 253, 192, 127, 63, 240, 248, 76, 115, 156, 233, 94, 91, 212, 171, 161, 35, 207, 155, 79, 213, 245, 63, 194, 122, 181, 37, 133, 230, 68, 213, 126, 100, 193, 0, 132, 3, 91, 126, 5, 130, 90, 12, 177}
var PubKeyUser5 = [48]byte{146, 160, 124, 119, 147, 133, 135, 145, 108, 241, 134, 252, 118, 145, 118, 116, 86, 114, 252, 185, 22, 34, 85, 151, 122, 8, 208, 238, 56, 95, 208, 86, 37, 100, 82, 195, 48, 54, 55, 189, 72, 21, 244, 3, 90, 18, 35, 219}

func init() {
	user = NewUsersTxVerifier(userIndexMock, &params.Mainnet)
	us := []*index.UserRow{
		{
			OutPoint: p2p.OutPoint{
				TxHash: chainhash.DoubleHashH([]byte("user-1")),
				Index:  0,
			},
			UserData: users.User{
				PubKey: PubKeyUser1,
				Name:   "test-user-1",
			},
		},
		{
			OutPoint: p2p.OutPoint{
				TxHash: chainhash.DoubleHashH([]byte("user-2")),
				Index:  0,
			},
			UserData: users.User{
				PubKey: PubKeyUser2,
				Name:   "test-user-2",
			},
		},
		{
			OutPoint: p2p.OutPoint{
				TxHash: chainhash.DoubleHashH([]byte("user-3")),
				Index:  0,
			},
			UserData: users.User{
				PubKey: PubKeyUser3,
				Name:   "test-user-3",
			},
		},
		{
			OutPoint: p2p.OutPoint{
				TxHash: chainhash.DoubleHashH([]byte("user-4")),
				Index:  0,
			},
			UserData: users.User{
				PubKey: PubKeyUser4,
				Name:   "test-user-4",
			},
		},
		{
			OutPoint: p2p.OutPoint{
				TxHash: chainhash.DoubleHashH([]byte("user-5")),
				Index:  0,
			},
			UserData: users.User{
				PubKey: PubKeyUser5,
				Name:   "test-user-5",
			},
		},
	}
	for _, user := range us {
		userIndexMock.Add(user)
	}
}

var mockPayloadUpload1 = users_txpayload.PayloadUpload{
	PubKey: PubKeyUser1,
	Sig:    [96]byte{132, 15, 236, 68, 246, 182, 105, 205, 15, 103, 166, 157, 63, 117, 67, 65, 165, 196, 5, 156, 192, 221, 194, 192, 162, 19, 21, 226, 72, 64, 20, 232, 189, 167, 9, 215, 11, 137, 10, 234, 240, 21, 232, 165, 81, 152, 116, 110, 11, 57, 138, 135, 106, 49, 41, 141, 253, 185, 41, 88, 17, 216, 8, 138, 25, 228, 103, 43, 145, 9, 126, 219, 81, 37, 118, 40, 228, 97, 94, 219, 30, 187, 118, 187, 18, 181, 25, 115, 117, 166, 197, 216, 40, 224, 29, 186},
	Name:   "new-user-1",
}
var mockPayloadUpload2 = users_txpayload.PayloadUpload{
	PubKey: PubKeyUser1,
	Sig:    [96]byte{180, 86, 41, 247, 190, 110, 203, 34, 188, 79, 91, 197, 246, 58, 196, 222, 142, 123, 251, 113, 248, 64, 215, 117, 241, 69, 67, 239, 60, 76, 111, 131, 154, 138, 207, 41, 177, 224, 227, 183, 8, 133, 216, 226, 30, 144, 82, 23, 6, 222, 101, 232, 142, 232, 115, 105, 38, 105, 70, 215, 203, 192, 114, 158, 123, 45, 91, 120, 39, 167, 229, 6, 192, 252, 81, 127, 200, 89, 22, 97, 190, 81, 204, 35, 53, 197, 95, 246, 68, 102, 180, 220, 42, 94, 158, 165},
	Name:   "new-user-2",
}
var mockPayloadUpload3 = users_txpayload.PayloadUpload{
	PubKey: PubKeyUser1,
	Sig:    [96]byte{170, 95, 215, 59, 196, 101, 240, 32, 230, 178, 165, 67, 63, 123, 225, 135, 80, 85, 205, 38, 167, 247, 118, 71, 126, 6, 179, 246, 13, 62, 94, 116, 118, 253, 17, 255, 12, 98, 200, 120, 208, 158, 93, 169, 212, 67, 110, 148, 15, 249, 32, 138, 87, 76, 156, 223, 255, 101, 185, 155, 71, 171, 157, 67, 150, 162, 189, 61, 178, 121, 79, 82, 166, 43, 95, 113, 87, 146, 130, 204, 106, 20, 82, 165, 102, 54, 116, 3, 17, 221, 146, 22, 81, 30, 86, 164},
	Name:   "new-user-3",
}
var mockPayloadUpload4 = users_txpayload.PayloadUpload{
	PubKey: PubKeyUser1,
	Sig:    [96]byte{151, 40, 3, 135, 50, 185, 38, 6, 150, 27, 201, 223, 201, 154, 164, 231, 8, 7, 178, 17, 185, 89, 236, 220, 225, 211, 186, 123, 202, 129, 205, 68, 190, 151, 189, 231, 68, 161, 191, 48, 116, 91, 155, 144, 206, 81, 142, 72, 6, 100, 233, 232, 254, 51, 47, 35, 42, 86, 63, 239, 27, 213, 179, 176, 210, 211, 56, 52, 30, 142, 178, 95, 211, 238, 192, 24, 137, 55, 174, 23, 51, 4, 48, 226, 63, 140, 166, 193, 244, 12, 233, 192, 26, 40, 3, 144},
	Name:   "new-user-4",
}
var mockPayloadUpload5 = users_txpayload.PayloadUpload{
	PubKey: PubKeyUser1,
	Sig:    [96]byte{179, 91, 153, 1, 191, 34, 50, 137, 120, 175, 27, 161, 86, 212, 203, 160, 74, 53, 146, 111, 193, 224, 250, 141, 84, 196, 162, 75, 208, 93, 215, 91, 41, 3, 152, 108, 125, 129, 214, 92, 7, 69, 224, 98, 215, 235, 131, 3, 17, 120, 166, 41, 177, 55, 103, 82, 20, 255, 14, 94, 62, 118, 21, 6, 200, 189, 111, 49, 34, 79, 251, 163, 56, 167, 68, 209, 127, 106, 246, 23, 157, 17, 102, 79, 208, 193, 21, 37, 76, 126, 193, 250, 110, 95, 47, 184},
	Name:   "new-user-5",
}

func TestMockUpload(t *testing.T) {
	buf := bytes.NewBuffer([]byte{})
	err := mockPayloadUpload1.Serialize(buf)
	if err != nil {
		t.Fatal("unable to serialize")
	}
	err = user.MatchVerify(buf.Bytes(), p2p.Upload)
	if err != nil {
		t.Fatal("verification failed")
	}
	err = user.SigVerify(buf.Bytes(), p2p.Upload)
	if err != nil {
		t.Fatal("verification failed")
	}
}

var mockPayloadUploadBatch = []users_txpayload.PayloadUpload{mockPayloadUpload1, mockPayloadUpload2, mockPayloadUpload3, mockPayloadUpload4, mockPayloadUpload5}

func TestMockUploadBatch(t *testing.T) {
	var payload [][]byte
	for _, singlePayload := range mockPayloadUploadBatch {
		buf := bytes.NewBuffer([]byte{})
		err := singlePayload.Serialize(buf)
		if err != nil {
			t.Fatal("unable to serialize")
		}
		payload = append(payload, buf.Bytes())
	}
	err := user.MatchVerifyBatch(payload, p2p.Upload)
	if err != nil {
		t.Fatal("verification failed")
	}
	err = user.SigVerifyBatch(payload, p2p.Upload)
	if err != nil {
		t.Fatal("verification failed")
	}
}

var mockPayloadUpdate1 = users_txpayload.PayloadUpdate{
	NewPubKey: PubKeyUser2,
	PubKey:    PubKeyUser1,
	Sig:       [96]byte{135, 247, 99, 185, 189, 35, 91, 17, 20, 181, 45, 95, 34, 208, 38, 18, 128, 118, 226, 219, 81, 37, 174, 163, 108, 7, 200, 38, 47, 30, 19, 228, 56, 199, 113, 192, 107, 102, 201, 104, 93, 254, 205, 156, 133, 236, 36, 92, 22, 31, 152, 112, 101, 11, 48, 4, 22, 205, 67, 202, 180, 137, 63, 36, 12, 179, 171, 163, 190, 112, 99, 105, 234, 8, 168, 159, 172, 61, 112, 24, 175, 150, 205, 32, 26, 101, 244, 187, 14, 67, 148, 76, 28, 187, 51, 243},
	Name:      "test-user-1",
}
var mockPayloadUpdate2 = users_txpayload.PayloadUpdate{
	NewPubKey: PubKeyUser1,
	PubKey:    PubKeyUser2,
	Sig:       [96]byte{152, 233, 212, 98, 156, 71, 18, 224, 33, 61, 149, 185, 50, 123, 99, 132, 191, 101, 96, 132, 174, 61, 70, 164, 88, 180, 107, 105, 153, 205, 90, 90, 3, 189, 105, 229, 96, 83, 0, 220, 230, 178, 21, 164, 162, 250, 198, 17, 9, 12, 8, 152, 251, 188, 163, 194, 184, 65, 53, 194, 15, 202, 211, 172, 250, 49, 155, 229, 24, 178, 212, 28, 94, 108, 163, 108, 225, 102, 89, 156, 4, 226, 155, 111, 180, 212, 176, 233, 8, 162, 222, 209, 253, 162, 92, 118},
	Name:      "test-user-2",
}
var mockPayloadUpdate3 = users_txpayload.PayloadUpdate{
	NewPubKey: PubKeyUser1,
	PubKey:    PubKeyUser3,
	Sig:       [96]byte{152, 142, 24, 132, 32, 159, 157, 124, 53, 116, 112, 238, 152, 111, 128, 166, 66, 99, 104, 16, 66, 231, 185, 242, 97, 185, 210, 161, 229, 199, 113, 233, 138, 219, 217, 55, 215, 221, 229, 120, 177, 221, 116, 85, 222, 107, 141, 6, 11, 20, 191, 18, 235, 117, 108, 217, 255, 89, 208, 251, 229, 33, 14, 214, 229, 159, 73, 219, 234, 235, 252, 85, 134, 202, 59, 229, 100, 137, 130, 54, 16, 154, 205, 134, 101, 126, 247, 31, 242, 44, 218, 189, 147, 130, 16, 112},
	Name:      "test-user-3",
}
var mockPayloadUpdate4 = users_txpayload.PayloadUpdate{
	NewPubKey: PubKeyUser1,
	PubKey:    PubKeyUser4,
	Sig:       [96]byte{129, 156, 42, 10, 254, 233, 176, 53, 31, 89, 52, 13, 234, 170, 197, 66, 216, 54, 128, 34, 231, 237, 76, 66, 172, 107, 213, 220, 151, 148, 214, 243, 228, 77, 172, 75, 21, 115, 64, 144, 136, 211, 139, 188, 149, 187, 244, 244, 5, 186, 93, 229, 120, 69, 25, 207, 79, 129, 198, 4, 35, 50, 154, 165, 35, 70, 246, 85, 87, 190, 74, 239, 20, 208, 213, 63, 75, 81, 167, 216, 68, 101, 163, 171, 86, 116, 198, 85, 55, 155, 254, 142, 222, 163, 110, 176},
	Name:      "test-user-4",
}
var mockPayloadUpdate5 = users_txpayload.PayloadUpdate{
	NewPubKey: PubKeyUser1,
	PubKey:    PubKeyUser5,
	Sig:       [96]byte{163, 129, 161, 213, 213, 149, 154, 169, 43, 9, 122, 185, 190, 117, 183, 112, 225, 228, 50, 239, 246, 170, 116, 88, 18, 228, 3, 123, 34, 207, 206, 5, 90, 213, 53, 128, 178, 17, 64, 88, 231, 103, 104, 117, 104, 171, 71, 55, 7, 75, 31, 10, 26, 227, 178, 99, 30, 139, 244, 239, 49, 53, 31, 201, 178, 91, 127, 237, 131, 80, 113, 90, 122, 192, 75, 100, 151, 143, 71, 172, 9, 211, 104, 90, 181, 191, 105, 236, 0, 159, 175, 139, 194, 139, 161, 33},
	Name:      "test-user-5",
}

func TestMockUpdate(t *testing.T) {
	buf := bytes.NewBuffer([]byte{})
	err := mockPayloadUpdate1.Serialize(buf)
	if err != nil {
		t.Fatal("unable to serialize")
	}
	err = user.MatchVerify(buf.Bytes(), p2p.Update)
	if err != nil {
		t.Fatal("verification failed")
	}
	err = user.SigVerify(buf.Bytes(), p2p.Update)
	if err != nil {
		t.Fatal("verification failed")
	}
}

var mockPayloadUpdateBatch = []users_txpayload.PayloadUpdate{mockPayloadUpdate1, mockPayloadUpdate2, mockPayloadUpdate3, mockPayloadUpdate4, mockPayloadUpdate5}

func TestMockUpdateBatch(t *testing.T) {
	var payload [][]byte
	for _, singlePayload := range mockPayloadUpdateBatch {
		buf := bytes.NewBuffer([]byte{})
		err := singlePayload.Serialize(buf)
		if err != nil {
			t.Fatal("unable to serialize")
		}
		payload = append(payload, buf.Bytes())
	}
	err := user.MatchVerifyBatch(payload, p2p.Update)
	if err != nil {
		t.Fatal("verification failed")
	}
	err = user.SigVerifyBatch(payload, p2p.Update)
	if err != nil {
		t.Fatal("verification failed")
	}
}

var mockPayloadRevoke1 = users_txpayload.PayloadRevoke{
	Sig:  [96]byte{147, 117, 13, 219, 3, 243, 195, 62, 190, 143, 76, 19, 159, 101, 247, 178, 55, 153, 88, 1, 150, 231, 126, 179, 174, 79, 43, 42, 127, 159, 37, 188, 159, 99, 208, 57, 156, 145, 200, 240, 201, 70, 91, 239, 232, 227, 123, 97, 9, 196, 240, 254, 94, 3, 174, 242, 105, 120, 152, 239, 22, 168, 133, 106, 187, 109, 111, 28, 254, 12, 27, 126, 181, 97, 33, 226, 136, 65, 10, 34, 103, 156, 157, 69, 26, 17, 160, 219, 95, 11, 74, 100, 152, 205, 68, 94},
	Name: "test-user-1",
}
var mockPayloadRevoke2 = users_txpayload.PayloadRevoke{
	Sig:  [96]byte{149, 8, 95, 247, 160, 56, 214, 63, 217, 17, 102, 27, 202, 74, 177, 251, 211, 78, 182, 222, 83, 155, 199, 64, 116, 23, 90, 197, 105, 236, 135, 75, 190, 202, 222, 141, 193, 62, 238, 30, 72, 35, 54, 164, 5, 188, 192, 182, 2, 88, 94, 95, 213, 219, 43, 6, 111, 32, 94, 193, 83, 42, 236, 122, 220, 186, 169, 9, 26, 0, 29, 249, 216, 17, 178, 1, 248, 66, 81, 164, 2, 239, 28, 206, 244, 192, 218, 92, 140, 129, 1, 196, 183, 243, 143, 202},
	Name: "test-user-2",
}
var mockPayloadRevoke3 = users_txpayload.PayloadRevoke{
	Sig:  [96]byte{168, 34, 149, 141, 216, 155, 18, 5, 218, 254, 77, 144, 18, 100, 181, 13, 251, 207, 235, 198, 104, 173, 36, 59, 72, 196, 100, 79, 145, 169, 186, 146, 25, 167, 87, 54, 110, 174, 120, 94, 59, 118, 144, 135, 43, 42, 205, 22, 16, 191, 21, 160, 8, 121, 52, 38, 162, 158, 22, 156, 166, 38, 77, 240, 160, 1, 12, 186, 73, 60, 212, 207, 148, 101, 90, 146, 144, 144, 227, 49, 11, 145, 149, 185, 238, 248, 152, 135, 242, 222, 0, 47, 166, 46, 251, 17},
	Name: "test-user-3",
}
var mockPayloadRevoke4 = users_txpayload.PayloadRevoke{
	Sig:  [96]byte{184, 208, 8, 181, 189, 161, 163, 148, 46, 54, 176, 163, 5, 126, 20, 221, 112, 146, 9, 255, 241, 39, 192, 78, 78, 92, 96, 188, 89, 190, 232, 214, 179, 11, 234, 231, 90, 8, 151, 104, 2, 7, 92, 144, 199, 118, 191, 78, 12, 181, 38, 94, 247, 25, 166, 228, 86, 26, 49, 234, 91, 177, 45, 146, 111, 139, 60, 176, 17, 128, 227, 119, 98, 152, 0, 130, 37, 148, 139, 141, 4, 178, 210, 58, 23, 235, 25, 232, 102, 92, 9, 218, 131, 122, 253, 39},
	Name: "test-user-4",
}
var mockPayloadRevoke5 = users_txpayload.PayloadRevoke{
	Sig:  [96]byte{147, 69, 10, 170, 61, 110, 192, 89, 24, 213, 147, 248, 86, 59, 106, 216, 50, 248, 230, 128, 1, 89, 135, 237, 152, 140, 254, 250, 175, 28, 211, 134, 145, 103, 49, 43, 90, 44, 127, 245, 146, 54, 93, 71, 137, 194, 86, 5, 4, 173, 88, 63, 219, 78, 163, 206, 2, 93, 67, 0, 70, 59, 181, 254, 73, 212, 179, 170, 45, 109, 44, 36, 190, 234, 14, 208, 109, 237, 97, 23, 48, 91, 158, 212, 102, 214, 248, 178, 155, 235, 193, 121, 29, 134, 70, 36},
	Name: "test-user-5",
}

var mockPayloadRevokeBatch = []users_txpayload.PayloadRevoke{mockPayloadRevoke1, mockPayloadRevoke2, mockPayloadRevoke3, mockPayloadRevoke4, mockPayloadRevoke5}

func TestMockRevoke(t *testing.T) {
	buf := bytes.NewBuffer([]byte{})
	err := mockPayloadRevoke1.Serialize(buf)
	if err != nil {
		t.Fatal("unable to serialize")
	}
	err = user.MatchVerify(buf.Bytes(), p2p.Revoke)
	if err != nil {
		t.Fatal("verification failed")
	}
	err = user.SigVerify(buf.Bytes(), p2p.Revoke)
	if err != nil {
		t.Fatal("verification failed")
	}
}

func TestMockRevokeBatch(t *testing.T) {
	var payload [][]byte
	for _, singlePayload := range mockPayloadRevokeBatch {
		buf := bytes.NewBuffer([]byte{})
		err := singlePayload.Serialize(buf)
		if err != nil {
			t.Fatal("unable to serialize")
		}
		payload = append(payload, buf.Bytes())
	}
	err := user.MatchVerifyBatch(payload, p2p.Revoke)
	if err != nil {
		t.Fatal("verification failed")
	}
	err = user.SigVerifyBatch(payload, p2p.Revoke)
	if err != nil {
		t.Fatal("verification failed")
	}
}
