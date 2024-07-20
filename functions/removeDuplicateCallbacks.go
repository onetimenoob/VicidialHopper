package functions

import "database/sql"

func RemoveDuplicateCallbacks(DBConn *sql.DB) {
	// remove duplicate callbacks
	query := "DELETE vc FROM vicidial_callbacks vc JOIN (    SELECT lead_id, entry_time   FROM vicidial_callbacks    WHERE (lead_id, entry_time) NOT IN (        SELECT lead_id, MAX(entry_time)       FROM vicidial_callbacks        GROUP BY lead_id    )) AS to_delete ON vc.lead_id = to_delete.lead_id AND vc.entry_time = to_delete.entry_time"
	_, err := DBConn.Exec(query)
	if err != nil {
		println(err)
	}
	return
}
