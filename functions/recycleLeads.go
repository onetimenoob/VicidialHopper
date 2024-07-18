package functions

import (
	"database/sql"
	"fmt"
	"github.com/onetimenoob/VicidialHopper/models"
	"strconv"
	"strings"
)

type hopperLead struct {
	leadId         string
	phoneNumber    string
	listId         string
	vendorLeadCode string
	priority       int
}

func RecycleLeads(campaign_settings models.CampaignSettings, campaign models.Agent_count, DBconn *sql.DB, hopperLevelNeeded int, dncNumbers *models.DNCNumbers, dncNumbersCampaign *models.DNCNumbersCampaign) {

	campaignActiveLists := getCampaignActiveLists(DBconn, campaign.Campaign_id)
	if len(campaignActiveLists) == 0 {
		println("No active lists")
		return
	}
	var recycle_rules []models.Lead_recycle_rule
	fmt.Println("Need to add leads to hopper")
	recycle_rules = getRecycleRules(DBconn, campaign.Campaign_id)
	fmt.Println(recycle_rules)

	var newCount int
	switch {
	case strings.HasSuffix(campaign_settings.Lead_order, " 2nd NEW"):
		newCount = 2
	case strings.HasSuffix(campaign_settings.Lead_order, " 3rd NEW"):
		newCount = 3
	case strings.HasSuffix(campaign_settings.Lead_order, " 4th NEW"):
		newCount = 4
	case strings.HasSuffix(campaign_settings.Lead_order, " 5th NEW"):
		newCount = 5
	case strings.HasSuffix(campaign_settings.Lead_order, " 6th NEW"):
		newCount = 6
	}

	FinishedSqlQueryWhere := ""
	for _, rule := range recycle_rules {
		if newCount > 0 && rule.Status == "NEW" {
			println("New count is greater than 0 and rule status is NEW so skipping")
			continue
		}
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
		if newCount > 0 && status == "NEW" {
			println("New count is greater than 0 and status is NEW so skipping")
			continue
		}
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
		FinishedSqlQueryWhere = FinishedSqlQueryWhere + " AND " + campaign_settings.Lead_filter_id + " AND lead_id NOT IN (SELECT lead_id FROM vicidial_hopper)"
	}

	var orderStmt string
	switch {
	case strings.HasPrefix(campaign_settings.Lead_order, "UP LAST NAME"):
		orderStmt = "order by last_name desc, "
	case strings.HasPrefix(campaign_settings.Lead_order, "DOWN LAST NAME"):
		orderStmt = "order by last_name, "
	case strings.HasPrefix(campaign_settings.Lead_order, "UP PHONE"):
		orderStmt = "order by phone_number desc, "
	case strings.HasPrefix(campaign_settings.Lead_order, "DOWN PHONE"):
		orderStmt = "order by phone_number, "
	case strings.HasPrefix(campaign_settings.Lead_order, "UP COUNT"):
		orderStmt = "order by called_count desc, "
	case strings.HasPrefix(campaign_settings.Lead_order, "DOWN COUNT"):
		orderStmt = "order by called_count, "
	case strings.HasPrefix(campaign_settings.Lead_order, "UP LAST CALL TIME"):
		orderStmt = "order by last_local_call_time desc, "
	case strings.HasPrefix(campaign_settings.Lead_order, "DOWN LAST CALL TIME"):
		orderStmt = "order by last_local_call_time, "
	case strings.HasPrefix(campaign_settings.Lead_order, "UP RANK"):
		orderStmt = "order by rank desc, "
	case strings.HasPrefix(campaign_settings.Lead_order, "DOWN RANK"):
		orderStmt = "order by rank, "
	case strings.HasPrefix(campaign_settings.Lead_order, "UP OWNER"):
		orderStmt = "order by owner desc, "
	case strings.HasPrefix(campaign_settings.Lead_order, "DOWN OWNER"):
		orderStmt = "order by owner, "
	case strings.HasPrefix(campaign_settings.Lead_order, "UP TIMEZONE"):
		orderStmt = "order by gmt_offset_now desc, "
	case strings.HasPrefix(campaign_settings.Lead_order, "DOWN TIMEZONE"):
		orderStmt = "order by gmt_offset_now, "
	case strings.HasPrefix(campaign_settings.Lead_order, "RANDOM"):
		orderStmt = "order by RAND(), "
	case strings.HasPrefix(campaign_settings.Lead_order, "UP"):
		orderStmt = "order by lead_id desc, "
	case strings.HasPrefix(campaign_settings.Lead_order, "DOWN"):
		orderStmt = "order by lead_id asc, "
	}

	println(newCount)
	var secondary_order_stmt string
	switch campaign_settings.Lead_order_secondary {
	case "LEAD_ASCEND":
		secondary_order_stmt = "lead_id asc"
	case "LEAD_DESCEND":
		secondary_order_stmt = "lead_id desc"
	case "CALLTIME_ASCEND":
		secondary_order_stmt = "last_local_call_time asc"
	case "CALLTIME_DESCEND":
		secondary_order_stmt = "last_local_call_time desc"
	default:
		secondary_order_stmt = ""
	}

	FinishedSqlQueryWhere = FinishedSqlQueryWhere + " " + orderStmt + secondary_order_stmt

	// Wait for DNC numbers to be loaded before proceeding
	dncNumbers.Mu.RLock()
	dncNumbersCampaign.Mu.RLock()
	defer dncNumbers.Mu.RUnlock()
	defer dncNumbersCampaign.Mu.RUnlock()
	println(len(dncNumbers.Numbers))
	println(len(dncNumbersCampaign.Numbers))
	println(FinishedSqlQueryWhere)
	var query string
	query = "SELECT lead_id,phone_number,list_id,vendor_lead_code FROM vicidial_list WHERE" + FinishedSqlQueryWhere + " LIMIT " + strconv.Itoa(hopperLevelNeeded)
	rows, err := DBconn.Query(query)
	if err != nil {
		println(err)
		return
	}
	defer rows.Close()
	var recycleHopperLeads []hopperLead
	for rows.Next() {
		var lead_id string
		var phone_number string
		var list_id string
		var vendor_lead_code string
		var doNotLoadLead bool
		var dncStatus string
		err := rows.Scan(&lead_id, &phone_number, &list_id, &vendor_lead_code)
		if err != nil {
			println(err)
			return
		}
		// Check if the number is in DNC
		if dncNumbers.Numbers[phone_number] {
			println("Number is in DNC")
			dncStatus = "DNCL"
			doNotLoadLead = true
		}
		if dncNumbersCampaign.Numbers[phone_number] {
			println("Number is in campaign DNC")
			dncStatus = "DNCC"
			doNotLoadLead = true
		}
		// Add lead to hopper
		if doNotLoadLead {
			query := "update vicidial_list set status = ? where lead_id = ?"
			_, err := DBconn.Exec(query, dncStatus, lead_id)
			if err != nil {
				println(err)
			}
		} else {
			recycleHopperLeads = append(recycleHopperLeads, hopperLead{leadId: lead_id, phoneNumber: phone_number, listId: list_id, vendorLeadCode: vendor_lead_code, priority: 0})
		}
	}
	var newLeadHopper []hopperLead
	if newCount > 0 {
		query = "SELECT lead_id,phone_number,list_id,vendor_lead_code FROM vicidial_list WHERE status = 'NEW' AND list_id IN(" + activeListsString + ") AND lead_id NOT IN (SELECT lead_id FROM vicidial_hopper)" + " " + orderStmt + secondary_order_stmt + " LIMIT " + strconv.Itoa(hopperLevelNeeded)
		rows, err := DBconn.Query(query)
		if err != nil {
			println(err)
			return
		}
		defer rows.Close()
		for rows.Next() {
			var lead_id string
			var phone_number string
			var list_id string
			var vendor_lead_code string
			err := rows.Scan(&lead_id, &phone_number, &list_id, &vendor_lead_code)
			if err != nil {
				println(err)
				return
			}
			newLeadHopper = append(newLeadHopper, hopperLead{leadId: lead_id, phoneNumber: phone_number, listId: list_id, vendorLeadCode: vendor_lead_code, priority: 0})
		}
	}

	var hopperLeads []hopperLead
	//load callbacks first
	callbackStatuses := GetCallbackStatuses(campaign.Campaign_id, DBconn)
	query = "SELECT vicidial_list.lead_id,phone_number,vicidial_list.list_id,vendor_lead_code FROM vicidial_callbacks INNER JOIN  vicidial_list ON (vicidial_callbacks.lead_id=vicidial_list.lead_id) WHERE recipient='ANYONE' AND callback_time<NOW() AND vicidial_list.list_id IN(" + activeListsString + ") AND vicidial_list.lead_id NOT IN (SELECT lead_id FROM vicidial_hopper) and vicidial_list.status IN(" + callbackStatuses + ")"
	rows, err = DBconn.Query(query)
	if err != nil {
		println(err)
	}
	defer rows.Close()
	for rows.Next() {
		var lead_id string
		var phone_number string
		var list_id string
		var vendor_lead_code string
		err := rows.Scan(&lead_id, &phone_number, &list_id, &vendor_lead_code)
		if err != nil {
			println(err)
			return
		}
		fmt.Printf("Adding Callback lead %s to hopper\n", lead_id)
		hopperLeads = append(hopperLeads, hopperLead{leadId: lead_id, phoneNumber: phone_number, listId: list_id, vendorLeadCode: vendor_lead_code, priority: 50})
	}

	var recycleLeadCounter int = 1
	for i := 0; i < hopperLevelNeeded; i++ {
		if len(recycleHopperLeads) > 0 {
			lead := recycleHopperLeads[0]
			recycleHopperLeads = recycleHopperLeads[1:]
			hopperLeads = append(hopperLeads, lead)
			fmt.Printf("Adding Recycled lead %s to hopper\n", lead.leadId)
			recycleLeadCounter++
		}

		if (recycleLeadCounter == newCount && newCount > 0) || (len(recycleHopperLeads) == 0 && newCount > 0) {
			recycleLeadCounter = 1
			if len(newLeadHopper) > 0 {
				lead := newLeadHopper[0]
				newLeadHopper = newLeadHopper[1:]
				hopperLeads = append(hopperLeads, lead)
				fmt.Printf("Adding New lead %s to hopper\n", lead.leadId)
				i++
			}
		}
		if (len(recycleHopperLeads) == 0 && len(newLeadHopper) == 0) || len(hopperLeads) >= hopperLevelNeeded {
			break
		}
	}
	for _, lead := range hopperLeads {
		query := "INSERT INTO vicidial_hopper (lead_id, campaign_id, status, list_id, vendor_lead_code, priority) VALUES(?,?,?,?,?,?)"
		_, err := DBconn.Exec(query, lead.leadId, campaign.Campaign_id, "READY", lead.listId, lead.vendorLeadCode, lead.priority)
		if err != nil {
			println(err)
		}
	}
}
