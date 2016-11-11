package dbopers

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/pharmacy72/gobak/backupitems"
	"github.com/pharmacy72/gobak/bservice"
	"github.com/pharmacy72/gobak/config"
	"github.com/pharmacy72/gobak/dbfile"
	"github.com/pharmacy72/gobak/fileutils"
	"github.com/pharmacy72/gobak/firebird"
	"github.com/pharmacy72/gobak/level"

	"github.com/arteev/fmttab"
	"github.com/nsf/termbox-go"
	"github.com/pharmacy72/gobak/snap"

	"github.com/pharmacy72/gobak/errout"
	"io"
	"bufio"
	"github.com/pharmacy72/gobak/md5f"
)

func doVerbose(verbose bool, a ...interface{}) {
	if verbose {
		fmt.Printf(a[0:1][0].(string), a[1:]...)
	}
}

// DoCheckBase checkbase copy: lock->copy->unlock->check copy
func DoCheckBase(verbose bool, notclear bool) (res bool, err error) {
	cfg := config.Current()
	dbf := dbfile.New(cfg.Physicalpathdb, cfg.AliasDb, cfg.User, cfg.Password, verbose)
	unlock := func() bool {
		e := dbf.Unlock(false)
		if e != nil {
			res = false
			err = e
		}
		return e == nil
	}

	if err := dbf.Lock(); err != nil {
		return false, err
	}
	defer unlock()

	dbnamecpy := fileutils.GetTempFile(filepath.Dir(dbf.Filename), "copy."+filepath.Base(dbf.Filename))
	dbfcopy, e := dbf.Copy(dbnamecpy, false)
	if e != nil {
		return false, e
	}
	if !notclear {
		defer func() {
			if err := dbfcopy.Remove(); err != nil {
				log.Println("failed remove copy db ", err)
			}
		}()
	}
	if !unlock() {
		return
	}
	if e := dbfcopy.Fixup(); e != nil {
		return false, e
	}
	if e := dbfcopy.Check(); e != nil {
		return false, e
	}
	return true, nil
}

//DoEndBackup Send database end backup
func DoEndBackup(verbose bool) error {
	cfg := config.Current()
	dbf := dbfile.New(cfg.Physicalpathdb, cfg.AliasDb, cfg.User, cfg.Password, verbose)
	if err := dbf.Unlock(true); err != nil {
		return err
	}
	return nil
}

//DoStartBackup Send database start backup
func DoStartBackup(verbose bool) error {
	cfg := config.Current()
	dbf := dbfile.New(cfg.Physicalpathdb, cfg.AliasDb, cfg.User, cfg.Password, verbose)
	if err := dbf.Lock(); err != nil {
		return err
	}
	return nil
}

//DoCopyDataBase Copy database with lock/unlock
func DoCopyDataBase(dest string, ovewrite bool, verbose bool) (err error) {
	cfg := config.Current()
	dbf := dbfile.New(cfg.Physicalpathdb, cfg.AliasDb, cfg.User, cfg.Password, verbose)
	unlock := func() bool {
		e := dbf.Unlock(false)
		if e != nil {
			err = e
		}
		return e == nil
	}
	defer unlock()
	if err := dbf.Lock(); err != nil {
		return err
	}

	log.Println("Copy to ", dest)
	dbfcopy, e := dbf.Copy(dest, ovewrite)
	if e != nil {
		return e
	}
	unlock()
	log.Println("Fixup", dbfcopy.Alias)
	if e := dbfcopy.Fixup(); e != nil {
		return e
	}

	return e
}

