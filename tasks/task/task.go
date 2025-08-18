package task

import "github.com/ecpartan/soap-server-tr069/utils"

type TaskRequestType int

const (
	NoTask TaskRequestType = iota
	GetParameterValues
	SetParameterValues
	AddObject
	DeleteObject
	GetParameterNames
	SetParameterAttributes
	GetParameterAttributes
	GetRPCMethods
	Download
	Upload
	FactoryReset
	Reboot
)

type Task struct {
	ID        utils.ID
	Action    TaskRequestType
	Params    any
	EventCode int
	Once      bool
}
