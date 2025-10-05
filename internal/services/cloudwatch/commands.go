package cloudwatch

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	tea "github.com/charmbracelet/bubbletea"
)

type logGroupsLoadedMsg struct {
	groups []types.LogGroup
	err    error
}

type logEventsLoadedMsg struct {
	events []types.FilteredLogEvent
	err    error
}

type filteredLogsMsg struct {
	output string
	err    error
}

func (m Model) filterWithRipgrep(pattern string) tea.Cmd {
	return func() tea.Msg {
		// Check if ripgrep is available
		cmd := exec.Command("rg", "--color", "never", pattern)
		cmd.Stdin = strings.NewReader(m.allLogsText)

		var out bytes.Buffer
		cmd.Stdout = &out

		if err := cmd.Run(); err != nil {
			// If ripgrep not found or no matches, return error
			if _, ok := err.(*exec.ExitError); ok {
				// Exit code 1 means no matches (not an error)
				return filteredLogsMsg{output: "No matches found", err: nil}
			}
			return filteredLogsMsg{err: err}
		}

		return filteredLogsMsg{output: out.String(), err: nil}
	}
}

func (m Model) loadLambdaLogGroups() tea.Cmd {
	return func() tea.Msg {
		var filteredGroups []types.LogGroup
		var nextToken *string

		for {
			// Still use Lambda prefix to reduce API calls
			result, err := m.client.DescribeLogGroups(
				context.TODO(),
				&cloudwatchlogs.DescribeLogGroupsInput{
					LogGroupNamePattern: aws.String(
						fmt.Sprintf("/aws/lambda/dev-cot*-%s", m.env),
					),
					Limit:     aws.Int32(50),
					NextToken: nextToken,
				},
			)
			if err != nil {
				return logGroupsLoadedMsg{err: err}
			}

			log.Printf("result %d", len(result.LogGroups))

			// Check tags for each log group
			/*
				for _, group := range result.LogGroups {
					if group.LogGroupArn == nil {
						continue
					}

					tagsResult, err := m.client.ListTagsForResource(
						context.TODO(),
						&cloudwatchlogs.ListTagsForResourceInput{
							ResourceArn: group.LogGroupArn,
						},
					)
					if err != nil {

						log.Printf("%s", err)
						continue
					}

					log.Printf("group %s", *group.Arn)

					// Check if group has matching tag
					if tagValue, ok := tagsResult.Tags[internal.AWSTagKey]; ok {
						if tagValue == internal.AWSTagValue {
							filteredGroups = append(filteredGroups, group)
						}
					}
				}
			*/
			filteredGroups = append(filteredGroups, result.LogGroups...)

			if result.NextToken == nil {
				break
			}
			nextToken = result.NextToken
		}

		return logGroupsLoadedMsg{groups: filteredGroups}
	}
}

func (m Model) loadLogEvents(logGroupName string) tea.Cmd {
	return func() tea.Msg {
		// Get logs from last 10 minutes
		endTime := time.Now()
		startTime := endTime.Add(-10 * time.Minute)

		result, err := m.client.FilterLogEvents(
			context.TODO(),
			&cloudwatchlogs.FilterLogEventsInput{
				LogGroupName: aws.String(logGroupName),
				StartTime:    aws.Int64(startTime.UnixMilli()),
				EndTime:      aws.Int64(endTime.UnixMilli()),
				Limit:        aws.Int32(500),
			},
		)
		if err != nil {
			return logEventsLoadedMsg{err: err}
		}
		return logEventsLoadedMsg{events: result.Events}
	}
}
