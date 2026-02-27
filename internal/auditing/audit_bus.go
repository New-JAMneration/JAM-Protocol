package auditing

import (
	"log"

	"github.com/New-JAMneration/JAM-Protocol/internal/types"
)

// CE144Announcement represents a decoded CE144 audit announcement received from
// another validator. The CE handler goroutine constructs this and pushes it into
// AuditMessageBus.announcementCh.
type CE144Announcement struct {
	HeaderHash     types.OpaqueHash
	Tranche        types.U8
	ValidatorIndex types.ValidatorIndex
	WorkReports    []types.WorkPackageHash // which reports this validator announced it will audit
	Signature      types.Ed25519Signature
}

// CE145Judgment represents a decoded CE145 judgment publication received from
// another validator. The CE handler goroutine constructs this and pushes it into
// AuditMessageBus.judgmentCh.
type CE145Judgment struct {
	WorkReportHash types.WorkPackageHash
	ValidatorIndex types.ValidatorIndex
	IsValid        bool
	Signature      types.Ed25519Signature
}

const defaultBusBufferSize = 2048

// AuditMessageBus is the bridge between CE recv handlers (networking layer)
// and the audit tranche loop. CE handlers push messages in; the audit loop
// drains them each tranche via SyncAssignmentMapFromBus / SyncPositiveJudgersFromBus.
//
// Both channels are buffered. If a channel is full the message is dropped
// with a warning — this should only happen under extreme load.
type AuditMessageBus struct {
	announcementCh chan CE144Announcement
	judgmentCh     chan CE145Judgment
}

func NewAuditMessageBus() *AuditMessageBus {
	return NewAuditMessageBusWithSize(defaultBusBufferSize)
}

func NewAuditMessageBusWithSize(bufferSize int) *AuditMessageBus {
	return &AuditMessageBus{
		announcementCh: make(chan CE144Announcement, bufferSize),
		judgmentCh:     make(chan CE145Judgment, bufferSize),
	}
}

// OnAuditAnnouncementReceived is called by the CE144 recv handler goroutine.
// Non-blocking: drops the message if the channel is full.
func (bus *AuditMessageBus) OnAuditAnnouncementReceived(msg CE144Announcement) {
	select {
	case bus.announcementCh <- msg:
	default:
		log.Printf("[AuditMessageBus] CE144 announcement channel full, dropping message from validator %d", msg.ValidatorIndex)
	}
}

// OnJudgmentReceived is called by the CE145 recv handler goroutine.
// Non-blocking: drops the message if the channel is full.
func (bus *AuditMessageBus) OnJudgmentReceived(msg CE145Judgment) {
	select {
	case bus.judgmentCh <- msg:
	default:
		log.Printf("[AuditMessageBus] CE145 judgment channel full, dropping message from validator %d", msg.ValidatorIndex)
	}
}

// SyncAssignmentMapFromBus drains all pending CE144 announcements and merges
// them into assignmentMap. Called once per tranche to batch-process messages
// that arrived since the last drain.
func SyncAssignmentMapFromBus(
	bus *AuditMessageBus,
	assignmentMap map[types.WorkPackageHash][]types.ValidatorIndex,
) map[types.WorkPackageHash][]types.ValidatorIndex {
	for {
		select {
		case msg := <-bus.announcementCh:
			for _, wpHash := range msg.WorkReports {
				assignmentMap[wpHash] = append(assignmentMap[wpHash], msg.ValidatorIndex)
			}
		default:
			return assignmentMap
		}
	}
}

// SyncPositiveJudgersFromBus drains all pending CE145 judgments and merges
// valid (IsValid == true) entries into positiveJudgers. Called once per tranche.
func SyncPositiveJudgersFromBus(
	bus *AuditMessageBus,
	positiveJudgers map[types.WorkPackageHash]map[types.ValidatorIndex]bool,
) map[types.WorkPackageHash]map[types.ValidatorIndex]bool {
	for {
		select {
		case msg := <-bus.judgmentCh:
			if msg.IsValid {
				if positiveJudgers[msg.WorkReportHash] == nil {
					positiveJudgers[msg.WorkReportHash] = make(map[types.ValidatorIndex]bool)
				}
				positiveJudgers[msg.WorkReportHash][msg.ValidatorIndex] = true
			}
		default:
			return positiveJudgers
		}
	}
}
