package functions

import (
	"database/sql"
	"fmt"
	"github.com/onetimenoob/VicidialHopper/models"
)

func CheckHopperDNC(campaign_id string, settings models.CampaignSettings, numbersCampaign *models.DNCNumbersCampaign, numbers *models.DNCNumbers, DBConn *sql.DB) {
	query := "select phone_number,vicidial_list.lead_id from vicidial_hopper inner join vicidial_list on (vicidial_hopper.lead_id=vicidial_list.lead_id) where campaign_id=?"
	rows, err := DBConn.Query(query, campaign_id)
	if err != nil {
		print("Error in query")
		println(err)
		return
	}
	println("Checking hopper for DNC numbers")
	defer rows.Close()
	numbers.Mu.RLock()
	defer numbers.Mu.RUnlock()
	numbersCampaign.Mu.RLock()
	defer numbersCampaign.Mu.RUnlock()
	for rows.Next() {
		var phone_number string
		var lead_id string
		err := rows.Scan(&phone_number, &lead_id)
		if err != nil {
			println(err)
			return
		}

		// Check if the number is in DNC
		if settings.Use_internal_dnc == "Y" {
			if numbers.Numbers[phone_number] {
				query := "update vicidial_list set status = ? where lead_id = ?"
				_, err := DBConn.Exec(query, "DNCL", lead_id)
				if err != nil {
					println(err)
				}
				query = "delete from vicidial_hopper where lead_id = ?"
				_, err = DBConn.Exec(query, lead_id)
				if err != nil {
					println(err)
				}
			}
		}
		if settings.Use_campaign_dnc == "Y" {
			if numbersCampaign.Numbers[phone_number] {
				query := "update vicidial_list set status = ? where lead_id = ?"
				_, err := DBConn.Exec(query, "DNCC", lead_id)
				if err != nil {
					println(err)
				}
				query = "delete from vicidial_hopper where lead_id = ?"
				_, err = DBConn.Exec(query, lead_id)
				if err != nil {
					println(err)
				}
			}
		}
	}
	fmt.Printf("Finished checking hopper for DNC numbers for campaign %s\n", campaign_id)
}
