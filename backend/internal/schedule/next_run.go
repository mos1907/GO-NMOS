package schedule

import (
	"strconv"
	"strings"
	"time"

	"go-nmos/backend/internal/models"
)

// NextRun returns the next run time for an automation job.
// For "interval" schedule_value is seconds (e.g. "1800").
// For "cron" schedule_value is a 5-field cron expression (e.g. "0 0 * * *" = daily at midnight).
// from is the time to compute from (e.g. last run or now).
func NextRun(job models.AutomationJob, from time.Time) *time.Time {
	if job.ScheduleType == "cron" && job.ScheduleValue != "" {
		next := nextCronRun(job.ScheduleValue, from)
		if next == nil {
			return nil
		}
		return next
	}
	// interval
	seconds, err := strconv.Atoi(strings.TrimSpace(job.ScheduleValue))
	if err != nil || seconds <= 0 {
		return nil
	}
	if job.LastRunAt == nil {
		// Never run: next run is now (run immediately)
		return &from
	}
	next := job.LastRunAt.Add(time.Duration(seconds) * time.Second)
	return &next
}

// nextCronRun parses a 5-field cron expression (min hour dom month dow) and returns the next run after from.
// Supports: number, * (every), step (e.g. */5), and ranges (e.g. 1-5). Month 1-12, dow 0-6 (0=Sunday).
func nextCronRun(cronExpr string, from time.Time) *time.Time {
	fields := strings.Fields(strings.TrimSpace(cronExpr))
	if len(fields) != 5 {
		return nil
	}
	// Start from the next minute to avoid running twice in the same minute
	cand := from.Truncate(time.Minute).Add(time.Minute)
	end := from.AddDate(2, 0, 0) // search up to 2 years
	for !cand.After(end) {
		if matchCronField(fields[0], cand.Minute(), 0, 59) &&
			matchCronField(fields[1], cand.Hour(), 0, 23) &&
			matchCronField(fields[2], cand.Day(), 1, 31) &&
			matchCronField(fields[3], int(cand.Month()), 1, 12) &&
			matchCronField(fields[4], int(cand.Weekday()), 0, 6) {
			return &cand
		}
		cand = cand.Add(time.Minute)
	}
	return nil
}

func matchCronField(field string, value, min, max int) bool {
	field = strings.TrimSpace(field)
	if field == "*" {
		return true
	}
	// Step: */5 or 0-30/5
	if strings.HasPrefix(field, "*/") {
		step, err := strconv.Atoi(strings.TrimPrefix(field, "*/"))
		if err != nil || step <= 0 {
			return false
		}
		return value%step == 0
	}
	if strings.Contains(field, "/") {
		parts := strings.SplitN(field, "/", 2)
		step, err := strconv.Atoi(strings.TrimSpace(parts[1]))
		if err != nil || step <= 0 {
			return false
		}
		// Range part
		if strings.Contains(parts[0], "-") {
			rangeParts := strings.SplitN(parts[0], "-", 2)
			lo, _ := strconv.Atoi(strings.TrimSpace(rangeParts[0]))
			hi, _ := strconv.Atoi(strings.TrimSpace(rangeParts[1]))
			if value < lo || value > hi {
				return false
			}
			return (value-lo)%step == 0
		}
		return value%step == 0
	}
	if strings.Contains(field, "-") {
		parts := strings.SplitN(field, "-", 2)
		lo, _ := strconv.Atoi(strings.TrimSpace(parts[0]))
		hi, _ := strconv.Atoi(strings.TrimSpace(parts[1]))
		return value >= lo && value <= hi
	}
	v, err := strconv.Atoi(field)
	if err != nil {
		return false
	}
	return value == v
}

// ShouldRun returns true if the job should run at time now (next run <= now or never run).
func ShouldRun(job models.AutomationJob, now time.Time) bool {
	next := NextRun(job, now)
	if next == nil {
		return false
	}
	return !now.Before(*next)
}
