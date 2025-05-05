package commonutils

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"math/big"
	"regexp"
	"strings"
	"sync/atomic"
	"time"

	"github.com/newrelic/nri-postgresql/src/collection"
)

var re = regexp.MustCompile(`'[^']*'|\d+|".*?"`)

func GetDatabaseListInString(dbMap collection.DatabaseList) string {
	if len(dbMap) == 0 {
		return ""
	}
	var quoted []string
	for n := range dbMap {
		quoted = append(quoted, fmt.Sprintf("'%s'", n))
	}
	return strings.Join(quoted, ",")
}

func AnonymizeQueryText(q string) string {
	return re.ReplaceAllString(q, "?")
}

var planCounter uint64

func GeneratePlanID() (string, error) {
	ctr := atomic.AddUint64(&planCounter, 1)
	rnd, err := rand.Int(rand.Reader, big.NewInt(RandomIntRange))
	if err != nil {
		return "", ErrUnExpectedError
	}
	ts := time.Now().UTC().Format(TimeFormat)
	hash := sha1.Sum([]byte(ts))
	return fmt.Sprintf("%06d-%06d-%s", rnd.Int64(), ctr%1_000_000, hex.EncodeToString(hash[:6])), nil
}
