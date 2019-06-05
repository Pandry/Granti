package config

//Config is the abstraction of the configuration file
type Config struct {
	Logfile         string
	DatabasePath    string
	LogExistTimeout string

	Jails []JailInfo
}

//JailInfo is the abstraction of a single jail
type JailInfo struct {
	Enabled         bool
	Name            string
	Regex           string
	IPGroupName     string
	TsGroupName     string
	LogFile         string
	CounterMaxValue uint
	FindTime        string
	Burst           uint
	TsLayout        string
	BanAction       string
}
