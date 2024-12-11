package notification

import (
	"fmt"

	"github.com/spf13/viper"

	"story-monitor/types"
)

type Event interface {
	Name() string
	Message() string
	IsEmpty() bool
}

type exception struct {
	validators []*struct {
		blockHeight int64
		moniker     string
	}
}

type (
	ValisActiveException exception
	SyncException        exception
)

func (e *ValisActiveException) Name() string {
	return "Validator InActive Exception\n"
}
func (e *ValisActiveException) Message() string {
	var msg string
	if !e.IsEmpty() {
		msg = e.Name()
		for _, val := range e.validators {
			msg += fmt.Sprintf("%s validator is Inactive\n", val.moniker)
		}
	}
	return msg
}
func (e *ValisActiveException) IsEmpty() bool {
	return 0 == len(e.validators)
}

func (e *SyncException) Name() string {
	return "Sync Exception \n"
}
func (e *SyncException) Message() string {
	proportion := viper.GetFloat64("alert.proportion")
	var msg string
	if !e.IsEmpty() {
		msg = e.Name()
		for _, val := range e.validators {
			msg += fmt.Sprintf("The %s validator has not signed for 5 consecutive blocks or the last 100 blocks without signature rate reaches %f at block height of %d. \n",
				val.moniker, proportion, val.blockHeight)
		}
	}
	return msg
}

func (e *SyncException) IsEmpty() bool {
	return 0 == len(e.validators)
}

func ParseValisActiveException(valisActive []string) *ValisActiveException {
	if len(valisActive) == 0 {
		logger.Error("validator inActive is empty, please check")
		return nil
	}

	valisActiveException := &ValisActiveException{
		validators: make([]*struct {
			blockHeight int64
			moniker     string
		}, 0),
	}

	return valisActiveException
}

func ParseSyncException(missedSign []*types.ValSignMissed) *SyncException {
	if len(missedSign) == 0 {
		logger.Error("validator missed sign is empty, please check")
		return nil
	}

	syncException := &SyncException{
		validators: make([]*struct {
			blockHeight int64
			moniker     string
		}, 0),
	}
	for _, valMissedSign := range missedSign {
		syncException.validators = append(syncException.validators, &struct {
			blockHeight int64
			moniker     string
		}{moniker: valMissedSign.OperatorAddr,
			blockHeight: valMissedSign.BlockHeight})

	}
	return syncException
}
