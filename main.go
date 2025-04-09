package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/dolow/mcp_sandbox/cluster_public"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type ClusterEventResponse struct {
	Events []ClusterEvent `json:"events"`
}

type ClusterEvent struct {
	Status    string `json:"status"`
	Url       string `json:"url"`
	Title     string `json:"title"`
	OwnerName string `json:"ownerName"`
	OwnerBio  string `json:"ownerBio"`
	OpenAt    string `json:"openAt"`
	EndAt     string `json:"endAt"`
	PageView  int    `json:"pageView"`
}

func clusterEventUrlFromId(id string) string {
	return fmt.Sprintf("https://cluster.mu/e/%s", id)
}

func reduceClusterEvents(eventSets ...[]cluster_public.Event) []ClusterEvent {
	allEventLength := 0
	for _, events := range eventSets {
		allEventLength = allEventLength + len(events)
	}

	clusterEvents := make([]ClusterEvent, allEventLength)

	i := 0
	for _, events := range eventSets {
		for _, event := range events {
			s := event.Summary
			clusterEvents[i] = ClusterEvent{
				Status:    s.EventStatus,
				Url:       clusterEventUrlFromId(s.ID),
				Title:     s.Name,
				OwnerName: s.Owner.DisplayName,
				OwnerBio:  s.Owner.Bio,
				OpenAt:    s.Reservation.OpenDatetime.Format("2006-01-02T15:04:05Z07:00"),
				EndAt:     s.Reservation.CloseDatetime.Format("2006-01-02T15:04:05Z07:00"),
				PageView:  s.WatchCount,
			}
			i = i + 1
		}
	}

	return clusterEvents
}

func main() {
	// Create a new MCP server

	hooks := &server.Hooks{}
	hooks.AddBeforeAny(func(ctx context.Context, id any, method mcp.MCPMethod, message any) {
		fmt.Printf("beforeAny: %s, %v, %v\n", method, id, message)
	})

	s := server.NewMCPServer(
		"Cluster Fetcher",
		"1.0.0",
		server.WithResourceCapabilities(true, true),
		server.WithLogging(),
		server.WithHooks(hooks),
	)

	clusterEventFetchTool := mcp.NewTool("Cluster event fetcher",
		mcp.WithDescription("Fetch events from cluster.mu"),
		mcp.WithString("event_type",
			mcp.Required(),
			mcp.Description("Type of events to fetch"),
			mcp.Enum("all", "featured", "open soon"),
		),
	)

	s.AddTool(clusterEventFetchTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		typ := request.Params.Arguments["event_type"].(string)

		switch typ {
		case "all":
			featured, err := cluster_public.GetFeaturedEvents()
			if err != nil {
				return mcp.NewToolResultError("failed to fetch events from cluster"), err
			}
			openSoon, err := cluster_public.GetInPreparationEvents()
			if err != nil {
				return mcp.NewToolResultError("failed to fetch events from cluster"), err
			}

			events := reduceClusterEvents(featured.Events, openSoon.Events)
			jsonStr, err := json.Marshal(events)
			if err != nil {
				return mcp.NewToolResultError("failed to parse eventsfrom cluster"), err
			}

			return mcp.NewToolResultText(string(jsonStr)), nil
		case "featured":
			featured, err := cluster_public.GetFeaturedEvents()
			if err != nil {
				return mcp.NewToolResultError("failed to fetch events from cluster"), err
			}

			events := reduceClusterEvents(featured.Events)
			jsonStr, err := json.Marshal(events)
			if err != nil {
				return mcp.NewToolResultError("failed to parse eventsfrom cluster"), err
			}

			return mcp.NewToolResultText(string(jsonStr)), nil
		case "open soon":
			openSoon, err := cluster_public.GetInPreparationEvents()
			if err != nil {
				return mcp.NewToolResultError("failed to fetch events from cluster"), err
			}

			events := reduceClusterEvents(openSoon.Events)
			jsonStr, err := json.Marshal(events)
			if err != nil {
				return mcp.NewToolResultError("failed to parse eventsfrom cluster"), err
			}

			return mcp.NewToolResultText(string(jsonStr)), nil
		}

		return mcp.NewToolResultError("invalid event type"), nil
	})

	port := 8082
	baseUrl := fmt.Sprintf("http://localhost:%d", port)

	sse := server.NewSSEServer(s, server.WithBaseURL(baseUrl))
	fmt.Printf("Server launched with base url: %s\n", baseUrl)
	if err := sse.Start(fmt.Sprintf(":%d", port)); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
