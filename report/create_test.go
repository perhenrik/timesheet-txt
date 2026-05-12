package report

import (
	"testing"
	"time"

	"github.com/JamesClonk/go-todotxt"
)

func TestCreateAggregatesAndFiltersTasks(t *testing.T) {
	end := time.Date(2026, 5, 10, 0, 0, 0, 0, time.UTC)
	durationHours := 48.0 // 2 days

	tasks := todotxt.TaskList{
		{ // included
			Completed:     true,
			CompletedDate: time.Date(2026, 5, 9, 0, 0, 0, 0, time.UTC),
			Projects:      []string{"alpha"},
			AdditionalTags: map[string]string{
				"task":  "api",
				"hours": "1.5",
			},
		},
		{ // included and aggregated with first
			Completed:     true,
			CompletedDate: time.Date(2026, 5, 9, 0, 0, 0, 0, time.UTC),
			Projects:      []string{"alpha"},
			AdditionalTags: map[string]string{
				"task":  "api",
				"hours": "2.0",
			},
		},
		{ // included as separate key
			Completed:     true,
			CompletedDate: time.Date(2026, 5, 10, 0, 0, 0, 0, time.UTC),
			Projects:      []string{"beta"},
			AdditionalTags: map[string]string{
				"task":  "ops",
				"hours": "1.0",
			},
		},
		{ // excluded: not completed
			Completed: false,
			Projects:  []string{"beta"},
			AdditionalTags: map[string]string{
				"task":  "ops",
				"hours": "5.0",
			},
		},
		{ // excluded: invalid hours
			Completed:     true,
			CompletedDate: time.Date(2026, 5, 9, 0, 0, 0, 0, time.UTC),
			Projects:      []string{"alpha"},
			AdditionalTags: map[string]string{
				"task":  "api",
				"hours": "invalid",
			},
		},
		{ // excluded: out of range
			Completed:     true,
			CompletedDate: time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
			Projects:      []string{"gamma"},
			AdditionalTags: map[string]string{
				"task":  "legacy",
				"hours": "8.0",
			},
		},
	}

	items := Create(tasks, end, durationHours)

	if len(items) != 2 {
		t.Fatalf("Create() items len = %d, want 2", len(items))
	}

	if items[0].Date.Format("2006-01-02") != "2026-05-09" || items[0].Task != "alpha.api" || items[0].Hours != 3.5 {
		t.Fatalf("Create() item[0] = %+v, want date=2026-05-09 task=alpha.api hours=3.5", items[0])
	}

	if items[1].Date.Format("2006-01-02") != "2026-05-10" || items[1].Task != "beta.ops" || items[1].Hours != 1.0 {
		t.Fatalf("Create() item[1] = %+v, want date=2026-05-10 task=beta.ops hours=1.0", items[1])
	}
}
