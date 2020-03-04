package application

import (
	"bytes"
	"fmt"
	"github.com/kardianos/service"
	"github.com/pharmacy72/gobak/config"
	"github.com/pharmacy72/gobak/dbopers"
	"github.com/pharmacy72/gobak/errout"
	"github.com/pharmacy72/gobak/fileutils"
	"github.com/pharmacy72/gobak/smail"
	"github.com/pharmacy72/gobak/snap"
	"github.com/pharmacy72/gobak/svc"
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"path/filepath"
	"time"
)

const version = "0.5.0"
const nameApp = "GoBak"
const copyright = "AO Pharmacy,Tyumen, Russia, 2015-2020"

type Application struct {
	cliApp    *cli.App
	svc       service.Service
	sMail     *smail.MailApp
	dbopers   *dbopers.Database
	fileutils *fileutils.FileUtils
	Verbose   bool
	Start     bool
	Debug     bool
}

var app *Application

func (a *Application) PrintVerbose(s string) {
	if a.Verbose {
		fmt.Println(s)
	}
}
func (a *Application) logerror(err error) {
	if err != nil {
		log.Println(err)
		if a.Verbose {
			fmt.Println(err)
		}
	}
}

func (a *Application) ErrOut2mail(err interface{}) bool {
	if eo, ok := err.(*errout.ErrOut); ok {
		so, se := eo.StdOutput(), eo.StdErrOutput()
		fmt.Fprint(os.Stdout, so)
		fmt.Fprint(os.Stderr, se)
		if eo.Report {
			a.sMail.MailSend(so+"\n"+se, config.Current().NameBase+eo.Subject, "", "")
		}
		return true
	}
	return false
}

func (a *Application) DefineCommands() *Application {
	a.cliApp.Commands = []cli.Command{
		{
			Name:        "database",
			Aliases:     []string{"db"},
			Usage:       "Database manipulations",
			Description: "Database manipulations",
			Subcommands: []cli.Command{
				{
					Name:   "lock",
					Usage:  "Lock for backup",
					Action: a.lock,
				}, //lock
				{
					Name:   "unlock",
					Usage:  "Unlock for backup",
					Action: a.unlock,
				}, //unlock

				{
					Name:  "check",
					Usage: "Check database",
					Flags: []cli.Flag{
						cli.BoolFlag{
							Name:  "noclear,n",
							Usage: "Don't remove a database copy",
						},
					},
					Action: a.check,
				}, //check

				{
					Name:    "copy",
					Aliases: []string{"cp"},
					Usage:   "Copy database file with lock and fixup destination database",
					Flags: []cli.Flag{
						cli.BoolFlag{
							Name:  "force, f",
							Usage: "Warning!!! overwrite destination ",
						},
						cli.StringFlag{
							Name:  "output, o",
							Usage: "FileName of destination",
						},
					},
					Action: a.dbCopy,
				},

				{
					Name:  "restore",
					Usage: "Restore database",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "file",
							Usage: "Restory from file(must be full path)",
						},
						cli.StringFlag{
							Name:  "id",
							Usage: "Restory by identifier",
						},
						cli.StringFlag{
							Name:  "output, o",
							Usage: "FileName of destination",
						},
						cli.BoolTFlag{
							Name:  "hash",
							Usage: "Check hash for each the backup file",
						},
					},
					Action: a.dbRestore,
				},
			},
		},
		{
			Name:        "repository",
			Aliases:     []string{"repo"},
			Usage:       "Repository operations",
			Description: "Repository operations",
			Subcommands: []cli.Command{
				{
					Name:        "list",
					Usage:       "show all backup entites into repository",
					Description: "list backups",
					Action:      a.repolist,
				},
				{
					Name:        "statistic",
					Aliases:     []string{"stat"},
					Usage:       "show statistics from the repository: count files, count packed, last backups",
					Description: "show statistics",
					Flags: []cli.Flag{
						cli.BoolFlag{
							Name:  "hash",
							Usage: "Check hash of file",
						},
						cli.StringSliceFlag{
							Name:  "id",
							Usage: "id(s) backup.View information about a backup",
						},
					},
					Action: a.repostat,
				},
				{
					Name:   "pack",
					Usage:  "Packed all backup files except active chain",
					Action: a.repopack,
				},
			},
		},

		{
			Name:        "service",
			Aliases:     []string{"svc", "daemon"},
			Usage:       "sevice(daemon) managment",
			Description: "sevice(daemon) managment",
			Subcommands: []cli.Command{
				{
					Name:   "install",
					Usage:  "install service(daemon)",
					Action: a.handlerSvcExec("Service installed", a.svc.Install),
				},
				{
					Name:   "uninstall",
					Usage:  "uninstall service(daemon)",
					Action: a.handlerSvcExec("Service uninstalled", a.svc.Uninstall),
				},
				{
					Name:   "start",
					Usage:  "start service(daemon)",
					Action: a.handlerSvcExec("Service started", a.svc.Start),
				},
				{
					Name:   "stop",
					Usage:  "stop service(daemon)",
					Action: a.handlerSvcExec("Service stoped", a.svc.Stop),
				},
				{
					Name:   "restart",
					Usage:  "restart service(daemon)",
					Action: a.handlerSvcExec("Service restarted", a.svc.Restart),
				},
			},
		},
	}
	return a
}

