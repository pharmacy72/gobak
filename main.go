package main

// 31.08.2015 created by Formeo

//TODO: Logger
import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"bytes"

	"github.com/pharmacy72/gobak/bservice"
	"github.com/pharmacy72/gobak/config"
	"github.com/pharmacy72/gobak/dbopers"
	"github.com/pharmacy72/gobak/errout"
	"github.com/pharmacy72/gobak/fileutils"
	"github.com/pharmacy72/gobak/smail"
	"github.com/pharmacy72/gobak/snap"
	"github.com/pharmacy72/gobak/svc"

	"github.com/kardianos/service"
	"github.com/urfave/cli"
)

const version = "0.4.1"
const nameApp = "GoBak"
const copyright = "AO Pharmacy,Tyumen, Russia, 2015-2017"

type application struct {
	cliApp  *cli.App
	svc     service.Service
	Verbose bool
	Start   bool
	Debug   bool
}

var app *application

func (a *application) PrintVerbose(s string) {
	if a.Verbose {
		fmt.Println(s)
	}
}

func (a *application) logerror(err error) {
	if err != nil {
		log.Println(err)
		if a.Verbose {
			fmt.Println(err)
		}
	}
}

func newApplication() *application {

	snap.Start()

	result := &application{}
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
	}, internalRun)
	if err != nil {
		panic(err)
	}

	return result.DefineCommands().
		GlobalFlags().
		BeforeAction().
		DefaultAction()
}
func (a *application) DefaultAction() *application {

	a.cliApp.Action = func(c *cli.Context) {

		if err := config.Current().Check(); err != nil {
			panic(err)
		}
		fileutils.MakeDirsLevels(config.Current().PathToBackupFolder, config.Current().LevelsConfig.MaxLevel().Int())
		if a.Start && !c.Args().Present() {
			if a.Debug {
				// CLI mode
				fmt.Println(config.Current())
				fmt.Println("Debug mode start...")
				fmt.Println("Hit [Ctrl+C] for exit")
				if err := internalRun(); err != nil {
					panic(err)
				}
			} else {
				a.logerror(a.svc.Run())
			}
		}
		if (!c.Args().Present()) && (c.NumFlags() == 0) {
			a.logerror(c.App.Command("help").Run(c))
			os.Exit(1)
		}
		if c.Args().Present() {
			fmt.Printf("Unknow command: %s\n", c.Args().First())
			os.Exit(1)
		} else {
			a.logerror(c.App.Command("help").Run(c))
			os.Exit(1)
		}
	}
	return a
}

func (a *application) BeforeAction() *application {
	a.cliApp.Before = func(c *cli.Context) error {
		//Flags
		a.Verbose = c.Bool("verbose")
		a.Start = c.Bool("start") || c.Bool("s")
		a.Debug = c.Bool("debug") || c.Bool("d")
		return nil
	}
	return a
}

func (a *application) Run() *application {
	if e := a.cliApp.Run(os.Args); e != nil {
		panic(e)
	}
	return a
}

