package models

type CampaignSettings struct {
	Lead_order       string
	Lead_filter_id   string
	Use_internal_dnc string
	Use_campaign_dnc string
	Dial_method      string
	Dial_timeout     int
	Dial_statuses    []string
	Hopper_level     int
	Call_count_limit int
}

type Lead_recycle_rule struct {
	Status        string
	Attempt_delay int
	Max_attempts  int
}

type Agent_count struct {
	Campaign_id string
	Agent_count int
}
