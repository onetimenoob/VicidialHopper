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
	source         string
}

func RecycleLeads(campaignSettings models.CampaignSettings, campaign models.AgentCount, DBconn *sql.DB, hopperLevelNeeded int, dncNumbers *models.DNCNumbers, dncNumbersCampaign *models.DNCNumbersCampaign) {

	campaignActiveLists := getCampaignActiveLists(DBconn, campaign.CampaignId)

	if len(campaignActiveLists) == 0 {
		println("No active lists")
		return
	}
	var recycleRules []models.LeadRecycleRule

	recycleRules = getRecycleRules(DBconn, campaign.CampaignId)

	var newCount int
	switch {
	case strings.HasSuffix(campaignSettings.LeadOrder, " 2nd NEW"):
		newCount = 2
	case strings.HasSuffix(campaignSettings.LeadOrder, " 3rd NEW"):
		newCount = 3
	case strings.HasSuffix(campaignSettings.LeadOrder, " 4th NEW"):
		newCount = 4
	case strings.HasSuffix(campaignSettings.LeadOrder, " 5th NEW"):
		newCount = 5
	case strings.HasSuffix(campaignSettings.LeadOrder, " 6th NEW"):
		newCount = 6
	}

	var FinishedSqlQueryWhere string
	FinishedSqlQueryWhere, activeListsString := recycleRuleBuilder(campaignSettings, recycleRules, newCount, FinishedSqlQueryWhere, campaignActiveLists)

	orderStmt := getSortParams(campaignSettings)

	println(newCount)

	FinishedSqlQueryWhere = FinishedSqlQueryWhere + " " + orderStmt

	// Wait for DNC numbers to be loaded before proceeding
	dncNumbers.Mu.RLock()
	dncNumbersCampaign.Mu.RLock()
	defer dncNumbers.Mu.RUnlock()
	defer dncNumbersCampaign.Mu.RUnlock()

	//get Recycled Leads
	var recycleHopperLeads []hopperLead
	recycleHopperLeads, done := getRecycleLeads(dncNumbers, dncNumbersCampaign, DBconn, recycleHopperLeads, FinishedSqlQueryWhere, hopperLevelNeeded)
	if done {
		return
	}

	//get New Leads
	var newLeadHopper []hopperLead
	newLeadHopper, done2 := getNewLeads(campaignSettings, newCount, activeListsString, orderStmt, hopperLevelNeeded, DBconn, newLeadHopper)
	if done2 {
		return
	}

	//build list with recycled and new leads
	var hopperLeads []hopperLead
	hopperLeads = buildPreHopperLoadList(hopperLevelNeeded, recycleHopperLeads, hopperLeads, newCount, newLeadHopper)

	//load leads into vicidial hopper
	loadLeadsIntoHopper(hopperLeads, DBconn, campaign)
}

func loadLeadsIntoHopper(hopperLeads []hopperLead, DBconn *sql.DB, campaign models.AgentCount) {
	for _, lead := range hopperLeads {
		query := "INSERT INTO vicidial_hopper (lead_id, campaign_id, status, list_id, vendor_lead_code, priority,source) VALUES(?,?,?,?,?,?,?)"
		_, err := DBconn.Exec(query, lead.leadId, campaign.CampaignId, "READY", lead.listId, lead.vendorLeadCode, lead.priority, lead.source)
		if err != nil {
			println(err)
		}
	}
}

func buildPreHopperLoadList(hopperLevelNeeded int, recycleHopperLeads []hopperLead, hopperLeads []hopperLead, newCount int, newLeadHopper []hopperLead) []hopperLead {
	var recycleLeadCounter = 1
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
	return hopperLeads
}

