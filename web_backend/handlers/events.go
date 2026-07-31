/*
 * Large-Scale Discovery, a network scanning solution for information gathering in large IT/OT network environments.
 *
 * Copyright (c) Siemens AG, 2016-2026.
 *
 * This work is licensed under the terms of the MIT license. For a copy, see the LICENSE file in the top-level
 * directory or visit <https://opensource.org/licenses/MIT>.
 *
 * SPDX-License-Identifier: MIT
 */

package handlers

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/siemens/Large-Scale-Discovery/web_backend/core"
	"github.com/siemens/Large-Scale-Discovery/web_backend/database"
)

// Events requests events of a certain type and returns either individual events or grouped data.
var Events = func() gin.HandlerFunc {

	// Define EventStats type
	type EventStats map[string]database.ResultSeries

	// Define expected request structure
	type requestBody struct {
		// - Avoid pointer types for mandatory request arguments, to prevent nil pointer panics.
		// - Use pointer types to represent optional request arguments. Pointer types allow modelling ternary states
		//   (e.g. not set, empty string, string), but need to be handled carefully to avoid nil pointer panics.
		Event database.Event `json:"event"`
		Since *time.Time     `json:"since"`
		Group bool           `json:"group"`
		Limit *int           `json:"limit"`
	}

	// Define expected response structure
	type responseBody struct {
		Result interface{} `json:"result"` // Can be []database.T_event or EventStats
	}

	// Return request handling function
	return func(ctx *gin.Context) {

		// Get logger for current request context
		logger := core.GetContextLogger(ctx)

		// Get user from context storage
		contextUser := core.GetContextUser(ctx)

		// Check if user has rights (is admin) to perform action
		if !contextUser.Admin {
			core.RespondAuthError(ctx)
			return
		}

		// Declare expected request struct
		var req requestBody

		// Decode JSON request into struct
		errReq := ctx.BindJSON(&req)
		if errReq != nil {
			logger.Errorf("Could not decode request: %s", errReq)
			core.RespondInternalError(ctx) // Return generic error information
			return
		}

		// Initialize standard time, if none was supplied
		if req.Since == nil {
			req.Since = &time.Time{}
		}

		// Query related events
		var eventEntries []database.T_event
		var errEventEntries error
		if req.Event == "*" {
			eventEntries, errEventEntries = database.GetEventsAll(*req.Since, req.Limit)
			if errEventEntries != nil {
				logger.Errorf("Could not query all events: %s", errEventEntries)
				core.RespondInternalError(ctx) // Return generic error information
				return
			}
		} else {
			eventEntries, errEventEntries = database.GetEvents(req.Event, *req.Since, req.Limit)
			if errEventEntries != nil {
				logger.Errorf("Could not query events: %s", errEventEntries)
				core.RespondInternalError(ctx) // Return generic error information
				return
			}
		}

		// Prepare response body
		var body responseBody

		// Process and group events
		if req.Group {
			stats := make(EventStats)
			for _, eventEntry := range eventEntries {
				var key string

				if eventEntry.User != nil {

					// Get department and company
					department := strings.ToUpper(strings.TrimSpace(eventEntry.User.Department))
					company := strings.ToUpper(strings.TrimSpace(eventEntry.User.Company))

					// Primary grouping: Use the first part of the department
					departmentParts := strings.Fields(department)
					if len(departmentParts) > 0 {
						key = departmentParts[0]
					} else {

						// Fallback: Use company name
						key = company
					}
				} else {
					key = "Unknown"
				}

				// Get the ISO week number
				year, week := eventEntry.Timestamp.ISOWeek()
				weekKey := fmt.Sprintf("%d-W%02d", year, week)

				// Initialize the key's data series if not present
				if _, exists := stats[key]; !exists {
					stats[key] = database.ResultSeries{}
				}

				// Check if the week is already in the key's data series
				found := false
				for i, result := range stats[key] {
					if result["week"] == weekKey {
						// Increment count
						stats[key][i]["count"] = result["count"].(int) + 1
						found = true
						break
					}
				}
				if !found {
					// Add new entry for this week
					stats[key] = append(stats[key], database.Result{
						"week":  weekKey,
						"count": 1,
					})
				}
			}

			// Group the stats
			stats = trimStatsGroups(stats)

			body = responseBody{
				Result: stats,
			}
		} else {
			// Return individual events
			body = responseBody{
				Result: eventEntries,
			}
		}

		// Return response
		core.Respond(ctx, false, "Events returned.", body)
	}
}

// trimStatsGroups reduces a map of results to the top 24 largest items
// and aggregates all subsequent ones into a final "others" item.
func trimStatsGroups(stats map[string]database.ResultSeries) map[string]database.ResultSeries {

	// Map to store total counts per key
	totalDataByKey := make(map[string]int)

	// Calculate total counts per key
	for key, values := range stats {
		total := 0
		for _, value := range values {
			count := value["count"].(int)
			total += count
		}
		totalDataByKey[key] = total
	}

	// Create a slice of key data for sorting
	type keyData struct {
		Key   string
		Total int
	}
	var keyList []keyData
	for key, total := range totalDataByKey {
		keyList = append(keyList, keyData{Key: key, Total: total})
	}

	// Sort keys by total count in descending order
	sort.Slice(keyList, func(i, j int) bool {
		return keyList[i].Total > keyList[j].Total
	})

	// Identify top 24 keys
	topKeys := make(map[string]bool)
	for i, data := range keyList {
		if i < 24 {
			topKeys[data.Key] = true
		} else {
			break
		}
	}

	// Group results
	grouped := make(map[string]database.ResultSeries)
	othersWeekCounts := make(map[string]int)

	for key, values := range stats {
		if topKeys[key] {
			grouped[key] = values
		} else {
			// Sum counts per week for "Others"
			for _, value := range values {
				week := value["week"].(string)
				count := value["count"].(int)
				othersWeekCounts[week] += count
			}
		}
	}

	// Combine "Others" data
	if len(othersWeekCounts) > 0 {
		othersSeries := database.ResultSeries{}
		for week, count := range othersWeekCounts {
			othersSeries = append(othersSeries, database.Result{
				"week":  week,
				"count": count,
			})
		}
		grouped["Others"] = othersSeries
	}

	// Return rearranged stat groups
	return grouped
}
