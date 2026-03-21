package ce

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"sync"

	"github.com/New-JAMneration/JAM-Protocol/internal/blockchain"
	"github.com/New-JAMneration/JAM-Protocol/internal/database"
	"github.com/New-JAMneration/JAM-Protocol/internal/types"
	"github.com/New-JAMneration/JAM-Protocol/internal/utilities/hash"
)

var (
	ceDB   database.Database
	ceDBMu sync.RWMutex
)

// SetDatabase sets the database for CE storage (tests). Production uses bc.Database().
func SetDatabase(db database.Database) {
	ceDBMu.Lock()
	defer ceDBMu.Unlock()
	ceDB = db
}

// GetDatabase returns the package-level db set by SetDatabase (tests).
func GetDatabase() database.Database {
	ceDBMu.RLock()
	defer ceDBMu.RUnlock()
	return ceDB
}

// DB returns the database for CE storage: bc.Database() when non-nil, else GetDatabase() (tests).
func DB(bc blockchain.Blockchain) database.Database {
	if bc != nil {
		if d := bc.Database(); d != nil {
			return d
		}
	}
	return GetDatabase()
}

// --- Key prefixes (CE namespace, avoid collision with store)
var (
	ceJustificationPrefix = []byte("ce/j/")
	ceAssurancePrefix     = []byte("ce/a/")
	cePreimagePrefix      = []byte("ce/p/")
	cePreimageAnnPrefix   = []byte("ce/pa/")
	ceAuditAnnPrefix      = []byte("ce/aa/")
	ceJudgmentPrefix      = []byte("ce/jg/")
	ceSetPrefix           = []byte("ce/s/")
	ceSetSep              = byte(0)
	ceWpBundlePrefix      = []byte("ce/wp_bundle/")
)

func ceJustificationKey(erasureRoot []byte, shardIndex uint32) []byte {
	k := make([]byte, 0, len(ceJustificationPrefix)+hex.EncodedLen(len(erasureRoot))+1+4)
	k = append(k, ceJustificationPrefix...)
	k = append(k, hex.EncodeToString(erasureRoot)...)
	k = append(k, ':')
	k = append(k, strconv.FormatUint(uint64(shardIndex), 10)...)
	return k
}

func ceAssuranceKey(headerHash types.HeaderHash) []byte {
	k := make([]byte, 0, len(ceAssurancePrefix)+32)
	k = append(k, ceAssurancePrefix...)
	k = append(k, headerHash[:]...)
	return k
}

func ceAssuranceSetKey(headerHash types.HeaderHash) string {
	return "availability_assurances_set:" + hex.EncodeToString(headerHash[:])
}

func cePreimageKey(h types.OpaqueHash) []byte {
	k := make([]byte, 0, len(cePreimagePrefix)+hex.EncodedLen(len(h)))
	k = append(k, cePreimagePrefix...)
	k = append(k, hex.EncodeToString(h[:])...)
	return k
}

func cePreimageAnnKey(h types.OpaqueHash) []byte {
	k := make([]byte, 0, len(cePreimageAnnPrefix)+hex.EncodedLen(len(h)))
	k = append(k, cePreimageAnnPrefix...)
	k = append(k, hex.EncodeToString(h[:])...)
	return k
}

func cePreimageAnnServiceSetKey(serviceID types.ServiceID) string {
	return "service_preimage_announcements:" + strconv.FormatUint(uint64(serviceID), 10)
}

func ceAuditAnnKey(headerHash types.OpaqueHash, tranche uint8) []byte {
	k := make([]byte, 0, len(ceAuditAnnPrefix)+hex.EncodedLen(len(headerHash))+1+2)
	k = append(k, ceAuditAnnPrefix...)
	k = append(k, hex.EncodeToString(headerHash[:])...)
	k = append(k, ':')
	k = append(k, strconv.FormatUint(uint64(tranche), 10)...)
	return k
}

func ceAuditAnnHeaderSetKey(headerHash types.OpaqueHash) string {
	return "header_audit_announcements:" + hex.EncodeToString(headerHash[:])
}