//DoBackup do backup current database
func DoBackup(verbose bool) error {
	var FintoF bool
	backupLevels := make(map[level.Level]*backupitems.BackupItem)
	repo := backupitems.GetRepository()
	//defer repo.Close()
	backups := repo.All()
	all, err := backups.Get()
	if err != nil {
		return err
	}

	if all != nil {
		all = all[len(all)-1].ChainWithAllParents()
		FintoF, err = firebird.LastLastChainIntoFirebird(all)
		if err != nil {
			return err
		}
		for _, b := range all {
			backupLevels[b.Level] = b
		}
	}

	// for each level

	levelFirst := config.Current().LevelsConfig.First().Level
	maxLevel := *config.Current().LevelsConfig.MaxLevel()
	var parentGUID string
	for i := levelFirst; !i.Equals(maxLevel.Next()); i = i.Next() {
		isActual := false
		if bkp, ok := backupLevels[i]; ok {
			isActual, err = config.Current().LevelsConfig.IsActual(bkp.Level, bkp.Date, time.Now())
			if err != nil {
				return err
			}
			if isActual || (parentGUID == "" && !i.IsFirst()) {
				parentGUID = bkp.GUID
			}
		}
		if !FintoF {
			isActual = false
		}
		if !isActual {
			doVerbose(verbose, "Start do backup level:%s\n", i)
			snap.Incr(config.Current().NameBase,"counters",snap.CounterStartBackup,1)
			if newbkp, e := bservice.Backup(verbose, i, parentGUID); e == nil {
				doVerbose(verbose, "Successful backup %s. file %s\n", i, newbkp.FileName)
				backupLevels[i] = newbkp
				parentGUID = newbkp.GUID

				snap.Incr(config.Current().NameBase,"counters",snap.CounterSuccessBackup,1)
				size:=fileutils.SizeToFredly(fileutils.Size(newbkp.FilePath()))
				snap.BackupDone(config.Current().NameBase,i.String(),size,"")
			} else {
				log.Printf("FAILED: %s\n", e)
				return e
			}
			checkLvl, _ := config.Current().LevelsConfig.Find(i)
			if checkLvl.Check {
				snap.Incr(config.Current().NameBase,"counters",snap.CounterCheck,1)
				_, e := DoCheckBase(verbose, false)

				if e!=nil {
					var errortext string
					 if eo, ok := err.(*errout.ErrOut); ok {
						 errortext = eo.StdErrOutput()
					 } else {
						 errortext = e.Error()
					 }
					snap.CheckDB(config.Current().NameBase, errortext,true)
				} else {
					snap.CheckDB(config.Current().NameBase,"OK",false)
				}


				if (e != nil) && (e != dbfile.ErrCheckBase) {
					log.Print("CHECKBASE FAILED\n")
					return e
				}
			}
		}
	}
	//Save into repository
	for i := levelFirst; !i.Equals(maxLevel.Next()); i = i.Next() {
		item, _ := backupLevels[i]
		if item.Insert {
			if e := repo.Append(item); e != nil {
				return e
			}
		} else if item.Modified {
			if e := repo.Update(item); e != nil {
				return e
			}
		}

		//snap.
		//
	}
	return nil
}

//DoPackItemsServ Packing All Items who are not actual
func DoPackItemsServ(verbose bool) (err error) {
	var arr []*backupitems.BackupItem
	repo := backupitems.GetRepository()
	defer repo.Close()
	backups := repo.All()
	if arr, err = backups.Get(); err != nil {
		return err
	} else if arr == nil {
		return nil
	}
	count := 0
	var sizewas, sizenow int64
	for i := 0; i < len(arr); i++ {
		actual, _ := config.Current().LevelsConfig.IsActual(arr[i].Level, arr[i].Date, time.Now())
		if actual || arr[i].IsArchived() {
			continue
		}
		//check hash before pack

		doVerbose(verbose, "Check hash\n")
		if ok, e := arr[i].HashValid(); !ok {
			return e
		}
		swl := fileutils.Size(arr[i].FilePath())
		doVerbose(verbose, "Pack: %s\n", arr[i].FilePath())

		if err = arr[i].PackItem(true); err != nil {
			log.Println("Error packing:", arr[i].FilePath(), err)
			return err
		}
		sizewas += swl
		sizenow += fileutils.Size(arr[i].FilePath())
		if err := arr[i].ComputeHash(); err != nil {
			return err
		}
		count++
	}

	for _, item := range arr {
		var err error
		if item.Insert {
			err = repo.Append(item)
		} else if item.Modified {
			err = repo.Update(item)
		}
		if err != nil {
			return err
		}
	}
	/*if err = dbase.WriteBackupItems(arr); err != nil {
		return err
	}*/
	doVerbose(verbose, "Packed files: %d, released space: %s \n", count, fileutils.SizeToFredly(sizewas-sizenow))
	return nil
}

