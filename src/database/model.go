package database

import (
	"github.com/GSH-LAN/Unwindia_common/src/go/matchservice"
	"time"
)

type DotlanStatusEvents []DotlanStatusEvent

func (d *DotlanStatusEvents) Contains(state DotlanStatusEvent) bool {
	for _, v := range *d {
		if v == state {
			return true
		}
	}
	return false
}

type DotlanStatusEvent string

const (
	CMS_CONTEST_NEW       DotlanStatusEvent = "CMS_CONTEST_NEW"
	CMS_CONTEST_READY_1   DotlanStatusEvent = "CMS_CONTEST_READY_1" // contest.ready_a > 0
	CMS_CONTEST_READY_2   DotlanStatusEvent = "CMS_CONTEST_READY_2" // contest.ready_b > 0
	CMS_CONTEST_READY_ALL DotlanStatusEvent = "CMS_CONTEST_READY_ALL"
	CMS_CONTEST_FINISHED  DotlanStatusEvent = "CMS_CONTEST_FINISHED" // contest.won > 0
)

type Result struct {
	Result DotlanStatusList
	Error  error
}

type DotlanStatusList []DotlanStatus

// GetElementById returns the element with the given id
func (d *DotlanStatusList) GetElementById(id uint) *DotlanStatus {
	for _, element := range *d {
		if element.DotlanContestID == id {
			return &element
		}
	}
	return nil
}

type DotlanStatus struct {
	DotlanContestID uint                   `bson:"_id" json:"dotlanContestId"`
	Events          DotlanStatusEvents     `bson:"events,omitempty"`
	MatchInfo       matchservice.MatchInfo `bson:"matchInfo,omitempty"`
	CreatedAt       time.Time              `bson:"createdAt,omitempty"`
	UpdatedAt       time.Time              `bson:"updatedAt,omitempty"`
}

// DotlanMatchChanges is a list of changes to a dotlan match which needs to be published to the message queue.
type DotlanMatchChanges struct {
	DotlanStatusList
}
