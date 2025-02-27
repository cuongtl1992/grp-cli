package models

import "time"

// ExecutionResult contains the outcome of a plan execution
type ExecutionResult struct {
	ID            string
	Success       bool
	TotalStages   int
	TotalJobs     int
	CompletedJobs int
	FailedJobs    int
	StartTime     time.Time
	EndTime       time.Time
	Duration      time.Duration
	Stages        []StageResult
}

// StageResult contains the outcome of a stage execution
type StageResult struct {
	Name      string
	Success   bool
	Jobs      []JobResult
	StartTime time.Time
	EndTime   time.Time
	Duration  time.Duration
}

// JobResult contains the outcome of a job execution
type JobResult struct {
	Name      string
	Type      string
	Success   bool
	Message   string
	StartTime time.Time
	EndTime   time.Time
	Duration  time.Duration
	Data      map[string]interface{}
}

// Artifact represents a file or data produced by a plugin
type Artifact struct {
	Name        string
	Type        string
	ContentType string
	Path        string
	Data        []byte
} 