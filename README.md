```
======================================
gobak (Go utility for nbackup firebird)
======================================
```
[![Build Status](https://travis-ci.org/pharmacy72/gobak.svg?branch=master)](https://travis-ci.org/pharmacy72/gobak)

## Description
 * utility for nbackup [FireBird](http://firebirdsql.org) on Golang

## Documentation
 * [Quickstart](quickstart.md)
 * [Changelog](changelog.txt)
 
## Description config.json
 * "PathToNbackup": "/usr/bin/nbackup", - **path to nbackup utility**
 * "DirectIO":true,
 * "PathToBackupFolder": "/home/test/backup/", - **path to backup folder**
 * "AliasDb": "/home/bases/clear_base.fdb", - **alias Database**
 * "User": "sysdba", - **username**
 * "Password": "masterkey", - **password**
 * "Physicalpathdb": "/home/bases/clear_base.fdb", 
 * "EmailFrom": "testtest@test.ru", 
 * "EmailTo": "test@test.ru",
 * "SmtpServer": "127.0.0.1:25", - **SMTP server, need set for correct sending email with backup errors**
 * "Pathtogfix": "/usr/bin/gfix", - **path to gfix utility**
 * "NameBase": "TESTDB", - **name database file , default value alias** 
 * "TimeMlsc": 6000, - **interval check backups ms**
 * "levels":[
    {
      "level":0,
      "tick":"H",
      "check":false
    },
    {
      "level":1,
      "tick":"N:5",
      "check" : false
    }
  ]
} 
- **backup level, level -level, tick - (hour,week,day,hour), check - checking base(gfix -v -full)**



## Install
go get github.com/pharmacy72/gobak

## Usage

## License
MIT:

## Authors
Arteev Aleksey

Gordienko Roman
