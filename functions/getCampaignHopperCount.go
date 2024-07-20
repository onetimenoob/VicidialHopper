package functions

import (
	"database/sql"
	"log"
)

func GetCampaignHopperCount(campaignId string, database *sql.DB) int {
	query := "SELECT COUNT(*) as lead_count FROM vicidial_hopper WHERE campaign_id = ?"
	rows, err := database.Query(query, campaignId)
	if err != nil {
		log.Println(err)
		return 0
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			println(err.Error())
		}
	}(rows)
	var leadCount int
	for rows.Next() {
		err := rows.Scan(&leadCount)
		if err != nil {
			log.Println(err)
			return 0
		}
	}
	return leadCount
}