//DoList print a table with information about backups
func DoList() error {
	repo := backupitems.GetRepository()
	defer repo.Close()
	backups := repo.All()

	arr, err := backups.Get() //dbase.FetchBackupItems()
	if err != nil {

		return err
	}
	if arr == nil {
		fmt.Println("Not found records")
		return nil
	}
	ach := map[bool]string{
		false: " ",
		true:  "+",
	}
	tab := fmttab.New("Backups", fmttab.BorderDouble, nil)
	if err := termbox.Init(); err != nil {
		return nil
	}
	tw, _ := termbox.Size()
	termbox.Close()

	tab.AddColumn("ID", 8, fmttab.AlignLeft).
		AddColumn("LV", 2, fmttab.AlignRight).
		AddColumn("P", 1, fmttab.AlignLeft).
		AddColumn("A", 1, fmttab.AlignLeft).
		AddColumn("GUID", 36, fmttab.AlignLeft).
		AddColumn("PREV", fmttab.WidthAuto, fmttab.AlignLeft).
		AddColumn("HASH", fmttab.WidthAuto, fmttab.AlignLeft).
		AddColumn("DATE", fmttab.WidthAuto, fmttab.AlignLeft).
		AddColumn("SIZE", fmttab.WidthAuto, fmttab.AlignRight).
		AddColumn("PATH", fmttab.WidthAuto, fmttab.AlignLeft)
	for _, b := range arr {
		pt, err := filepath.Rel(config.Current().PathToBackupFolder, b.FilePath())
		if err != nil {
			pt = b.FileName
		} else {
			pt = "{bkp}/" + pt
		}
		isActual, _ := config.Current().LevelsConfig.IsActual(b.Level, b.Date, time.Now())
		var id string
		if !b.Level.IsFirst() {
			id = fmt.Sprintf("%s%d", strings.Repeat(" ", b.Level.Int()), b.ID)
		} else {
			id = strconv.Itoa(b.ID)
		}
		tab.AppendData(map[string]interface{}{
			"ID":   id,
			"LV":   b.Level,
			"GUID": b.GUID,
			"PREV": b.GUIDParent,
			"HASH": b.Hash,
			"DATE": b.Date.Format("2006-01-02 15:04"),
			"PATH": pt,
			"P":    ach[b.IsArchived()],
			"A":    ach[isActual],
			"SIZE": fileutils.SizeToFredly(fileutils.Size(b.FilePath())),
		})
	}
	tab.AutoSize(true, tw)
	tab.WriteTo(os.Stdout)
	return nil
}

type statistic struct {
	min   int64
	max   int64
	all   int64
	count int
}

