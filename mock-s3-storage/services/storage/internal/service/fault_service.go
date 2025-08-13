package service

type FaultService interface {
	StartFault(name string) error
	StopFault(name string) error
	GetFaultStatus(name string) (string, error)
	ListFaults() ([]string, error)
}
