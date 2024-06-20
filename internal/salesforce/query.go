package salesforce

import (
	"fmt"
)

const apexLogsQuery = `
SELECT
  Id,
  Application,
  Location,
  LogUserId,
  Operation,
  Request,
  RequestIdentifier,
  Status,
  StartTime,
  DurationMilliseconds,
  LogLength
FROM ApexLog
ORDER BY StartTime DESC
LIMIT 100
`

const debugLogsQuery = `
SELECT
  Id,
  ApexCode,
  ApexProfiling,
  Callout,
  Database,
  DeveloperName,
  Language,
  MasterLabel,
  System,
  Validation,
  Visualforce,
  Workflow
FROM DebugLevel
WHERE DeveloperName = '%s'
LIMIT 1
`

const traceFlagQuery = `
SELECT
  Id,
  DebugLevelId,
  ExpirationDate,
  TracedEntityId
FROM TraceFlag
WHERE TracedEntityId = '%s'
AND LogType = 'DEVELOPER_LOG'
LIMIT 1
`

type ApexLog struct {
	Attributes           Attributes
	ID                   string
	Application          string
	Location             string
	LogUserId            string
	Operation            string
	Request              string
	RequestIdentifier    string
	StartTime            string
	Status               string
	DurationMilliseconds int
	LogLength            int
}

type DebugLevel struct {
	Id            string
	ApexCode      string
	ApexProfiling string
	Callout       string
	Database      string
	DeveloperName string
	Language      string
	MasterLabel   string
	System        string
	Validation    string
	Visualforce   string
	Workflow      string
}

type TraceFlag struct {
	Id             string
	DebugLevelId   string
	ExpirationDate string
	TracedEntityId string
	LogType        string
}

func SelectApexLogs() string {
	return apexLogsQuery
}

func SelectDebugLogByDeveloperName(n string) string {
	return fmt.Sprintf(debugLogsQuery, n)
}

func SelectDebugLogTraceFlagByTracedId(i string) string {
	return fmt.Sprintf(traceFlagQuery, i)
}
