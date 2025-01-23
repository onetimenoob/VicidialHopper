package functions

import (
	"database/sql"
	"strconv"

	"github.com/onetimenoob/VicidialHopper/models"
)

func UserOnlyToAnyoneCallbacks(campaignId string, DBConn *sql.DB, campaignSettings models.CampaignSettings) {
	query := "SELECT lead_id FROM vicidial_callbacks WHERE campaign_id='" + campaignId + "' AND recipient='USERONLY' AND callback_time<NOW() - INTERVAL " + strconv.Itoa(campaignSettings.CallbackUserOnlyMoveMinutes) + " MINUTE"
	rows, err := DBConn.Query(query)
	if err != nil {
		println(err)
	}
	defer rows.Close()
	for rows.Next() {
		var leadId string
		rows.Scan(&leadId)
		//println(leadId)
		query = "UPDATE vicidial_callbacks SET recipient='ANYONE' WHERE lead_id='" + leadId + "' AND campaign_id='" + campaignId + "' AND recipient='USERONLY'"
		_, err = DBConn.Exec(query)
		if err != nil {
			println(err)
		}
	}
}
