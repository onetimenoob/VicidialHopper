package functions

import "database/sql"

func RemoveDuplicateCallbacks(DBConn *sql.DB) {
	// remove duplicate callbacks
	query := "DELETE vc\nFROM vicidial_callbacks vc\nJOIN (\n    SELECT lead_id, entry_time\n    FROM vicidial_callbacks\n    WHERE (lead_id, entry_time) NOT IN (\n        SELECT lead_id, MAX(entry_time)\n        FROM vicidial_callbacks\n        GROUP BY lead_id\n    )\n) AS to_delete\nON vc.lead_id = to_delete.lead_id AND vc.entry_time = to_delete.entry_time"
	_, err := DBConn.Exec(query)
	if err != nil {
		println(err)
	}
	return
}
