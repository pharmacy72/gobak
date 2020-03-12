package application

import (
	"fmt"
	"github.com/pharmacy72/gobak/config"
	"github.com/urfave/cli/v2"
	"os"
)

func (a *Application) BeforeAction() *Application {
	a.cliApp.Before = func(c *cli.Context) error {
		//Flags
		a.Verbose = c.Bool("verbose")
		a.Start = c.Bool("start") || c.Bool("s")
		a.Debug = c.Bool("debug") || c.Bool("d")
		return nil
	}
	return a
}

func (a *Application) DefaultAction() *Application {

	a.cliApp.Action = func(c *cli.Context) error {

		if err := config.Current().Check(); err != nil {
			return err
		}
		a.fileutils.MakeDirsLevels(config.Current().PathToBackupFolder, config.Current().LevelsConfig.MaxLevel().Int())
		if a.Start && !c.Args().Present() {
			if a.Debug {
				// CLI mode
				fmt.Println(config.Current())
				fmt.Println("Debug mode start...")
				fmt.Println("Hit [Ctrl+C] for exit")
				if err := a.internalRun(); err != nil {
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
		return nil
	}
	return a
}
