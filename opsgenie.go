package main

import (
	"context"
	"fmt"
	"strings"

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

var priorities = []string{"P1", "P2", "P3", "P4", "P5"}

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

// countAlerts accepts a search query to limit the results returned.
// See https://support.atlassian.com/opsgenie/docs/search-queries-for-alerts/
func (c *OpsgenieClient) countAlerts(ctx context.Context, query string) (float64, error) {
	res, err := c.alerts.CountAlerts(ctx, &alert.CountAlertsRequest{
		Query: query,
	})
	if err != nil {
		return 0, err
	}

	return float64(res.Count), nil
}

type countAlertsParams struct {
	Team     string
	Status   string
	Priority string
}

func (p *countAlertsParams) ToQuery() string {
	var filters []string
	if p.Team != "" {
		filters = append(filters, fmt.Sprintf("teams: %s", p.Team))
	}
	if p.Status != "" {
		filters = append(filters, fmt.Sprintf("status: %s", p.Team))
	}
	if p.Priority != "" {
		filters = append(filters, fmt.Sprintf("priority: %s", p.Priority))
	}

	if len(filters) > 0 {
		return strings.Join(filters, " AND ")
	}
	return ""
}

func (c *OpsgenieClient) CountAlerts() (float64, error) {
	return c.countAlerts(context.Background(), "")
}

func (c *OpsgenieClient) CountAlertsBy(params countAlertsParams) (float64, error) {
	return c.countAlerts(context.Background(), params.ToQuery())
}

func (c *OpsgenieClient) ListTeams() ([]team.ListedTeams, error) {
	res, err := c.teams.List(context.Background(), &team.ListTeamRequest{})
	if err != nil {
		return nil, err
	}
	return res.Teams, nil
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
