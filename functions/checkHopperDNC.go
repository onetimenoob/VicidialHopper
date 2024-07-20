package functions

import (
	"database/sql"
	"fmt"
	"github.com/onetimenoob/VicidialHopper/models"
)

func CheckHopperDNC(campaignId string, settings models.CampaignSettings, numbersCampaign *models.DNCNumbersCampaign, numbers *models.DNCNumbers, DBConn *sql.DB) {
	query := "select phone_number,vicidial_list.lead_id from vicidial_hopper inner join vicidial_list on (vicidial_hopper.lead_id=vicidial_list.lead_id) where campaign_id=?"
	rows, err := DBConn.Query(query, campaignId)
	if err != nil {
		print("Error in query")
		println(err)
		return
	}
	println("Checking hopper for DNC numbers")
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			println(err.Error())
		}
	}(rows)
	numbers.Mu.RLock()
	defer numbers.Mu.RUnlock()
	numbersCampaign.Mu.RLock()
	defer numbersCampaign.Mu.RUnlock()
	for rows.Next() {
		var phoneNumber string
		var leadId string
		err := rows.Scan(&phoneNumber, &leadId)
		if err != nil {
			println(err)
			return
		}

		// Check if the number is in DNC
		if settings.UseInternalDnc == "Y" {
			if numbers.Numbers[phoneNumber] {
				query := "update vicidial_list set status = ? where lead_id = ?"
				_, err := DBConn.Exec(query, "DNCL", leadId)
				if err != nil {
					println(err)
				}
				query = "delete from vicidial_hopper where lead_id = ?"
				_, err = DBConn.Exec(query, leadId)
				if err != nil {
					println(err)
				}
			}
		}
		if settings.UseCampaignDnc == "Y" {
			if numbersCampaign.Numbers[phoneNumber] {
				query := "update vicidial_list set status = ? where lead_id = ?"
				_, err := DBConn.Exec(query, "DNCC", leadId)
				if err != nil {
					println(err)
				}
				query = "delete from vicidial_hopper where lead_id = ?"
				_, err = DBConn.Exec(query, leadId)
				if err != nil {
					println(err)
				}
			}
		}
	}
	fmt.Printf("Finished checking hopper for DNC numbers for campaign %s\n", campaignId)
}