func ceJudgmentKey(workReportHash types.WorkReportHash, epochIndex types.U32, validatorIndex types.ValidatorIndex) []byte {
	k := make([]byte, 0, len(ceJudgmentPrefix)+64+1+4+1+2)
	k = append(k, ceJudgmentPrefix...)
	k = append(k, hex.EncodeToString(workReportHash[:])...)
	k = append(k, ':')
	k = append(k, strconv.FormatUint(uint64(epochIndex), 10)...)
	k = append(k, ':')
	k = append(k, strconv.FormatUint(uint64(validatorIndex), 10)...)
	return k
}

func ceJudgmentWorkReportSetKey(workReportHash types.WorkReportHash) string {
	return "work_report_judgments:" + hex.EncodeToString(workReportHash[:])
}

func ceJudgmentEpochSetKey(epochIndex types.U32) string {
	return "epoch_judgments:" + strconv.FormatUint(uint64(epochIndex), 10)
}

func ceJudgmentValidatorSetKey(validatorIndex types.ValidatorIndex) string {
	return "validator_judgments:" + strconv.FormatUint(uint64(validatorIndex), 10)
}

func wpBundleKey(erasureRoot []byte) []byte {
	k := make([]byte, 0, len(ceWpBundlePrefix)+hex.EncodedLen(len(erasureRoot)))
	k = append(k, ceWpBundlePrefix...)
	k = append(k, hex.EncodeToString(erasureRoot)...)
	return k
}

func setMemberKey(setKey string, memberHash []byte) []byte {
	k := make([]byte, 0, len(ceSetPrefix)+len(setKey)+1+len(memberHash))
	k = append(k, ceSetPrefix...)
	k = append(k, setKey...)
	k = append(k, ceSetSep)
	k = append(k, memberHash...)
	return k
}

func setMemberPrefix(setKey string) []byte {
	k := make([]byte, 0, len(ceSetPrefix)+len(setKey)+1)
	k = append(k, ceSetPrefix...)
	k = append(k, setKey...)
	k = append(k, ceSetSep)
	return k
}

// --- Justification (CE137/CE138)
func GetJustification(db database.Reader, erasureRoot []byte, shardIndex uint32) ([]byte, error) {
	if db == nil {
		return nil, fmt.Errorf("database not available")
	}
	data, found, err := db.Get(ceJustificationKey(erasureRoot, shardIndex))
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, nil
	}
	return data, nil
}

func PutJustification(db database.Writer, erasureRoot []byte, shardIndex uint32, data []byte) error {
	if db == nil {
		return fmt.Errorf("database not available")
	}
	return db.Put(ceJustificationKey(erasureRoot, shardIndex), data)
}

// --- Generic KV
func GetKV(db database.Reader, key []byte) ([]byte, error) {
	if db == nil {
		return nil, fmt.Errorf("database not available")
	}
	data, found, err := db.Get(key)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, nil
	}
	return data, nil
}

func PutKV(db database.Writer, key []byte, value []byte) error {
	if db == nil {
		return fmt.Errorf("database not available")
	}
	return db.Put(key, value)
}

// --- Set (prefix-based: ce/s/<setKey>\x00<hash(member)>)
func setMemberHash(member []byte) []byte {
	h := hash.Blake2bHash(types.ByteSequence(member))
	return h[:]
}

func SAdd(db database.Database, setKey string, member []byte) error {
	if db == nil {
		return fmt.Errorf("database not available")
	}
	return db.Put(setMemberKey(setKey, setMemberHash(member)), member)
}

func SMembers(db database.Iterable, setKey string) ([][]byte, error) {
	if db == nil {
		return nil, fmt.Errorf("database not available")
	}
	prefix := setMemberPrefix(setKey)
	iter, err := db.NewIterator(prefix, nil)
	if err != nil {
		return nil, err
	}
	defer iter.Close()
	var out [][]byte
	for iter.Next() {
		val := iter.Value()
		if len(val) > 0 {
			cp := make([]byte, len(val))
			copy(cp, val)
			out = append(out, cp)
		}
	}
	if err := iter.Error(); err != nil {
		return nil, err
	}
	return out, nil
}

func SIsMember(db database.Reader, setKey string, member []byte) (bool, error) {
	if db == nil {
		return false, fmt.Errorf("database not available")
	}
	_, found, err := db.Get(setMemberKey(setKey, setMemberHash(member)))
	return found, err
}
