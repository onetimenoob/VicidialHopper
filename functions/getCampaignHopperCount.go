package functions

import (
	"database/sql"
	"log"
)

func GetCampaignHopperCount(campaign_id string, database *sql.DB) int {
	query := "SELECT COUNT(*) as lead_count FROM vicidial_hopper WHERE campaign_id = ?"
	rows, err := database.Query(query, campaign_id)
	if err != nil {
		log.Println(err)
		return 0
	}
	defer rows.Close()
	var lead_count int
	for rows.Next() {
		err := rows.Scan(&lead_count)
		if err != nil {
			log.Println(err)
			return 0
		}
	}
	return lead_count
}
