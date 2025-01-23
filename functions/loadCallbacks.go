package functions

import (
	"database/sql"
	"fmt"
	"strings"
)

func LoadCallbacks(campaignId string, DBconn *sql.DB) {
	var hopperLeads []hopperLead
	//load callbacks first

	activeListsStr := strings.Join(getCampaignActiveLists(DBconn, campaignId), ",")
	callbackStatuses := GetCallbackStatuses(campaignId, DBconn)
	query := "SELECT vicidial_list.lead_id,phone_number,vicidial_list.list_id,vendor_lead_code FROM vicidial_callbacks INNER JOIN  vicidial_list ON (vicidial_callbacks.lead_id=vicidial_list.lead_id) WHERE recipient='ANYONE' AND callback_time<NOW() AND vicidial_list.list_id IN(" + activeListsStr + ") AND vicidial_list.lead_id NOT IN (SELECT lead_id FROM vicidial_hopper) and vicidial_list.status IN(" + callbackStatuses + ")"
	rows, err := DBconn.Query(query)
	if err != nil {
		println(err)
		return
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
			return
		}
		fmt.Printf("Adding Callback lead %s to hopper\n", leadId)
		hopperLeads = append(hopperLeads, hopperLead{leadId: leadId, phoneNumber: phoneNumber, listId: listId, vendorLeadCode: vendorLeadCode, priority: 50, source: "C"})
	}

	for _, lead := range hopperLeads {
		query := "INSERT INTO vicidial_hopper (lead_id, campaign_id, status, list_id, vendor_lead_code, priority,source) VALUES(?,?,?,?,?,?,?)"
		_, err := DBconn.Exec(query, lead.leadId, campaignId, "READY", lead.listId, lead.vendorLeadCode, lead.priority, lead.source)
		if err != nil {
			println(err)
		}
	}
}