// CLI:Flag application
func (a *application) GlobalFlags() *application {
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

/*func (a *application) handlerContext(fn func(context *cli.Context)) {
	return fn
}*/

func (a *application) lock(c *cli.Context) {
	if err := dbopers.DoStartBackup(a.Verbose); err != nil {
		panic(err)
	}
}

func (a *application) unlock(c *cli.Context) {
	if err := dbopers.DoEndBackup(a.Verbose); err != nil {
		panic(err)
	}
}

func (a *application) check(c *cli.Context) {
	log.Println("Check database start")
	if _, err := dbopers.DoCheckBase(a.Verbose, c.Bool("noclear") || c.Bool("n")); err != nil {

		panic(errout.AddSubject(err, ": Check base is not correct"))
	} else {
		log.Println("Check database done")
	}
}

func (a *application) dbCopy(c *cli.Context) {
	var dst string

	if c.IsSet("output") {
		dst = c.String("output")
	}
	if c.IsSet("o") {
		dst = c.String("o")
	}
	if dst == "" {
		panic(errors.New("File destination unknow"))
	}
	if err := dbopers.DoCopyDataBase(dst, c.Bool("force") || c.Bool("f"), a.Verbose); err != nil {
		panic(err)
	} else {
		log.Println("Copy done.")
	}
}

func (a *application) dbRestore(c *cli.Context) {
	dst := c.String("output")
	if dst == "" {
		dst = c.String("o")
		if dst == "" {
			panic(errors.New("File destination unknow"))
		}
	}
	file := c.String("file")
	id := c.String("id")
	if file != "" && id != "" {
		panic(errors.New("Source backup must be set once"))
	}
	var err error
	if file != "" {
		err = bservice.RestoreFromFile(file, dst, c.Bool("hash"), a.Verbose)
	} else if id != "" {
		var iid int
		if iid, err = strconv.Atoi(id); err != nil {
			panic(err)
		}
		err = bservice.RestoreFromID(iid, dst, c.Bool("hash"), a.Verbose)
	} else {
		panic(errors.New("Source backup unknow"))
	}
	if err != nil {
		smail.MailSend("Restore error "+err.Error(), "Restore error", "", "")
		panic(err)
	}
	log.Println("Restore done ")
}

func (a *application) repolist(c *cli.Context) {
	if err := dbopers.DoList(); err != nil {
		panic(err)
	}
}

func (a *application) repostat(c *cli.Context) {
	if c.IsSet("id") {
		if err := dbopers.DoStatBackup(c.StringSlice("id")[:]...); err != nil {
			panic(err)
		}
	} else if err := dbopers.DoStat(os.Stdout, c.Bool("hash"), true); err != nil {
		panic(err)
	}
}

func (a *application) repopack(c *cli.Context) {
	if err := dbopers.DoPackItemsServ(a.Verbose); err != nil {

		smail.MailSend("Zipping err "+err.Error(), "Zipping err", "", "")
		panic(err)
	} else {
		a.PrintVerbose("Packed.")
	}
}

func (a *application) handlerSvcExec(out string, fsvc func() error) func(c *cli.Context) {
	return func(c *cli.Context) {
		if e := fsvc(); e != nil {
			panic(e)
		} else {
			a.PrintVerbose(out)
		}
	}
}

// CLI:Commands application
func (a *application) DefineCommands() *application {
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

func internalRun() error {
	var statsend time.Time

	for {
		snap.Ping(config.Current().NameBase)
		if time.Now().After(statsend.Add(config.Current().Redis.PeriodStats)) {
			statsend = time.Now()
			snap.Incr(config.Current().NameBase, "counters", snap.CountStats, 1)
			var buf bytes.Buffer
			if err := dbopers.DoStat(&buf, true, false); err != nil {
				log.Println(err)
				snap.Stats(config.Current().NameBase, err.Error(), true)
			} else {
				snap.Stats(config.Current().NameBase, buf.String(), false)
			}
		}
		time.Sleep(time.Duration(config.Current().TimeMsec) * time.Millisecond)
		// formeo
		err := fileutils.DeleteFiles(config.Current().PathToBackupFolder+"//"+config.Current().LevelsConfig.MaxLevel().String(), config.Current().DeleteInt)
		if err != nil {

			log.Println(err)

		}

		err = dbopers.DoBackup(app.Verbose)
		if err != nil {
			snap.Incr(config.Current().NameBase, "counters", snap.CounterErrorBackup, 1)
			log.Println("Backup error: ", err)
			if eo, ok := err.(*errout.ErrOut); ok {
				log.Println(eo.StdErrOutput())
				//log.Println(eo.StdOutput())
				snap.BackupDone(config.Current().NameBase, "", "", eo.StdErrOutput())

			} else {
				snap.BackupDone(config.Current().NameBase, "", "", err.Error())
			}
			if !errOut2mail(err) {
				smail.MailSend("Backup error:"+err.Error(), "Error backup", "", "")
			}
		}

	}
}

func errOut2mail(err interface{}) bool {
	if eo, ok := err.(*errout.ErrOut); ok {
		so, se := eo.StdOutput(), eo.StdErrOutput()
		fmt.Fprint(os.Stdout, so)
		fmt.Fprint(os.Stderr, se)
		if eo.Report {
			smail.MailSend(string(so+"\n"+se), config.Current().NameBase+eo.Subject, "", "")
		}
		return true
	}
	return false
}

func main() {
	defer func() {
		if e := recover(); e != nil {
			if app == nil || !app.Verbose {
				if !errOut2mail(e) {
					fmt.Fprintln(os.Stderr, e)
				}
			} else if app.Verbose {
				fmt.Fprintln(os.Stderr, e.(error))
			}
			os.Exit(1)
		}
	}()
	app = newApplication()
	app.Run()
	return

}
