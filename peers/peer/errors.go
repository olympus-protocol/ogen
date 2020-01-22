package peer

import "errors"

var (
	ErrorReadRemote           = errors.New("unable to read message from remote peer")
	ErrorWriteRemote          = errors.New("unable to write message to remote peer")
	ErrorNoVersionFirst       = errors.New("version message wasn't the first received")
	ErrorNoVerackAfterVersion = errors.New("verack message wasn't received after version")
	ErrorNoVersionAfterVerack = errors.New("version message wasn't received after verack")
	ErrorPingNonceMismatch    = errors.New("pong nonce doesn't match last ping")
)