func getNewLeads(campaignSettings models.CampaignSettings, newCount int, activeListsString string, orderStmt string, hopperLevelNeeded int, DBconn *sql.DB, newLeadHopper []hopperLead) ([]hopperLead, bool) {
	var query string
	if newCount > 0 {
		if campaignSettings.LeadFilterId != "" {
			query = "SELECT lead_id,phone_number,list_id,vendor_lead_code FROM vicidial_list WHERE status = 'NEW' AND list_id IN(" + activeListsString + ") AND lead_id NOT IN (SELECT lead_id FROM vicidial_hopper)" + " AND " + campaignSettings.LeadFilterId + " " + orderStmt + " LIMIT " + strconv.Itoa(hopperLevelNeeded)
		} else {
			query = "SELECT lead_id,phone_number,list_id,vendor_lead_code FROM vicidial_list WHERE status = 'NEW' AND list_id IN(" + activeListsString + ") AND lead_id NOT IN (SELECT lead_id FROM vicidial_hopper)" + " " + orderStmt + " LIMIT " + strconv.Itoa(hopperLevelNeeded)
		}
		rows, err := DBconn.Query(query)
		if err != nil {
			println(err)
			return nil, true
		}
		defer func(rows *sql.Rows) {
			err := rows.Close()
			if err != nil {
				println(err.Error())
			}
		}(rows)
		for rows.Next() {
			var leadId string
			var phoneNumber string
			var listId string
			var vendorLeadCode string
			err := rows.Scan(&leadId, &phoneNumber, &listId, &vendorLeadCode)
			if err != nil {
				println(err)
				return nil, true
			}
			newLeadHopper = append(newLeadHopper, hopperLead{leadId: leadId, phoneNumber: phoneNumber, listId: listId, vendorLeadCode: vendorLeadCode, priority: 0, source: "R"})
		}
	}
	return newLeadHopper, false
}

func recycleRuleBuilder(campaignSettings models.CampaignSettings, recycleRules []models.LeadRecycleRule, newCount int, FinishedSqlQueryWhere string, campaignActiveLists []string) (string, string) {
	for _, rule := range recycleRules {
		if newCount > 0 && rule.Status == "NEW" {
			println("New count is greater than 0 and rule status is NEW so skipping")
			continue
		}
		inDialStatus := false
		for i, status := range campaignSettings.DialStatuses {
			if status == rule.Status {
				inDialStatus = true
				//remove the dial status from the list
				campaignSettings.DialStatuses = append(campaignSettings.DialStatuses[:i], campaignSettings.DialStatuses[i+1:]...)
				break
			}
		}
		var lastCalledArray []string
		for attempts := 0; attempts <= rule.MaxAttempts; attempts++ {
			if attempts == 0 {
				lastCalledArray = append(lastCalledArray, "'Y'")
			} else {
				lastCalledArray = append(lastCalledArray, "'Y"+fmt.Sprintf("%d'", attempts))
			}
		}
		lastCalledString := strings.Join(lastCalledArray, ",")

		sqlQueryWhere := fmt.Sprintf(("last_local_call_time < DATE_SUB(NOW(), INTERVAL %d SECOND) and called_since_last_reset IN(%s)"), rule.AttemptDelay, lastCalledString)
		if FinishedSqlQueryWhere != "" {
			FinishedSqlQueryWhere = FinishedSqlQueryWhere + " OR "
		}
		if inDialStatus {
			FinishedSqlQueryWhere = FinishedSqlQueryWhere + fmt.Sprintf("(status = '%s' and ((%s) or (called_since_last_reset IN('N'))))", rule.Status, sqlQueryWhere)
		} else {
			FinishedSqlQueryWhere = FinishedSqlQueryWhere + fmt.Sprintf("(status = '%s' and (%s))", rule.Status, sqlQueryWhere)
		}

	}
	for _, status := range campaignSettings.DialStatuses {
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
	if campaignSettings.CallCountLimit > 0 {
		FinishedSqlQueryWhere = FinishedSqlQueryWhere + " AND called_count < " + strconv.Itoa(campaignSettings.CallCountLimit)
	}
	activeListsString := strings.Join(campaignActiveLists, ",")
	FinishedSqlQueryWhere = FinishedSqlQueryWhere + " AND list_id IN(" + activeListsString + ")"

	if campaignSettings.LeadFilterId != "" {
		FinishedSqlQueryWhere = FinishedSqlQueryWhere + " AND " + campaignSettings.LeadFilterId
	}
	FinishedSqlQueryWhere = FinishedSqlQueryWhere + " AND lead_id NOT IN (SELECT lead_id FROM vicidial_hopper)"
	return FinishedSqlQueryWhere, activeListsString
}

func getRecycleLeads(dncNumbers *models.DNCNumbers, dncNumbersCampaign *models.DNCNumbersCampaign, DBconn *sql.DB, recycleHopperLeads []hopperLead, FinishedSqlQueryWhere string, hopperLevelNeeded int) ([]hopperLead, bool) {
	var query string
	query = "SELECT lead_id,phone_number,list_id,vendor_lead_code FROM vicidial_list WHERE" + FinishedSqlQueryWhere + " LIMIT " + strconv.Itoa(hopperLevelNeeded)
	println(query)
	rows, err := DBconn.Query(query)
	if err != nil {
		println(err)
		return nil, true
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			println(err.Error())
		}
	}(rows)

	for rows.Next() {
		var leadId string
		var phoneNumber string
		var listId string
		var vendorLeadCode string
		var doNotLoadLead bool
		var dncStatus string
		err := rows.Scan(&leadId, &phoneNumber, &listId, &vendorLeadCode)
		if err != nil {
			println(err)
			return nil, true
		}
		// Check if the number is in DNC
		if dncNumbers.Numbers[phoneNumber] {
			println("Number is in DNC")
			dncStatus = "DNCL"
			doNotLoadLead = true
		}
		if dncNumbersCampaign.Numbers[phoneNumber] {
			println("Number is in campaign DNC")
			dncStatus = "DNCC"
			doNotLoadLead = true
		}
		// Add lead to hopper
		if doNotLoadLead {
			query := "update vicidial_list set status = ? where lead_id = ?"
			_, err := DBconn.Exec(query, dncStatus, leadId)
			if err != nil {
				println(err)
			}
		} else {
			recycleHopperLeads = append(recycleHopperLeads, hopperLead{leadId: leadId, phoneNumber: phoneNumber, listId: listId, vendorLeadCode: vendorLeadCode, priority: 0, source: "R"})
		}
	}
	return recycleHopperLeads, false
}

