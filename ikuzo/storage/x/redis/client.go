package redis

type Config struct {
	Address  string
	Password string
	UserName string
	Database int
	Sentinel struct {
		UserName   string
		Address    string
		Password   string
		MasterName string
	}
}
