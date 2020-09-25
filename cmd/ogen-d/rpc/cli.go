package rpc

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/olympus-protocol/ogen/api/proto"
	"github.com/olympus-protocol/ogen/cmd/ogen/indexer"
	"github.com/olympus-protocol/ogen/pkg/primitives"
	"github.com/olympus-protocol/ogen/pkg/rpcclient"
	"github.com/spf13/viper"
	"io"
	"os"
	""
	"sync"
)



