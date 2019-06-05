package config

//This package contains the example of the standard configuration file

//ConfigurationString is the base configuration string, written in case a log file is not found
var ConfigurationString = `
Logfile = "/var/log/granti.log"
DatabasePath = "/var/granti.db"
LogExistTimeout = "1s"

[[jail]]
Name = "myjail"
Enabled = true
Regex = "^(?P<IP>(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)) (?P<timestamp>\d{10}) .*$"
RegexBlacklist = ["curl","zmap"]
BlacklistBanCommand = "/home/user/banBotIP.sh <IP>"
IPWhitelist = ["127.0.0.1/8", "10.0.0.0/8", "172.16.0.0/16", "192.168.0.0/16"]
IPGroupName = "IP"
TsGroupName = "timestamp"
LogFile = "/var/log/log.log"
CounterMaxValue = 100
FindTime = "60m"
Burst = 0
TsLayout = "Mon Jan 2 15:04:05 -0700 MST 2006"
BanAction = "/home/user/banip.sh <IP>"
`
