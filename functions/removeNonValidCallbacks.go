package functions

import (
	"database/sql"
)

func RemoveNonValidCallbacks(campaign_id string, DBConn *sql.DB) {
	// remove non valid callbacks
	getCallbackStatus := GetCallbackStatuses(campaign_id, DBConn)
	query := "DELETE vc FROM vicidial_callbacks vc INNER JOIN vicidial_list vl ON(vl.`lead_id`=vc.`lead_id`) WHERE vc.campaign_id='" + campaign_id + "' and vl.`status` NOT IN(" + getCallbackStatus + ")"
	println(query)
	_, err := DBConn.Exec(query)
	if err != nil {
		println(err)
	}

	return
}
