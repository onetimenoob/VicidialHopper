package main

import (
	"database/sql"
	"fmt"
	"github.com/onetimenoob/VicidialHopper/functions"
	"github.com/onetimenoob/VicidialHopper/models"
	"log"
	"sync"

	_ "github.com/go-sql-driver/mysql"
)

func makeDBConnection() *sql.DB {
	// Database connection parameters
	dsn := "cron:1234@tcp(10.20.30.40:3306)/asterisk?parseTime=true"
	DB, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
		return nil
	}
	err = DB.Ping()
	if err != nil {
		log.Fatal(err)
	}
	return DB
}

func get_callbacks(database *sql.DB) {
	// Implementation here
}

func main() {
	DBconn := makeDBConnection()
	if DBconn == nil {
		return
	}
	defer DBconn.Close()
	fmt.Println("Database connected successfully")

	campaigns := functions.GetAgentCounts(DBconn)
	var wg sync.WaitGroup

	for _, campaign := range campaigns {
		wg.Add(1)
		go func(campaign models.Agent_count) {
			defer wg.Done()
			var campaign_settings models.CampaignSettings

			campaign_settings, err := functions.GetCampaignSettings(DBconn, campaign.Campaign_id)
			if err != nil {
				log.Println(err)
				log.Println("Error in getting campaign settings")
				return
			}
			fmt.Println(campaign_settings)

			var hopperCurrentCount int
			hopperCurrentCount = functions.GetCampaignHopperCount(campaign.Campaign_id, DBconn)
			//Number of Active Agents * Auto Dial Level * ( 60 seconds / Dial Timeout ) * Auto Hopper
			var hopperLevelNeeded int
			if campaign_settings.Dial_method == "RATIO" {
				println("Ratio Dialing")
				hopperLevelNeeded = campaign.Agent_count * campaign_settings.Hopper_level * (60 / campaign_settings.Dial_timeout)
			} else {
				hopperLevelNeeded = campaign.Agent_count * 1 * (60 / campaign_settings.Dial_timeout)
				println("Predictive Dialing")
			}

			fmt.Printf("Hopper Level Needed: %d\n", hopperLevelNeeded)
			fmt.Printf("Hopper Current Count: %d\n", hopperCurrentCount)
			if hopperCurrentCount < hopperLevelNeeded {
				functions.RecycleLeads(campaign_settings, campaign, DBconn)
			}
			println("Done")

		}(campaign)
	}
	wg.Wait()
}
