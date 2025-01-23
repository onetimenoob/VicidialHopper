package functions

import (
	"database/sql"
	"fmt"
	"strings"
)

func UpdateCallbacks(campaignID string, DBCONN *sql.DB) {
	activeListsStr := strings.Join(getCampaignActiveLists(DBCONN, campaignID), ",")
	if activeListsStr == "" {
		return
	}
	query := "SELECT vicidial_callbacks.lead_id FROM vicidial_callbacks,vicidial_list where callback_time <= NOW() and vicidial_callbacks.status IN('ACTIVE') and vicidial_callbacks.lead_id=vicidial_list.lead_id and vicidial_callbacks.campaign_id='" + campaignID + "' and vicidial_callbacks.list_id IN(" + activeListsStr + ")"

	rows, err := DBCONN.Query(query)
	if err != nil {
		println(err.Error())
		return
	}
	defer rows.Close()
	var updateLeads []string
	for rows.Next() {
		var leadID string
		err := rows.Scan(&leadID)
		if err != nil {
			fmt.Printf("!!!!!!!!!!!!!!!!!!!!!!!!!!!1Error scanning row: %s\n", err.Error())
			return
		}
		updateLeads = append(updateLeads, leadID)
	}
	if len(updateLeads) == 0 {
		return
	}
	updateStr := strings.Join(updateLeads, ",")
	query = "UPDATE vicidial_callbacks set status='LIVE' where lead_id IN(" + updateStr + ") and status NOT IN('INACTIVE','DEAD','ARCHIVE')"

	_, err = DBCONN.Exec(query)
	if err != nil {
		println(err.Error())
		return
	}

}
