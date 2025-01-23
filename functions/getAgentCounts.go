package functions

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/onetimenoob/VicidialHopper/models"
)

func GetAgentCounts(database *sql.DB) []models.AgentCount {
	//create a new instance of Agent_count and return it with values
	// return []models.AgentCount{{CampaignId: "VODAFUN2", AgentCount: 30}}

	query := "SELECT campaign_id, COUNT(*) as agent_count FROM vicidial_live_agents GROUP BY campaign_id"
	rows, err := database.Query(query)
	if err != nil {
		log.Println(err)
		return []models.AgentCount{}
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			println(err.Error())
		}
	}(rows)
	var agents []models.AgentCount
	for rows.Next() {
		var agent models.AgentCount
		err := rows.Scan(&agent.CampaignId, &agent.AgentCount)
		if err != nil {
			log.Println(err)
			return []models.AgentCount{}

		}
		fmt.Println(agent.CampaignId, agent.AgentCount)
		agents = append(agents, agent)
	}
	return agents
}
