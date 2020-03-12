package application

import (


	"github.com/pharmacy72/gobak/errout"

	"errors"
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"strconv"
)

func (a *Application) lock(c *cli.Context) error {
	if err := a.dbopers.DoStartBackup(a.Verbose); err != nil {
		return err
	}
	return nil
}

func (a *Application) unlock(c *cli.Context) error {
	if err := a.dbopers.DoEndBackup(a.Verbose); err != nil {
		return err
	}
	return nil
}

func (a *Application) check(c *cli.Context) error {
	log.Println("Check database start")
	if _, err := a.dbopers.DoCheckBase(a.Verbose, c.Bool("noclear") || c.Bool("n")); err != nil {

		return errout.AddSubject(err, ": check base is not correct")
	} else {
		a.log.Info("Check database done")
	}
	return nil
}

func (a *Application) dbCopy(c *cli.Context) error {
	var dst string

	if c.IsSet("output") {
		dst = c.String("output")
	}
	if c.IsSet("o") {
		dst = c.String("o")
	}
	if dst == "" {
		panic(errors.New("file destination unknown"))
	}
	if err := a.dbopers.DoCopyDataBase(dst, c.Bool("force") || c.Bool("f"), a.Verbose); err != nil {
		return err
	} else {
		a.log.Info("Copy done.")
	}
	return nil

}

func (a *Application) dbRestore(c *cli.Context) error {
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
		err = a.bservice.RestoreFromFile(file, dst, c.Bool("hash"), a.Verbose)
	} else if id != "" {
		var iid int
		if iid, err = strconv.Atoi(id); err != nil {
			panic(err)
		}
		err = a.bservice.RestoreFromID(iid, dst, c.Bool("hash"), a.Verbose)
	} else {
		return errors.New("source backup unknown")
	}
	if err != nil {
		a.sMail.MailSend("Restore error "+err.Error(), "Restore error", "", "")
		return err
	}
	a.log.Info("Restore done ")
	return nil
}

func (a *Application) repolist(c *cli.Context) error {
	if err := a.dbopers.DoList(); err != nil {
		return err
	}
	return nil
}

func (a *Application) repostat(c *cli.Context) error {
	if c.IsSet("id") {
		if err := a.dbopers.DoStatBackup(c.StringSlice("id")[:]...); err != nil {
			return err
		}
	} else if err := a.dbopers.DoStat(os.Stdout, c.Bool("hash"), true); err != nil {
		return err
	}
	return nil
}

func (a *Application) repopack(c *cli.Context) error {
	if err := a.dbopers.DoPackItemsServ(a.Verbose); err != nil {

		a.sMail.MailSend("Zipping err "+err.Error(), "Zipping err", "", "")
		return err
	} else {
		a.PrintVerbose("Packed.")
	}

	return nil
}

func (a *Application) handlerSvcExec(out string, fsvc func() error) func(c *cli.Context) error {
	return func(c *cli.Context) error {
		if err := fsvc(); err != nil {
			return err
		} else {
			a.PrintVerbose(out)
		}
		return nil
	}
}
