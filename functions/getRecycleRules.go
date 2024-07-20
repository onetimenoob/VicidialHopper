package functions

import (
	"database/sql"
	"github.com/onetimenoob/VicidialHopper/models"
	"log"
)

func getRecycleRules(database *sql.DB, campaignId string) []models.LeadRecycleRule {
	query := "select status,attempt_delay,attempt_maximum from vicidial_lead_recycle where active='Y' and campaign_id = ?"
	rows, err := database.Query(query, campaignId)
	if err != nil {
		log.Println(err)
		return []models.LeadRecycleRule{}
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			println(err.Error())
		}
	}(rows)
	var rules []models.LeadRecycleRule
	for rows.Next() {
		var rule models.LeadRecycleRule
		err := rows.Scan(&rule.Status, &rule.AttemptDelay, &rule.MaxAttempts)
		if err != nil {
			log.Println(err)
			return []models.LeadRecycleRule{}
		}
		rules = append(rules, rule)
	}
	return rules
}
