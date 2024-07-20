package functions

import (
	"database/sql"
	"log"
)

func getCampaignActiveLists(database *sql.DB, campaignId string) []string {
	query := "SELECT list_id FROM vicidial_lists WHERE campaign_id = ? and active='Y' and expiration_date > NOW()"
	rows, err := database.Query(query, campaignId)
	if err != nil {
		log.Println(err)
		return []string{}
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			println(err.Error())
		}
	}(rows)
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
