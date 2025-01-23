package models

import "sync"

type CampaignSettings struct {
	LeadOrder                   string
	LeadFilterId                string
	UseInternalDnc              string
	UseCampaignDnc              string
	DialMethod                  string
	DialTimeout                 int
	DialStatuses                []string
	HopperLevel                 int
	CallCountLimit              int
	LeadOrderSecondary          string
	AutoDialLevel               float32
	CallbackUserOnlyMoveMinutes int
}

type LeadRecycleRule struct {
	Status       string
	AttemptDelay int
	MaxAttempts  int
}

type AgentCount struct {
	CampaignId string
	AgentCount int
}

type DNCNumbers struct {
	Numbers map[string]bool
	Mu      sync.RWMutex
	Loaded  sync.WaitGroup
}

func (dnc *DNCNumbers) IsLoaded() bool {
	c := make(chan struct{})
	go func() {
		dnc.Loaded.Wait()
		close(c)
	}()
	select {
	case <-c:
		return true
	default:
		return false
	}
}

type DNCNumbersCampaign struct {
	Numbers map[string]bool
	Mu      sync.RWMutex
	Loaded  sync.WaitGroup
}

func (dnc *DNCNumbersCampaign) IsLoaded() bool {
	c := make(chan struct{})
	go func() {
		dnc.Loaded.Wait()
		close(c)
	}()
	select {
	case <-c:
		return true
	default:
		return false
	}
}
