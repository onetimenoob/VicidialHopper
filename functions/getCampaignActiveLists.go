package functions

import (
	"database/sql"
	"log"
)

func getCampaignActiveLists(database *sql.DB, campaign_id string) []string {
	query := "SELECT list_id FROM vicidial_lists WHERE campaign_id = ? and active='Y' and expiration_date > NOW()"
	rows, err := database.Query(query, campaign_id)
	if err != nil {
		log.Println(err)
		return []string{}
	}
	defer rows.Close()
	var lists []string
	for rows.Next() {
		var list string
		err := rows.Scan(&list)
		if err != nil {
			log.Println(err)
			return []string{}
		}
		lists = append(lists, list)
	}
	return lists
}
