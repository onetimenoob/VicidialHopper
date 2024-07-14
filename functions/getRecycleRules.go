package functions

import (
	"database/sql"
	"github.com/onetimenoob/VicidialHopper/models"
	"log"
)

func getRecycleRules(database *sql.DB, campaign_id string) []models.Lead_recycle_rule {
	query := "select status,attempt_delay,attempt_maximum from vicidial_lead_recycle where active='Y' and campaign_id = ?"
	rows, err := database.Query(query, campaign_id)
	if err != nil {
		log.Println(err)
		return []models.Lead_recycle_rule{}
	}
	defer rows.Close()
	var rules []models.Lead_recycle_rule
	for rows.Next() {
		var rule models.Lead_recycle_rule
		err := rows.Scan(&rule.Status, &rule.Attempt_delay, &rule.Max_attempts)
		if err != nil {
			log.Println(err)
			return []models.Lead_recycle_rule{}
		}
		rules = append(rules, rule)
	}
	return rules
}
