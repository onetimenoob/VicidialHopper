package functions

import (
	"database/sql"
	"fmt"
	"github.com/onetimenoob/VicidialHopper/models"
	"log"
)

func GetAgentCounts(database *sql.DB) []models.Agent_count {
	// create a new instance of Agent_count and return it with values
	//return []models.Agent_count{{Campaign_id: "VODAFUN2", Agent_count: 210}}

	query := "SELECT campaign_id, COUNT(*) as agent_count FROM vicidial_live_agents GROUP BY campaign_id"
	rows, err := database.Query(query)
	if err != nil {
		log.Println(err)
		return []models.Agent_count{}
	}
	defer rows.Close()
	var agents []models.Agent_count
	for rows.Next() {
		var agent models.Agent_count
		err := rows.Scan(&agent.Campaign_id, &agent.Agent_count)
		if err != nil {
			log.Println(err)
			return []models.Agent_count{}

		}
		fmt.Println(agent.Campaign_id, agent.Agent_count)
		agents = append(agents, agent)
	}
	return agents
}
