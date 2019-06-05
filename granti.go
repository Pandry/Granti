package main

import (
	"crypto/sha1"
	"database/sql"
	"encoding/base64"
	"flag"
	"fmt"
	"granti/config"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/hpcloud/tail"
	_ "github.com/mattn/go-sqlite3"
)

const (
	LogDebug = iota
	LogInfo
	LogWarn
	LogErr
	LogCrit
)

func main() {

	//Variables for the program
	var (
		verbose     bool
		veryVerbose bool
		confPath    string
	)

	l := func(logLevel int, jailName string, args ...interface{}) {
		if (logLevel > LogInfo) || (logLevel > LogDebug && verbose) || veryVerbose {
			var (
				prefix string
				jn     string
			)
			switch logLevel {
			case LogDebug:
				prefix = "[\033[34mDEBUG\033[39m]"
			case LogInfo:
				prefix = "[\033[32mINFO\033[39m]"
			case LogWarn:
				prefix = "[\033[93mWARN\033[39m]"
			case LogErr:
				prefix = "[\033[91mERROR\033[39m]"
			case LogCrit:
				prefix = "[\033[101mCRIT\033[49m]"
			}
			if jailName != "" {
				jn = " [" + jailName + "] "
			}
			//log.Println(prefix, jn, args)
			fmt.Fprintln(os.Stdout, prefix, jn, args)
		}
	}

	flag.BoolVar(&veryVerbose, "vv", false, "Enables very verbose output - Shows debug")
	flag.BoolVar(&verbose, "v", false, "Enables verbose output - Shows info")
	flag.StringVar(&confPath, "c", config.ConfigurationPath, "Point to a different configuration path from "+config.ConfigurationPath)
	flag.Parse()
	l(LogDebug, "", "Flags parsed")

	configFile, err := os.Open(confPath)
	l(LogDebug, "", "Opening configuration file")
	defer configFile.Close()
	if os.IsPermission(err) {
		// Config file is unaccessible
		l(LogCrit, "", "The configuration file is unaccessible; Permission issue.")
		os.Exit(1)
		//panic()
		//log.Panic(".")
	}
	if os.IsNotExist(err) {
		// Config file does not existsConfiguratiConfigurationPathConfigurationPathConfigurationPathConfigurationPathConfigurationPathConfigurationPathonPath
		l(LogInfo, "", "The configuration file does not exists.")
		l(LogInfo, "", "Trying to create a configuration file...")
		confFile, err := os.Create(confPath)
		if err == nil {
			defer confFile.Close()
			fmt.Fprintf(confFile, config.ConfigurationString)
			l(LogWarn, "", "The configuration file has been created.\nExiting to allow editing.")
			os.Exit(0)
			//log.Panic("The configuration file does not exist; One has been created...")
		} else {
			l(LogCrit, "", "The configuration was not created.\nImpossible to create one.")
			os.Exit(2)
			//log.Panic("The configuration file does not exist; Cannot create one...")
		}
	}
	if err != nil {
		//log.Panic("Unspecified error while getting the configuration file")
		l(LogCrit, "", "An error occourred while getting the configuration file.\n  Error: ", err.Error())
		os.Exit(3)
	}
	l(LogDebug, "", "Opening configuration file.")
	//Read the config file to a string
	configFileBytes, err := ioutil.ReadAll(configFile)
	if err != nil {
		//log.Panic("Cannot read the configuration file content; Aborting...")
		l(LogCrit, "", "An error occourred while reading the configuration file.\n  Error: ", err.Error())
		os.Exit(4)
	}
	configFileString := string(configFileBytes)
	l(LogDebug, "", "The configuration file was read correctly.")

	var conf config.Config
	//parse config file
	l(LogDebug, "", "Starting the parting of the configuration file.")
	if _, err := toml.Decode(configFileString, &conf); err != nil {
		// handle error
		l(LogCrit, "", "Cannot parse the configuration file!\n  Error:", err.Error())
		//log.Panic("Error while parsing the configuration file!\n  Error: ", err.Error())
	}
	l(LogDebug, "", "The configuration file was prsed correctly.")

	l(LogDebug, "", "Instantiating the database...")

	db, err := sql.Open("sqlite3", "file:"+conf.DatabasePath+"?")
	defer db.Close()

	if err != nil {
		l(LogCrit, "", "Error instantiating the database.\n  Error: "+err.Error())
		os.Exit(5)
		//log.Panic("Error opening the database; Aborting...\n  Error: " + err.Error())
	}
	l(LogDebug, "", "Database opened correctly.")

	l(LogDebug, "", "Seeding the database.")
	_, err = db.Exec(config.DBSchema)
	if err != nil {
		l(LogCrit, "", "Error executing the seed query.\n  Error: "+err.Error())
		os.Exit(6)
		//log.Panic("Error executing the initialization query")
	}
	l(LogDebug, "", "The database was seeded correctly and is now ready.")

	l(LogDebug, "", "Checking the jails in the configuration file.")
	for _, jail := range conf.Jails {
		l(LogDebug, jail.Name, "making sure the jail exists in the database.")
		_, err = db.Exec("INSERT OR IGNORE INTO Jails (Name) VALUES (?)", jail.Name)
		if err != nil {
			l(LogErr, jail.Name, "Error making sure the jail exists in the database.\n  Error: "+err.Error())
		}
	}
	l(LogDebug, "", "The jails was insterted in the database.")
	//TODO: Need to delete old ones

	l(LogDebug, "", "Creating map for maps durations.")
	jailDurations := make(map[string]time.Duration)

	//For each jail, create a new goroutine
	l(LogDebug, "", "Creating a routine for each jail...")
	for _, jail := range conf.Jails {

		//Fill the jailDurations hashmap with the duration for each jail
		jailDurations[jail.Name], err = time.ParseDuration(jail.FindTime)
		if err != nil {
			l(LogCrit, jail.Name, "Cannot parse the FindTime of the jail. Please, check for typos.")
			continue
		}

		l(LogDebug, jail.Name, "Creating a routine the jail...")
		//Iterates all the jails in the configuration file
		go func(jail config.JailInfo, conf config.Config, db *sql.DB) {
			l(LogDebug, jail.Name, "The routine was created successfully.")
			//get the local jail
			if !jail.Enabled {
				l(LogDebug, jail.Name, "The jail is not enabled. Shutting down the routine")
				return
			}

			var jailID int64
			err = db.QueryRow("SELECT ID FROM Jails WHERE Name=? LIMIT 1", jail.Name).Scan(&jailID)
			if err != nil {
				l(LogCrit, jail.Name, "Cannot get the \"", jail, "\" Jail ID. Returning.\n  Error: ", err.Error())
				return
			}

			l(LogDebug, jail.Name, "Starting checking the log file")
			//A while loop to be able to wait some seconds if the file does not exist
			for {
				l(LogDebug, jail.Name, "Checking if the log files exists")
				//See if the log file exists
				_, err := os.Stat(jail.LogFile)

				if os.IsPermission(err) || os.IsNotExist(err) {
					l(LogDebug, jail.Name, "The log file does not exists, or the permission was denied.\n  Error: ", err.Error())
					duration, dErr := time.ParseDuration(conf.LogExistTimeout)
					if dErr != nil {
						//log.Println("Cannot parse the LogExistTimeout variable. Defaulting to 1s.")
						l(LogErr, jail.Name, "Cannot parse the LogExistTimeout variable. Defaulting to 1s.")
						duration, _ = time.ParseDuration("1s")
					}
					if os.IsNotExist(err) {
						l(LogWarn, jail.Name, "The log file does not exist; Waiting ", duration, ".")
					} else if os.IsPermission(err) {
						l(LogWarn, jail.Name, "The log file is unaccessible (check permissions); Waiting ", duration, ".")
					}

					time.Sleep(duration)

				} else if err != nil {
					l(LogCrit, jail.Name, "Unspecified error while getting the configuration file. Returning.\n  Error: ", err.Error())
					return
				} else {
					l(LogDebug, jail.Name, "Log file found and accessible.")
					break
				}
			}

			l(LogDebug, jail.Name, "Attaching to the log file.")
			//Stars a tail process on the log file in the goroutine
			t, err := tail.TailFile(jail.LogFile, tail.Config{Follow: true, Poll: runtime.GOOS == "windows", ReOpen: true})
			if err != nil {
				l(LogCrit, jail.Name, "Couldn't read logs from the jail. Returning.\n  Error: "+err.Error())
				return
			}
			l(LogDebug, jail.Name, "Attached successfully to the log file.")
			//newFile indicates if the line that is being read is from the old file or not (has been parsed or not)
			newFile := false

			//lineNumber is used as a counter to know the actual line number
			var lineNumber uint64
			//oldLinesCounter is the value of the last line checked. It is in the DB and read only once
			var oldLinesCounter sql.NullInt64
			//For each line in the log file
			for line := range t.Lines {
				//Increment the line number
				lineNumber++
				l(LogInfo, jail.Name, "Reading line number", lineNumber, ".")

				defer func() {
					// recover from panic if one occured. Set err to nil otherwise.
					if r := recover(); r != nil {
						l(LogCrit, jail.Name, "An error occourred.", "Recovered.\n  Error: ", r)
					}
				}()

				//Every n lines, check if the file is new(?)

				//If the process has just started, check the hash of the first line to check if the file has changed
				if lineNumber == 1 {
					l(LogDebug, jail.Name, "The current line is the first line.")
					l(LogDebug, jail.Name, "Getting the hash file of the first line")
					//Get the hash of the line
					hash := SHAHash(line.Text)

					//Get the old hash from the database
					var oldHash sql.NullString
					//TODO: Check for error

					l(LogDebug, jail.Name, "Extracting the old hash of the first line fro mthe database")
					dbErr := db.QueryRow("SELECT FirstLineHash FROM Jails WHERE ID=? LIMIT 1", jailID).Scan(&oldHash)

					if dbErr != nil {
						l(LogCrit, jail.Name, "Error with the database query while reading the old hash. Returning.\n  Error:", dbErr.Error())
						return
					}
					//Check if they're equal
					if hash != oldHash.String {
						//If the old file is different, update the hash and set the boolean flag
						//TODO: Check for error
						l(LogDebug, jail.Name, "The new log file is different from the previous one. Updating the hash in DB")
						_, dbErr = db.Exec("UPDATE Jails SET FirstLineHash=?, LastScannedLine=0 WHERE ID=?", hash, jailID)
						if dbErr != nil {
							l(LogCrit, jail.Name, "Error with the database query while updating the hash of the first line. Returning.\n  Error:", dbErr.Error())
							return
						}

						newFile = true
					}

					//Get the last scanned file from the database
					//	while this line number is higher than the current line number, the file is considered old and every line is ignored
					//TODO: Check for error
					l(LogDebug, jail.Name, "Getting the last scanned line from the database.")

					dbErr = db.QueryRow("SELECT LastScannedLine FROM Jails WHERE ID=? LIMIT 1", jailID).Scan(&oldLinesCounter)

					if dbErr != nil {
						l(LogCrit, jail.Name, "Error with the database query for getting the lines scanned in the previous instance.\n  Error:", dbErr.Error())
						return
					}
					l(LogDebug, jail.Name, "The last line parsed from the previous instance was the line n.", oldLinesCounter, ".")
				}

				//Check if the file is old (line was already read)
				if !newFile {
					//If the line number in the DB is inferior or equal to the current line number
					if uint64(oldLinesCounter.Int64) <= lineNumber {
						//Then the file we're reading hasn't been processed, yet
						newFile = true
						l(LogDebug, jail.Name, "The current line was not parsed. Starting parsing from this one.")
					} else {
						l(LogDebug, jail.Name, "The line has already been processed. Continuing.")
						continue
					}
				}

				if newFile {
					//The row hasn't been processed, yet

					var IP string
					var timestamp time.Time
					//Parsing of the line
					//Create the regex parsing
					re := regexp.MustCompile(jail.Regex)
					//Reading all the groups
					groupNames := re.SubexpNames()
					//For each match in the log (eg. ip, timestamp)
					l(LogDebug, jail.Name, "Parsing the regex of the string")
					for _, match := range re.FindAllStringSubmatch(line.Text, 1) {
						//For each group
						for groupIdx, group := range match {
							//Get the group name of the match
							name := groupNames[groupIdx]
							//IF the group name is the IP field, assign it to the IP filed
							if name == jail.IPGroupName {

								l(LogDebug, jail.Name, "The IP of the line is ", group)
								IP = group
								//If it's the TS group, assign it to the TS
							} else if name == jail.TsGroupName {
								//Parse the timestamp
								ts, err := time.Parse(jail.TsLayout, group)
								l(LogDebug, jail.Name, "The timestamp should be ", group, ". Trying to parse.")
								if err != nil {
									l(LogCrit, jail.Name, "Cannot parse the timestamp of the line ", lineNumber, "\n  Error:"+err.Error())
									//increments number of lines in the DB
									db.Exec("UPDATE Jails SET LastScannedLine=? WHERE ID=?", lineNumber, jailID)
									continue
									//panic("Cannot parse the timestamp of the line")
								} else {
									//Assign the timestamp to the req
									l(LogDebug, jail.Name, "The timestamp was parsed succesfully.")
									timestamp = ts
								}
							}
						}
					}

					//If the IP or the timestamp are empty, something wrong happened
					if IP == "" {
						l(LogCrit, jail.Name, "Couldn't been able to fetch the IP from the log on the line", lineNumber)
						//panic("Couldn't been able to fetch the IP or the timestamp")
						//Ignore but increment the line number in the database
						db.Exec("UPDATE Jails SET LastScannedLine=? WHERE ID=?", lineNumber, jailID)
						continue
					} else if timestamp.Unix() == 0 {
						l(LogCrit, jail.Name, "Couldn't been able to fetch the timestamp from the log on the line", lineNumber)
						db.Exec("UPDATE Jails SET LastScannedLine=? WHERE ID=?", lineNumber, jailID)
						//panic("Couldn't been able to fetch the IP or the timestamp")
						continue
					}

					//Makes sure the IP exists in the database
					//TODO: Check for error
					l(LogDebug, jail.Name, "Making sure the IP is in the database.")
					_, dbErr := db.Exec("INSERT OR IGNORE INTO IPsCounter(Jail, IP, Counter) VALUES (?,?,0)", jailID, IP)

					if dbErr != nil {
						l(LogCrit, jail.Name, "Cannot make sure the IP is in the database.\n  Error:", dbErr.Error())
					}

					var counterValue sql.NullInt64
					//Get the value of the counter in the database
					//TODO: Check for error

					l(LogDebug, jail.Name, "Getting the number of requests made from the IP.")
					dbErr = db.QueryRow("SELECT Counter FROM IPsCounter WHERE Jail=? AND IP=?", jailID, IP).Scan(&counterValue)

					if dbErr != nil {
						l(LogCrit, jail.Name, "Error with the database query. Returning.\n  Error:", dbErr.Error())
						return
					}
					//Increment the counter value
					counter := uint64(counterValue.Int64) + 1
					//Update the value in the database

					l(LogDebug, jail.Name, "Updating the request number of the IP.")
					_, dbErr = db.Exec("UPDATE IPsCounter SET Counter=? WHERE Jail=? AND IP=?", counter, jailID, IP)

					if dbErr != nil {
						l(LogCrit, jail.Name, "Error with the database query.\n  Error:", dbErr.Error())
						return
					}

					//enteredBurst tells if the user was going to be banned but was in a burst moment
					enteredBurst := false
					//burst is the value of the burst itself - eg 10 means that 10 requests above the limit was made without banning the user
					var burst uint
					//If the IP has made more than N requests, the "ring" closes up and we need to start checking the requests
					if counter > uint64(jail.CounterMaxValue) {
						//This timestamp is the one present in the database
						var timestampToBeOverwritten time.Time
						//The tmp var is the raw value from the database that needs to be parsed
						var tmpTimestamp sql.NullInt64
						//We get the log from the database
						//  The request number is obtaned by doung the module from the counter (eg. 6021) and the number of the ring elements (eg. 5000), then checking the result (1021)
						//TODO: Check for error

						dbErr = db.QueryRow("SELECT Timestamp, Burst FROM Logs WHERE IP=? AND Jail=? AND RequestNumber=?", IP, jailID, counter%uint64(jail.CounterMaxValue)).Scan(&tmpTimestamp, &burst)

						if dbErr != nil {
							l(LogCrit, jail.Name, "Error while selecting the timestamp from the logs. Line number: ", lineNumber, ". Trying to fix this...\n  Error:", dbErr.Error())

							//Trying to fix the issue manually...
							db.Exec("INSERT INTO Logs(Jail, IP, RequestNumber, Timestamp, Burst) VALUES (?,?,?,?,?) "+
								"ON CONFLICT(Jail, IP, RequestNumber) DO UPDATE SET RequestNumber = Excluded.RequestNumber, Burst = Excluded.Burst", jailID, IP, counter%uint64(jail.CounterMaxValue), timestamp.Unix(), burst)

							//Increment the line number
							db.Exec("UPDATE Jails SET LastScannedLine=? WHERE ID=?", lineNumber, jailID)
							continue

							//return
						}
						//Then we parse the timestamp
						timestampToBeOverwritten = time.Unix(tmpTimestamp.Int64, 0)

						//We parse the findtime string from the config
						//TODO: this should not be done for every request, but only once - PERFORMANCE IMPACT
						/*
							findTime, err := time.ParseDuration(jail.FindTime)
							if err != nil {
								panic("Cannot parse the FindTime duration in the configuration file")
							}*/

						//Then we calculate the delta value between the old timestamp and the currest request we are elaborating.
						//  If the delta is less then the find time, we came in the old ring status in a timeframe too short
						if timestamp.Sub(timestampToBeOverwritten).Seconds() < jailDurations[jail.Name].Seconds() {
							if burst < jail.Burst {
								enteredBurst = true
								l(LogWarn, jail.Name, "The IP ", IP, " made some request and gone above the treshold, burst-catched.")
							} else {

								err := db.QueryRow("SELECT 1 FROM Bans WHERE IP=? AND Jail=?", IP, jailID).Scan()
								if err != sql.ErrNoRows {
									l(LogWarn, jail.Name, "The IP ", IP, " should be banned but is already banned. Ignoring...")
								} else {

									//IP da bannare
									l(LogWarn, jail.Name, "The IP ", IP, " made too many requests, banning...")

									_, dbErr = db.Exec("INSERT INTO Bans(Jail, IP) VALUES (?,?)", jailID, IP)

									argstr := []string{"-c", strings.Replace(
										strings.Replace(
											strings.Replace(
												jail.BanAction, "<"+jail.IPGroupName+">", IP, -1),
											"<"+jail.TsGroupName+">", timestamp.String(), -1),
										"<line>", line.Text, -1)}
									out, err := exec.Command("/bin/bash", argstr...).Output()
									if err != nil {
										//log.Panic("Execution of the command failed.\n  STDOUT:\n" + string(out))
										log.Panic("Execution of the command failed.\n  STDOUT:\n" + string(out))
										//panic("Execution of the command failed.\n  STDOUT:\n" + string(out))

									}
								}
							}
						}
					}
					if enteredBurst {
						burst++
					} else {
						burst = 0
					}

					_, dbErr = db.Exec("INSERT INTO Logs(Jail, IP, RequestNumber, Timestamp, Burst) VALUES (?,?,?,?,?) "+
						"ON CONFLICT(Jail, IP, RequestNumber) DO UPDATE SET RequestNumber = Excluded.RequestNumber, Burst = Excluded.Burst", jailID, IP, counter%uint64(jail.CounterMaxValue), timestamp.Unix(), burst)

					if dbErr != nil {
						l(LogCrit, jail.Name, "Error with the database query.\n  Error:", dbErr.Error())
					}
					/*
						db.Exec("INSERT INTO IPsCounter(Jail, IP, Counter) VALUES (?,?,0) " +
						"ON CONFLICT(Jail, IP) DO UPDATE SET Counter = Excluded.MessageCount", lineNumber, jail.Name)*/

					//Update the last line read in the log file in the database

					_, dbErr = db.Exec("UPDATE Jails SET LastScannedLine=? WHERE ID=?", lineNumber, jailID)

					if dbErr != nil {
						l(LogCrit, jail.Name, "Error with the database query.\n  Error:", dbErr.Error())
					}
				}
			}
		}(jail, conf, db)
	}
	for {
		fmt.Scanln()
	}
}

//SHAHash returns the SHA1 crypto function result of a given string
func SHAHash(line string) string {
	hasher := sha1.New()
	io.WriteString(hasher, line)
	sha := base64.URLEncoding.EncodeToString(hasher.Sum(nil))
	return string(sha)
}
