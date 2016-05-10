package firebird

/* for connecting to firebird and scan rdb$backup_history table*/

import (
    _"fmt"
    "database/sql"
    _ "github.com/nakagami/firebirdsql"
	"github.com/pharmacy72/gobak/config"
	"github.com/pharmacy72/gobak/backupitems"
	
)

func LastLastChainIntoFirebird(c []*backupitems.BackupItem) (bool, error){
	var n int	
	conn, err := sql.Open("firebirdsql", config.Current().User+":"+config.Current().Password+"@127.0.0.1/"+config.Current().Physicalpathdb)	
	if err != nil {
		return false,err
	}
	defer conn.Close()
	
	for _, itm := range c {
		
		stmt, err := conn.Prepare("SELECT Count(*) FROM rdb$backup_history where rdb$file_name =?")
			if err != nil {
				return false, err
			}
		
		row,err:= stmt.Query(itm.FilePath())
			if err != nil {
				return false, err
			}	
		for row.Next() {
		    if err := row.Scan(&n); err != nil {return false, err}		
	        if n==0{return false, nil}
		}
	
    }	
    return true,nil
}	

