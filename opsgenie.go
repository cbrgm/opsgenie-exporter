package main

import (
	"context"
	"fmt"
	"github.com/opsgenie/opsgenie-go-sdk-v2/alert"
	"github.com/opsgenie/opsgenie-go-sdk-v2/client"
	"github.com/opsgenie/opsgenie-go-sdk-v2/team"
	"github.com/opsgenie/opsgenie-go-sdk-v2/user"
)

type OpsgenieClient struct {
	alerts *alert.Client
	teams  *team.Client
	users  *user.Client
}

const (
	alertOpenStatus   = "open"
	alertClosedStatus = "closed"
)

func NewOpsgenieClient(apikey string) (*OpsgenieClient, error) {
	alertClient, err := alert.NewClient(&client.Config{ApiKey: apikey})
	if err != nil {
		return nil, err
	}

	teamClient, err := team.NewClient(&client.Config{ApiKey: apikey})
	if err != nil {
		return nil, err
	}

	userClient, err := user.NewClient(&client.Config{ApiKey: apikey})
	if err != nil {
		return nil, err
	}

	return &OpsgenieClient{
		alerts: alertClient,
		teams:  teamClient,
		users:  userClient,
	}, nil
}

func (c *OpsgenieClient) CountAlerts() (float64, error) {
	res, err := c.alerts.CountAlerts(context.Background(), &alert.CountAlertsRequest{
		Query: "",
	})
	if err != nil {
		return 0, err
	}
	return float64(res.Count), nil
}

func (c *OpsgenieClient) CountClosedAlerts() (float64, error) {
	res, err := c.alerts.CountAlerts(context.Background(), &alert.CountAlertsRequest{
		Query: fmt.Sprintf("status: %s", alertClosedStatus),
	})
	if err != nil {
		return 0, err
	}
	return float64(res.Count), nil
}

func (c *OpsgenieClient) CountOpenAlerts() (float64, error) {
	res, err := c.alerts.CountAlerts(context.Background(), &alert.CountAlertsRequest{
		Query: fmt.Sprintf("status: %s", alertOpenStatus),
	})
	if err != nil {
		return 0, err
	}
	return float64(res.Count), nil
}

func (c *OpsgenieClient) CountTeams() (float64, error) {
	res, err := c.teams.List(context.Background(), &team.ListTeamRequest{})
	if err != nil {
		return 0, err
	}
	return float64(len(res.Teams)), nil
}

func (c *OpsgenieClient) CountUsersByRole() (map[string]float64, error) {
	res, err := c.users.List(context.Background(), &user.ListRequest{})
	if err != nil {
		return make(map[string]float64), err
	}

	count := make(map[string]float64)
	for _, user := range res.Users {
		count[user.Role.RoleName] = count[user.Role.RoleName] + 1
	}
	return count, nil
}