//DoStat print a statistic with information about backups
func DoStat(w io.Writer, hashcheck bool, autosize bool) error {
	buf := bufio.NewWriter(w)
	repo := backupitems.GetRepository()
	//defer repo.Close()
	backups := repo.All()

	arr, err := backups.Get()
	if err != nil {
		return err
	}

	buf.WriteString("Statistics repository")
	KeysCommon := []string{"Count", "Found", "Archived", "Not Found", "Corrupt"}
	var maxLevel level.Level
	var AllSize int64
	var hashCorruptItems []*backupitems.BackupItem
	var notfoundItems []*backupitems.BackupItem
	levelStat := make(map[int]*statistic)
	CommonStat := make(map[string]int)
	CommonStat["Count"] = len(arr)
	CommonStat["Archived"] = 0
	CommonStat["Found"] = 0
	CommonStat["Not Found"] = 0

	if hashcheck {
		CommonStat["Corrupt"] = 0
	}

	for _, item := range arr {
		if maxLevel.Int() <= item.Level.Int() {
			maxLevel = item.Level
		}
		cz := fileutils.Size(item.FilePath())
		cur, ok := levelStat[item.Level.Int()]
		if !ok {
			levelStat[item.Level.Int()] = &statistic{count: 1, min: cz, max: cz, all: cz}
		} else {
			cur.count++
			cur.all += cz
			if cz > cur.max {
				cur.max = cz
			}
			if cz < cur.min && cz != 0 {
				cur.min = cz
			}
		}
		if item.IsArchived() {
			CommonStat["Archived"]++
		}
		if item.Exists() {
			CommonStat["Found"]++
			AllSize += fileutils.Size(item.FilePath())
			if hashcheck {
				if ok, err := item.HashValid(); !ok {
					if err != nil && err!=md5f.ErrFileCorrupt {
						buf.WriteString(err.Error())
					} else {
						hashCorruptItems = append(hashCorruptItems, item)
						CommonStat["Corrupt"]++
					}
				}
			}
		} else {
			CommonStat["Not Found"]++
			notfoundItems = append(notfoundItems, item)
		}
	}

	buf.WriteString("\nLevels:")
	for i := 0; i <= maxLevel.Int(); i++ {
		if value, ok := levelStat[i]; ok {
			buf.WriteString(fmt.Sprintf("\tLevel %d: %d, size %s,min/max/avg  %s/%s/%s\n", i, value.count,
				fileutils.SizeToFredly(value.all),
				fileutils.SizeToFredly(value.min),
				fileutils.SizeToFredly(value.max),
				fileutils.SizeToFredly(value.all/int64(value.count))))
		}
	}
	buf.WriteString("Statistic:")
	for _, key := range KeysCommon {
		if value, ok := CommonStat[key]; ok {
			buf.WriteString(fmt.Sprintf("\t%s: %d\n", key, value))

		}
	}
	buf.WriteString(fmt.Sprintf("\tTotal size: %s\n", fileutils.SizeToFredly(AllSize)))
	if arr != nil {
		tab := fmttab.New("Last chain", fmttab.BorderDouble, nil)
		tab.AddColumn("ID", 5, fmttab.AlignRight).
			AddColumn("LV", 2, fmttab.AlignRight).
			AddColumn("DATE", fmttab.WidthAuto, fmttab.AlignLeft).
			AddColumn("SIZE", fmttab.WidthAuto, fmttab.AlignRight).
			AddColumn("PATH", fmttab.WidthAuto, fmttab.AlignLeft)
		lastCh := arr[len(arr)-1].ChainWithAllParents()
		var totalsize int64
		for _, j := range lastCh {

			pt, err := filepath.Rel(config.Current().PathToBackupFolder, j.FilePath())
			if err != nil {
				pt = j.FileName
			} else {
				pt = "{bkp}/" + pt
			}
			size := fileutils.Size(j.FilePath())
			totalsize += size
			tab.AppendData(map[string]interface{}{
				"ID":   j.ID,
				"LV":   j.Level,
				"SIZE": fileutils.SizeToFredly(size),
				"DATE": j.Date.Format("2006-01-02 15:04"),
				"PATH": pt,
			})
		}

		tab.AppendData(map[string]interface{}{
			"DATE": fmt.Sprintf("Total:%d", len(lastCh)),
			"SIZE": fileutils.SizeToFredly(totalsize),
		})

		buf.WriteString("\nLast chain's data:\n")
		if autosize {
			if err := termbox.Init(); err != nil {
				return nil
			}
			tw, _ := termbox.Size()
			termbox.Close()
			tab.AutoSize(true, tw)
		}
		tab.WriteTo(buf)

	}

	if len(notfoundItems) != 0 {
		tab := fmttab.New("Missing backups", fmttab.BorderDouble, nil)
		tab.AddColumn("ID", 5, fmttab.AlignRight).
			AddColumn("LV", 2, fmttab.AlignRight).
			AddColumn("DATE", fmttab.WidthAuto, fmttab.AlignLeft).
			AddColumn("PATH", fmttab.WidthAuto, fmttab.AlignLeft)
		for _, j := range notfoundItems {
			pt, err := filepath.Rel(config.Current().PathToBackupFolder, j.FilePath())
			if err != nil {
				pt = j.FileName
			} else {
				pt = "{bkp}/" + pt
			}
			tab.AppendData(map[string]interface{}{
				"ID":   j.ID,
				"LV":   j.Level,
				"DATE": j.Date.Format("2006-01-02 15:04"),
				"PATH": pt,
			})
		}
		if autosize {
			if err := termbox.Init(); err != nil {
				return nil
			}
			tw, _ := termbox.Size()
			termbox.Close()
			tab.AutoSize(true, tw)
		}

		tab.WriteTo(buf)

	}
	if len(hashCorruptItems) != 0 {
		tab := fmttab.New("Corrupt files", fmttab.BorderDouble, nil)
		tab.AddColumn("ID", 5, fmttab.AlignRight).
			AddColumn("LV", 2, fmttab.AlignRight).
			AddColumn("DATE", fmttab.WidthAuto, fmttab.AlignLeft).
			AddColumn("HASH", fmttab.WidthAuto, fmttab.AlignLeft).
			AddColumn("PATH", fmttab.WidthAuto, fmttab.AlignLeft)
		for _, j := range hashCorruptItems {
			pt, err := filepath.Rel(config.Current().PathToBackupFolder, j.FilePath())
			if err != nil {
				pt = j.FileName
			} else {
				pt = "{bkp}/" + pt
			}
			tab.AppendData(map[string]interface{}{
				"ID":   j.ID,
				"LV":   j.Level,
				"HASH": j.Hash,
				"DATE": j.Date.Format("2006-01-02 15:04"),
				"PATH": pt,
			})
		}
		if autosize {
			if err := termbox.Init(); err != nil {
				return nil
			}

			tw, _ := termbox.Size()
			termbox.Close()
			tab.AutoSize(true, tw)
		}

		tab.WriteTo(buf)


	}

	return buf.Flush()
}