func getSortParams(campaignSettings models.CampaignSettings) string {
	var orderStmt, secondaryOrderStmt string
	switch {
	case strings.HasPrefix(campaignSettings.LeadOrder, "UP LAST NAME"):
		orderStmt = "order by last_name desc, "
	case strings.HasPrefix(campaignSettings.LeadOrder, "DOWN LAST NAME"):
		orderStmt = "order by last_name, "
	case strings.HasPrefix(campaignSettings.LeadOrder, "UP PHONE"):
		orderStmt = "order by phone_number desc, "
	case strings.HasPrefix(campaignSettings.LeadOrder, "DOWN PHONE"):
		orderStmt = "order by phone_number, "
	case strings.HasPrefix(campaignSettings.LeadOrder, "UP COUNT"):
		orderStmt = "order by called_count desc, "
	case strings.HasPrefix(campaignSettings.LeadOrder, "DOWN COUNT"):
		orderStmt = "order by called_count, "
	case strings.HasPrefix(campaignSettings.LeadOrder, "UP LAST CALL TIME"):
		orderStmt = "order by last_local_call_time desc, "
	case strings.HasPrefix(campaignSettings.LeadOrder, "DOWN LAST CALL TIME"):
		orderStmt = "order by last_local_call_time, "
	case strings.HasPrefix(campaignSettings.LeadOrder, "UP RANK"):
		orderStmt = "order by rank desc, "
	case strings.HasPrefix(campaignSettings.LeadOrder, "DOWN RANK"):
		orderStmt = "order by rank, "
	case strings.HasPrefix(campaignSettings.LeadOrder, "UP OWNER"):
		orderStmt = "order by owner desc, "
	case strings.HasPrefix(campaignSettings.LeadOrder, "DOWN OWNER"):
		orderStmt = "order by owner, "
	case strings.HasPrefix(campaignSettings.LeadOrder, "UP TIMEZONE"):
		orderStmt = "order by gmt_offset_now desc, "
	case strings.HasPrefix(campaignSettings.LeadOrder, "DOWN TIMEZONE"):
		orderStmt = "order by gmt_offset_now, "
	case strings.HasPrefix(campaignSettings.LeadOrder, "RANDOM"):
		orderStmt = "order by RAND(), "
	case strings.HasPrefix(campaignSettings.LeadOrder, "UP"):
		orderStmt = "order by lead_id desc, "
	case strings.HasPrefix(campaignSettings.LeadOrder, "DOWN"):
		orderStmt = "order by lead_id asc, "
	}
	switch campaignSettings.LeadOrderSecondary {
	case "LEAD_ASCEND":
		secondaryOrderStmt = "lead_id asc"
	case "LEAD_DESCEND":
		secondaryOrderStmt = "lead_id desc"
	case "CALLTIME_ASCEND":
		secondaryOrderStmt = "last_local_call_time asc"
	case "CALLTIME_DESCEND":
		secondaryOrderStmt = "last_local_call_time desc"
	default:
		secondaryOrderStmt = ""
	}

	return orderStmt + " " + secondaryOrderStmt
}
