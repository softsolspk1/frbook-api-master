package models

import (
	"errors"
	"strconv"
	"sync"
	"time"
)

const (
	Epoch    int64 = 1580794200
	nodeBits uint8 = 5
	stepBits uint8 = 16
	// DEPRECATED: the below four variables will be removed in a future release.
	nodeMax   int64 = -1 ^ (-1 << nodeBits)
	nodeMask        = nodeMax << stepBits
	stepMask  int64 = -1 ^ (-1 << stepBits)
	timeShift       = nodeBits + stepBits
	nodeShift       = stepBits
)

// ErrInvalidBase58 is returned by ParseBase58 when given an invalid []byte
var ErrInvalidBase58 = errors.New("invalid base58")

// ErrInvalidBase32 is returned by ParseBase32 when given an invalid []byte
var ErrInvalidBase32 = errors.New("invalid base32")

// Create maps for decoding Base58/Base32.
// This speeds up the process tremendously.

// A Node struct holds the basic information needed for a snowflake generator
// node
type IDNode struct {
	mu    sync.Mutex
	epoch time.Time
	time  int64
	node  int64
	step  int64

	// nodeMax   int64
	// nodeMask  int64
	// stepMask  int64
	// timeShift uint8
	// nodeShift uint8
}

// An ID is a custom type used for a snowflake ID.  This is used so we can
// attach methods onto the ID.
type ID int64

// NewIDNode returns a new snowflake node that can be used to generate snowflake
// IDs
func NewIDNode(node int64) (*IDNode, error) {

	// re-calc in case custom nodeBits or stepBits were set
	// DEPRECATED: the below block will be removed in a future release.

	n := IDNode{}
	n.node = node
	// n.nodeMax = -1 ^ (-1 << nodeBits)
	// n.nodeMask = n.nodeMax << stepBits
	// n.stepMask = -1 ^ (-1 << stepBits)
	// n.timeShift = nodeBits + stepBits
	// n.nodeShift = stepBits

	if n.node < 0 || n.node > nodeMax {
		return nil, errors.New("Node number must be between 0 and " + strconv.FormatInt(nodeMax, 10))
	}

	var curTime = time.Now()
	// add time.Duration to curTime to make sure we use the monotonic clock if available
	n.epoch = curTime.Add(time.Unix(Epoch, 0).Sub(curTime))

	return &n, nil
}

// Generate creates and returns a unique snowflake ID
// To help guarantee uniqueness
// - Make sure your system is keeping accurate system time
// - Make sure you never have multiple nodes running with the same node ID
func (n *IDNode) Generate() ID {
	n.mu.Lock()
	now := time.Since(n.epoch).Milliseconds() / 1000

	if now == n.time {
		n.step = (n.step + 1) & stepMask

		if n.step == 0 {
			for now <= n.time {
				now = time.Since(n.epoch).Milliseconds() / 1000
			}
		}
	} else {
		n.step = 0
	}

	n.time = now

	r := ID((now)<<timeShift |
		(n.node << nodeShift) |
		(n.step),
	)

	n.mu.Unlock()
	return r
}

// Int64 returns an int64 of the snowflake ID
func (f ID) Int64() int64 {
	return int64(f)
}

// ParseInt64 converts an int64 into a snowflake ID
func ParseInt64(id int64) ID {
	return ID(id)
}

// String returns a string of the snowflake ID
func (f ID) String() string {
	return strconv.FormatInt(int64(f), 10)
}

// ParseString converts a string into a snowflake ID
func ParseString(id string) (ID, error) {
	i, err := strconv.ParseInt(id, 10, 64)
	return ID(i), err

}

// Time returns an int64 unix timestamp in milliseconds of the snowflake ID time
// DEPRECATED: the below function will be removed in a future release.
func (f ID) Time() int64 {
	return (int64(f) >> timeShift) + Epoch
}

// IDNode returns an int64 of the snowflake ID node number
// DEPRECATED: the below function will be removed in a future release.
func (f ID) Node() int64 {
	return int64(f) & nodeMask >> nodeShift
}

// Step returns an int64 of the snowflake step (or sequence) number
// DEPRECATED: the below function will be removed in a future release.
func (f ID) Step() int64 {
	return int64(f) & stepMask
}
