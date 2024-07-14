package functions

import (
	"database/sql"
	"github.com/onetimenoob/VicidialHopper/models"
	"log"
	"strings"
)

func GetCampaignSettings(database *sql.DB, campaign_id string) (models.CampaignSettings, error) {
	query := "SELECT lead_order,lead_filter_id,use_internal_dnc,use_campaign_dnc,dial_method,dial_timeout,dial_statuses,hopper_level,call_count_limit FROM vicidial_campaigns WHERE active='y' and campaign_id = ? and no_hopper_dialing='N' limit 1"
	rows, err := database.Query(query, campaign_id)
	if err != nil {
		log.Println(err)
		return models.CampaignSettings{}, err
	}
	defer rows.Close()
	var settings models.CampaignSettings
	var dial_status string
	var lead_filter_id string
	for rows.Next() {
		err := rows.Scan(&settings.Lead_order, &lead_filter_id, &settings.Use_internal_dnc, &settings.Use_campaign_dnc, &settings.Dial_method, &settings.Dial_timeout, &dial_status, &settings.Hopper_level, &settings.Call_count_limit)
		if err != nil {
			log.Println(err)
			return models.CampaignSettings{}, err
		}
	}
	// Split dial_statuses string into slice
	dial_statuses := strings.Split(dial_status, " ")
	for _, status := range dial_statuses {
		//remove leading and trailing spaces
		status = strings.TrimSpace(status)
		if status != "-" && status != "" && status != " " {
			println(status)
			settings.Dial_statuses = append(settings.Dial_statuses, status)
		}
	}
	if lead_filter_id != "NONE" {
		// Get lead filter details
		query = "SELECT lead_filter_sql FROM vicidial_lead_filters WHERE lead_filter_id = ?"
		rows, err := database.Query(query, lead_filter_id)
		if err != nil {
			log.Println(err)
		} else {
			defer rows.Close()
			for rows.Next() {
				var filter_sql string
				err := rows.Scan(&filter_sql)
				if err != nil {
					log.Println(err)
				}
				println(filter_sql)
				settings.Lead_filter_id = filter_sql
			}
		}

	}
	return settings, nil

}
