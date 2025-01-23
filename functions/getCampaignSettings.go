package functions

import (
	"database/sql"
	"log"
	"strings"

	"github.com/onetimenoob/VicidialHopper/models"
)

func GetCampaignSettings(database *sql.DB, campaignId string) (models.CampaignSettings, error) {
	query := "SELECT lead_order,lead_filter_id,use_internal_dnc,use_campaign_dnc,dial_method,dial_timeout,dial_statuses,hopper_level,call_count_limit,lead_order_secondary,auto_dial_level,callback_useronly_move_minutes FROM vicidial_campaigns WHERE active='y' and campaign_id = ? and no_hopper_dialing='N' limit 1"
	rows, err := database.Query(query, campaignId)
	if err != nil {
		log.Println(err)
		return models.CampaignSettings{}, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			println(err.Error())
		}
	}(rows)
	var settings models.CampaignSettings
	var dialStatus string
	var leadFilterId string
	for rows.Next() {
		err := rows.Scan(&settings.LeadOrder, &leadFilterId, &settings.UseInternalDnc, &settings.UseCampaignDnc, &settings.DialMethod, &settings.DialTimeout, &dialStatus, &settings.HopperLevel, &settings.CallCountLimit, &settings.LeadOrderSecondary, &settings.AutoDialLevel, &settings.CallbackUserOnlyMoveMinutes)
		if err != nil {
			log.Println(err)
			return models.CampaignSettings{}, err
		}
	}
	// Split dial_statuses string into slice
	dialStatuses := strings.Split(dialStatus, " ")
	for _, status := range dialStatuses {
		//remove leading and trailing spaces
		status = strings.TrimSpace(status)
		if status != "-" && status != "" && status != " " {
			println(status)
			settings.DialStatuses = append(settings.DialStatuses, status)
		}
	}
	if leadFilterId != "NONE" {
		// Get lead filter details
		query = "SELECT lead_filter_sql FROM vicidial_lead_filters WHERE lead_filter_id = ?"
		rows, err := database.Query(query, leadFilterId)
		if err != nil {
			log.Println(err)
		} else {
			defer func(rows *sql.Rows) {
				err := rows.Close()
				if err != nil {
					println(err.Error())
				}
			}(rows)
			for rows.Next() {
				var filterSql string
				err := rows.Scan(&filterSql)
				if err != nil {
					log.Println(err)
				}
				println(filterSql)
				settings.LeadFilterId = filterSql
			}
		}

	}
	return settings, nil

}