//DoStatBackup print a statistic with information about a backup
func DoStatBackup(id ...string) error {
	repo := backupitems.GetRepository()
	defer repo.Close()
	col := repo.All()
	col.AddFilterID(id...)
	backups, err := col.Get()
	if err != nil {
		return err
	}
	if len(backups) == 0 {
		return bservice.ErrIDSourceNotFound
	}
	bstr := map[bool]string{
		false: "No",
		true:  "Yes",
	}
	for _, item := range backups {
		fmt.Println("---- Backup information: ----")
		fmt.Println("\tID:", item.ID)
		fmt.Println("\tLevel:", item.Level)
		fmt.Println("\tDate:", item.Date.Format("2006-01-02 15:04:01"))
		fmt.Println("\tHash:", item.Hash)
		fmt.Println("\tGUID:", item.GUID)
		fmt.Println("\tParent guid:", item.GUIDParent)
		if item.Parent != nil {
			fmt.Println("\tParent:", item.Parent.ID)
		} else {
			fmt.Println("\tParent: <nil>")
		}
		fmt.Println("\tName:", item.FileName)
		fmt.Println("\tPath:", item.FilePath())
		fmt.Println("\tPacked:", bstr[item.IsArchived()])
		if item.Exists() {
			fmt.Println("\tExists file: Yes")
			valid, _ := item.HashValid()
			fmt.Println("\tCorrupt file:", bstr[!valid])
			fmt.Println("\tSize:", fileutils.SizeToFredly(fileutils.Size(item.FilePath())))

		} else {
			fmt.Println("\tExists file: No")
		}
	}
	return nil
}