func (a *Application) internalRun() error {
	var statsend time.Time

	for {
		snap.Ping(config.Current().NameBase)
		if time.Now().After(statsend.Add(config.Current().Redis.PeriodStats)) {
			statsend = time.Now()
			snap.Incr(config.Current().NameBase, "counters", snap.CountStats, 1)
			var buf bytes.Buffer
			if err := a.dbopers.DoStat(&buf, true, false); err != nil {
				log.Println(err)
				snap.Stats(config.Current().NameBase, err.Error(), true)
			} else {
				snap.Stats(config.Current().NameBase, buf.String(), false)
			}
		}
		time.Sleep(time.Duration(config.Current().TimeMsec) * time.Millisecond)
		// formeo

		//err := fileutils.DeleteFiles(config.Current().PathToBackupFolder+"/"+config.Current().LevelsConfig.MaxLevel().String(), config.Current().DeleteInt)
		fs, err := a.fileutils.FreeSpace(config.Current().PathToBackupFolder)
		if err != nil {

			log.Println(err)

		}
		fmt.Println("freeSpace - ", fs/1024)

		err = a.fileutils.DeleteFiles(filepath.Join(config.Current().PathToBackupFolder, config.Current().LevelsConfig.MaxLevel().String()), config.Current().DeleteInt)

		if err != nil {

			log.Println(err)

		}

		err = a.dbopers.DoBackup(app.Verbose)

		if err != nil {
			snap.Incr(config.Current().NameBase, "counters", snap.CounterErrorBackup, 1)
			log.Println("Backup error: ", err)
			if eo, ok := err.(*errout.ErrOut); ok {
				log.Println(eo.StdErrOutput())
				snap.BackupDone(config.Current().NameBase, "", "", eo.StdErrOutput())

			} else {
				snap.BackupDone(config.Current().NameBase, "", "", err.Error())
			}

			if !a.ErrOut2mail(err) {
				a.sMail.MailSend("Backup error:"+err.Error(), "Error backup", "", "")
			}
		}

	}
}

func (a *Application) GlobalFlags() *Application {
	a.cliApp.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "start, s",
			Usage: "start service(daemon) or CLI mode",
		},
		cli.BoolFlag{
			Name:  "debug, d",
			Usage: "enable debug mode when -start on CLI mode",
		},
		cli.BoolFlag{
			Name:  "verbose",
			Usage: "Increment verbosity",
		},
	}
	return a
}

func (a *Application) Run() *Application {
	if e := a.cliApp.Run(os.Args); e != nil {
		panic(e)
	}
	return a
}
func NewApplication() *Application {

	snap.Start()

	result := &Application{}
	result.cliApp = cli.NewApp()
	result.cliApp.Name = nameApp
	result.cliApp.Version = version
	result.cliApp.Copyright = copyright
	result.cliApp.Usage = "The utility for backup and maintenance Firebird bases"

	var err error
	result.svc, err = svc.New(&service.Config{
		Name:        "gobak",
		DisplayName: "Increment backup for Firebird",
		Description: "Service Increment backup and check for Firebird over nbackup",
		Arguments:   []string{"-start"},
	}, result.internalRun)
	if err != nil {
		panic(err)
	}

	return result.DefineCommands().
		GlobalFlags().
		BeforeAction().
		DefaultAction()
}
