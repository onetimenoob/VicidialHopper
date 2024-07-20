package functions

import (
	"database/sql"
)

func RemoveNonValidCallbacks(campaignId string, DBConn *sql.DB) {
	// remove non valid callbacks
	getCallbackStatus := GetCallbackStatuses(campaignId, DBConn)
	query := "DELETE vc FROM vicidial_callbacks vc INNER JOIN vicidial_list vl ON(vl.`lead_id`=vc.`lead_id`) WHERE vc.campaign_id='" + campaignId + "' and vl.`status` NOT IN(" + getCallbackStatus + ")"
	println(query)
	_, err := DBConn.Exec(query)
	if err != nil {
		println(err)
	}

	return
}
