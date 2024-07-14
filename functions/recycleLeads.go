package functions

import (
	"database/sql"
	"fmt"
	"github.com/onetimenoob/VicidialHopper/models"
	"strconv"
	"strings"
)

func RecycleLeads(campaign_settings models.CampaignSettings, a models.Agent_count, DB *sql.DB) {

	campaignActiveLists := getCampaignActiveLists(DB, a.Campaign_id)
	if len(campaignActiveLists) == 0 {
		println("No active lists")
		return
	}
	var recycle_rules []models.Lead_recycle_rule
	fmt.Println("Need to add leads to hopper")
	recycle_rules = getRecycleRules(DB, a.Campaign_id)
	fmt.Println(recycle_rules)
	FinishedSqlQueryWhere := ""
	for _, rule := range recycle_rules {
		inDialStatus := false
		for i, status := range campaign_settings.Dial_statuses {
			if status == rule.Status {
				inDialStatus = true
				//remove the dial status from the list
				campaign_settings.Dial_statuses = append(campaign_settings.Dial_statuses[:i], campaign_settings.Dial_statuses[i+1:]...)
				break
			}
		}
		var lastCalledArray []string
		for attempts := 1; attempts <= rule.Max_attempts; attempts++ {
			lastCalledArray = append(lastCalledArray, "'Y"+fmt.Sprintf("%d'", attempts))
		}
		lastCalledString := strings.Join(lastCalledArray, ",")

		sqlQueryWhere := fmt.Sprintf(("last_local_call_time < DATE_SUB(NOW(), INTERVAL %d SECOND) and called_since_last_reset IN(%s)"), rule.Attempt_delay, lastCalledString)
		if FinishedSqlQueryWhere != "" {
			FinishedSqlQueryWhere = FinishedSqlQueryWhere + " OR "
		}
		if inDialStatus {
			FinishedSqlQueryWhere = FinishedSqlQueryWhere + fmt.Sprintf("(status = '%s' and ((%s) or (called_since_last_reset IN('N'))))", rule.Status, sqlQueryWhere)
		} else {
			FinishedSqlQueryWhere = FinishedSqlQueryWhere + fmt.Sprintf("(status = '%s' and (%s))", rule.Status, sqlQueryWhere)

		}

	}
	for _, status := range campaign_settings.Dial_statuses {
		if FinishedSqlQueryWhere != "" {
			FinishedSqlQueryWhere = FinishedSqlQueryWhere + " OR "
		}
		FinishedSqlQueryWhere = FinishedSqlQueryWhere + fmt.Sprintf("(status = '%s' and called_since_last_reset IN('N'))", status)
	}

	FinishedSqlQueryWhere = "(" + FinishedSqlQueryWhere + ")"
	if campaign_settings.Call_count_limit > 0 {
		FinishedSqlQueryWhere = FinishedSqlQueryWhere + " AND called_count < " + strconv.Itoa(campaign_settings.Call_count_limit)
	}
	activeListsString := strings.Join(campaignActiveLists, ",")
	FinishedSqlQueryWhere = FinishedSqlQueryWhere + " AND list_id IN(" + activeListsString + ")"

	if campaign_settings.Lead_filter_id != "" {
		FinishedSqlQueryWhere = FinishedSqlQueryWhere + " AND " + campaign_settings.Lead_filter_id
	}
	println(FinishedSqlQueryWhere)
}
