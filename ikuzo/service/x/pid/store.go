package pid

type Store interface {
	Get(id string) (pid PID, err error)
	Save(pid PID) (err error)
}
