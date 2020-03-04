package application

import (
	"github.com/pharmacy72/gobak/bservice"

	"github.com/pharmacy72/gobak/errout"

	"github.com/urfave/cli/v2"
	"log"
	"os"
	"strconv"
)

func (a *Application) lock(c *cli.Context) {
	if err := a.dbopers.DoStartBackup(a.Verbose); err != nil {
		panic(err)
	}
}

func (a *Application) unlock(c *cli.Context) {
	if err := a.dbopers.DoEndBackup(a.Verbose); err != nil {
		panic(err)
	}
}

func (a *Application) check(c *cli.Context) {
	log.Println("Check database start")
	if _, err := a.dbopers.DoCheckBase(a.Verbose, c.Bool("noclear") || c.Bool("n")); err != nil {

		panic(errout.AddSubject(err, ": Check base is not correct"))
	} else {
		log.Println("Check database done")
	}
}

func (a *Application) dbCopy(c *cli.Context) {
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
	if err := a.dbopers.DoCopyDataBase(dst, c.Bool("force") || c.Bool("f"), a.Verbose); err != nil {
		panic(err)
	} else {
		log.Println("Copy done.")
	}
}

func (a *Application) dbRestore(c *cli.Context) {
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
		a.sMail.MailSend("Restore error "+err.Error(), "Restore error", "", "")
		panic(err)
	}
	log.Println("Restore done ")
}

func (a *Application) repolist(c *cli.Context) {
	if err := a.dbopers.DoList(); err != nil {
		panic(err)
	}
}

func (a *Application) repostat(c *cli.Context) {
	if c.IsSet("id") {
		if err := a.dbopers.DoStatBackup(c.StringSlice("id")[:]...); err != nil {
			panic(err)
		}
	} else if err := a.dbopers.DoStat(os.Stdout, c.Bool("hash"), true); err != nil {
		panic(err)
	}
}

func (a *Application) repopack(c *cli.Context) {
	if err := a.dbopers.DoPackItemsServ(a.Verbose); err != nil {

		a.sMail.MailSend("Zipping err "+err.Error(), "Zipping err", "", "")
		panic(err)
	} else {
		a.PrintVerbose("Packed.")
	}
}

func (a *Application) handlerSvcExec(out string, fsvc func() error) func(c *cli.Context) {
	return func(c *cli.Context) {
		if e := fsvc(); e != nil {
			panic(e)
		} else {
			a.PrintVerbose(out)
		}
	}
}
