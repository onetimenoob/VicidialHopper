package functions

import (
	"database/sql"
	"strings"
)

func GetCallbackStatuses(campaign_id string, DBConn *sql.DB) string {
	var statuses []string
	//add 'CBHOLD' and 'CALLBK' to the statuses
	statuses = append(statuses, "'CBHOLD'")
	statuses = append(statuses, "'CALLBK'")

	query := "SELECT status FROM vicidial_campaign_statuses where campaign_id = ? and scheduled_callback='Y' union SELECT status FROM vicidial_statuses where scheduled_callback='Y'"
	rows, err := DBConn.Query(query, campaign_id)
	if err != nil {
		println(err)
		return "dhdfhdfhdfghdfgh"
	}
	defer rows.Close()
	for rows.Next() {
		var status string
		err := rows.Scan(&status)
		if err != nil {
			println(err)
			return "dhdfhdfhdfghdfgh"
		}
		status = "'" + status + "'"
		statuses = append(statuses, status)
	}
	if len(statuses) == 0 {
		return "dhdfhdfhdfghdfgh"
	}
	return strings.Join(statuses, ",")
}
