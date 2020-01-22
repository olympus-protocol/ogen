package workers

import (
	"github.com/grupokindynos/ogen/p2p"
	"github.com/grupokindynos/ogen/utils/serializer"
	"io"
)

type Worker struct {
	WorkerID          p2p.OutPoint
	PubKey            [48]byte
	LastBlockAssigned int64
	NextBlockAssigned int64
	Score             int64
	Version           int64
	Protocol          int64
	IP                string
	PayeeAddress      string
}

func (wk *Worker) Serialize(w io.Writer) error {
	err := wk.WorkerID.Serialize(w)
	if err != nil {
		return err
	}
	err = serializer.WriteElements(w, wk.PubKey, wk.LastBlockAssigned, wk.NextBlockAssigned, wk.Score, wk.Version, wk.Protocol)
	if err != nil {
		return err
	}
	err = serializer.WriteVarString(w, wk.IP)
	if err != nil {
		return err
	}
	err = serializer.WriteVarString(w, wk.PayeeAddress)
	if err != nil {
		return err
	}
	return nil
}

func (wk *Worker) Deserialize(r io.Reader) error {
	err := wk.WorkerID.Deserialize(r)
	if err != nil {
		return err
	}
	err = serializer.ReadElements(r, &wk.PubKey, &wk.LastBlockAssigned, &wk.NextBlockAssigned, &wk.Score, &wk.Version, &wk.Protocol)
	if err != nil {
		return err
	}
	wk.IP, err = serializer.ReadVarString(r)
	if err != nil {
		return err
	}
	wk.PayeeAddress, err = serializer.ReadVarString(r)
	if err != nil {
		return err
	}
	return nil
}
