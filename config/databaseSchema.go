package config

//DBSchema is the query used to seed the database
var DBSchema = `

/*
https://sqlite.org/pragma.html


PRAGMA page_size = 4096;
PRAGMA cache_size = 16384;
PRAGMA temp_store = MEMORY;

PRAGMA journal_mode = OFF;
PRAGMA synchronous = OFF;
PRAGMA locking_mode = EXCLUSIVE;

PRAGMA automatic_index = ON;
PRAGMA cache_size = 32768;
PRAGMA cache_spill = OFF;
PRAGMA foreign_keys = ON;
PRAGMA journal_size_limit = 67110000;
PRAGMA locking_mode = NORMAL;
PRAGMA page_size = 4096;
PRAGMA recursive_triggers = ON;
PRAGMA secure_delete = ON;
PRAGMA synchronous = NORMAL;
PRAGMA temp_store = MEMORY;
PRAGMA journal_mode = WAL;
PRAGMA wal_autocheckpoint = 16384;
*/





PRAGMA page_size = 4096;
PRAGMA cache_size = 131072;
PRAGMA temp_store = MEMORY;

--PRAGMA journal_mode = OFF;
PRAGMA synchronous = OFF;
PRAGMA locking_mode = EXCLUSIVE;

PRAGMA automatic_index = ON;
--PRAGMA foreign_keys = ON;
PRAGMA journal_size_limit = 67110000;
--PRAGMA locking_mode = NORMAL;
PRAGMA recursive_triggers = OFF;
PRAGMA secure_delete = OFF;
--PRAGMA synchronous = NORMAL;
PRAGMA journal_mode = WAL;
PRAGMA wal_autocheckpoint = 32768;

		
CREATE TABLE IF NOT EXISTS "Jails" (
	"ID"	INTEGER PRIMARY KEY AUTOINCREMENT,
	"Name"	TEXT NOT NULL UNIQUE,
	"FirstLineHash"	TEXT,
	"LastScannedLine"	INTEGER
);
CREATE TABLE IF NOT EXISTS "IPsCounter" (
	"ID"	INTEGER PRIMARY KEY AUTOINCREMENT,
	"Jail"	INTEGER,
	"IP"	TEXT NOT NULL,
	"Counter"	INTEGER NOT NULL DEFAULT 0,
	FOREIGN KEY("Jail") REFERENCES "Jails"("ID"),
	CONSTRAINT con_ipsconunter_jail_ip_unique UNIQUE ('Jail','IP')
);
CREATE TABLE IF NOT EXISTS "Logs" (
	"ID"	INTEGER PRIMARY KEY AUTOINCREMENT,
	"Jail"	INTEGER,
	"IP"	TEXT,
	"RequestNumber"	INTEGER NOT NULL,
	"Timestamp"	INTEGER NOT NULL,
	"Burst" INTEGER DEFAULT 0,
	FOREIGN KEY("IP") REFERENCES "IPsCounter"("IP"),
	FOREIGN KEY("Jail") REFERENCES "Jails"("ID"),
	CONSTRAINT con_logs_jail_ip_ts_unique UNIQUE ('Jail','IP','RequestNumber')
);
CREATE TABLE IF NOT EXISTS "Bans" (
	"ID"	INTEGER PRIMARY KEY AUTOINCREMENT,
	"Jail"	INTEGER,
	"IP"	TEXT,
	"Timestamp"	TEXT CURRENT_TIMESTAMP,
	FOREIGN KEY("IP") REFERENCES "IPsCounter"("IP"),
	FOREIGN KEY("Jail") REFERENCES "Jails"("ID"),
	CONSTRAINT con_bans_jail_ip_unique UNIQUE ('Jail','IP')
);
`
