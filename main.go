package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/onetimenoob/VicidialHopper/functions"
	"github.com/onetimenoob/VicidialHopper/models"
	"log"
	"sync"
	"time"
)

func NewDNCNumbers() *models.DNCNumbers {
	dnc := &models.DNCNumbers{
		Numbers: make(map[string]bool),
	}
	dnc.Loaded.Add(1)
	return dnc
}

func NewDNCNumbersCampaigns() *models.DNCNumbersCampaign {
	dnc := &models.DNCNumbersCampaign{
		Numbers: make(map[string]bool),
	}
	dnc.Loaded.Add(1)
	return dnc
}

func makeDBConnection() *sql.DB {
	// Database connection parameters
	dsn := "cron:1234@tcp(10.0.0.92:3306)/asterisk?parseTime=true"
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

func LoadDNCNumbers(database *sql.DB, numbers *models.DNCNumbers) {
	defer numbers.Loaded.Done()

	numbers.Mu.Lock()
	defer numbers.Mu.Unlock()

	query := "SELECT phone_number FROM vicidial_dnc"
	rows, err := database.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			println(err.Error())
		}
	}(rows)

	for rows.Next() {
		var phoneNumber string
		err := rows.Scan(&phoneNumber)
		if err != nil {
			log.Fatal(err)
		}
		numbers.Numbers[phoneNumber] = true
	}
	println("DNC Numbers Loaded!!")
}

func LoadDNCNumbersCampaign(database *sql.DB, numbers *models.DNCNumbersCampaign, campaignId string) {
	defer numbers.Loaded.Done()

	numbers.Mu.Lock()
	defer numbers.Mu.Unlock()

	query := "SELECT phone_number FROM vicidial_campaign_dnc where campaign_id = ?"
	rows, err := database.Query(query, campaignId)
	if err != nil {
		log.Fatal(err)
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			println(err.Error())
		}
	}(rows)

	for rows.Next() {
		var phoneNumber string
		err := rows.Scan(&phoneNumber)
		if err != nil {
			log.Fatal(err)
		}
		numbers.Numbers[phoneNumber] = true
	}
	println("DNC Numbers Loaded!!")
}

func main() {
	start := time.Now()
	DBconn := makeDBConnection()
	if DBconn == nil {
		return
	}
	defer func(DBconn *sql.DB) {
		err := DBconn.Close()
		if err != nil {
			println(err.Error())
		}
	}(DBconn)
	fmt.Println("Database connected successfully")

	dncNumbers := NewDNCNumbers()
	go LoadDNCNumbers(DBconn, dncNumbers)

	campaigns := functions.GetAgentCounts(DBconn)
	var wg sync.WaitGroup
	// Remove Duplicate callbacks only keep latest 1
	functions.RemoveDuplicateCallbacks(DBconn)
	for _, campaign := range campaigns {
		println("Campaign ID: ", campaign.CampaignId)
		wg.Add(1)
		go func(campaign models.AgentCount) {
			defer wg.Done()
			dncNumbersCampaign := NewDNCNumbersCampaigns()
			go LoadDNCNumbersCampaign(DBconn, dncNumbersCampaign, campaign.CampaignId)
			functions.RemoveNonValidCallbacks(campaign.CampaignId, DBconn)
			campaignSettings, err := functions.GetCampaignSettings(DBconn, campaign.CampaignId)
			if err != nil {
				log.Println(err)
				log.Println("Error in getting campaign settings")
				return
			}
			fmt.Println(campaignSettings)
			if campaignSettings.UseInternalDnc == "Y" || campaignSettings.UseCampaignDnc == "Y" {
				functions.CheckHopperDNC(campaign.CampaignId, campaignSettings, dncNumbersCampaign, dncNumbers, DBconn)
			}
			hopperCurrentCount := functions.GetCampaignHopperCount(campaign.CampaignId, DBconn)

			var calcHopperLevel float32
			var hopperLevelNeeded int
			if campaignSettings.DialMethod == "RATIO" {
				println("Ratio Dialing")
				calcHopperLevel = float32(campaign.AgentCount) * float32(campaignSettings.AutoDialLevel) * (60.0 / float32(campaignSettings.DialTimeout))
			} else {
				calcHopperLevel = float32(campaign.AgentCount) * 1 * (60.0 / float32(campaignSettings.DialTimeout))
				println("Predictive Dialing")
			}
			println(calcHopperLevel)
			hopperLevelNeeded = int(calcHopperLevel)
			fmt.Printf("Hopper Level Needed: %d\n", hopperLevelNeeded)
			fmt.Printf("Hopper Current Count: %d\n", hopperCurrentCount)

			if hopperCurrentCount < hopperLevelNeeded {
				// Wait for DNC numbers to be loaded before proceeding

				functions.RecycleLeads(campaignSettings, campaign, DBconn, hopperLevelNeeded, dncNumbers, dncNumbersCampaign)
			}
			println("Done")
		}(campaign)
	}
	wg.Wait()
	fmt.Printf("Time taken: %s\n", time.Since(start))
}
