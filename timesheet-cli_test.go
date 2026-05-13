package main

import (
	"math"
	"strconv"
	"testing"
	"time"

	"github.com/JamesClonk/go-todotxt"
)

func TestMarkTaskAsRunningAndStopped(t *testing.T) {
	task, err := todotxt.ParseTask("+proj task:api")
	if err != nil {
		t.Fatalf("ParseTask() error = %v", err)
	}

	startedAt := time.Date(2026, 5, 13, 8, 0, 0, 0, time.UTC)
	stoppedAt := startedAt.Add(90 * time.Minute)

	markTaskAsRunning(task, startedAt)
	if !isRunningTask(*task) {
		t.Fatalf("task should be running after markTaskAsRunning")
	}

	hours, err := markTaskAsStopped(task, stoppedAt)
	if err != nil {
		t.Fatalf("markTaskAsStopped() error = %v", err)
	}

	if math.Abs(hours-1.5) > 0.0001 {
		t.Fatalf("markTaskAsStopped() hours = %f, want 1.5", hours)
	}

	if isRunningTask(*task) {
		t.Fatalf("task should not be running after stop")
	}

	if !task.Completed {
		t.Fatalf("task should be completed after stop")
	}

	if !task.CompletedDate.Equal(stoppedAt) {
		t.Fatalf("task completed date = %s, want %s", task.CompletedDate, stoppedAt)
	}

	if task.AdditionalTags["hours"] != "1.5000" {
		t.Fatalf("task hours tag = %q, want %q", task.AdditionalTags["hours"], "1.5000")
	}
}

func TestStopRunningTasksStopsAllRunningTasks(t *testing.T) {
	taskA, err := todotxt.ParseTask("+alpha task:api")
	if err != nil {
		t.Fatalf("ParseTask() error = %v", err)
	}
	taskB, err := todotxt.ParseTask("+beta task:ops")
	if err != nil {
		t.Fatalf("ParseTask() error = %v", err)
	}

	markTaskAsRunning(taskA, time.Date(2026, 5, 13, 9, 0, 0, 0, time.UTC))
	markTaskAsRunning(taskB, time.Date(2026, 5, 13, 10, 0, 0, 0, time.UTC))

	tasklist := todotxt.TaskList{*taskA, *taskB}
	stoppedAt := time.Date(2026, 5, 13, 11, 0, 0, 0, time.UTC)

	stoppedCount, totalHours, err := stopRunningTasks(&tasklist, stoppedAt)
	if err != nil {
		t.Fatalf("stopRunningTasks() error = %v", err)
	}

	if stoppedCount != 2 {
		t.Fatalf("stopRunningTasks() stoppedCount = %d, want 2", stoppedCount)
	}

	if math.Abs(totalHours-3.0) > 0.0001 {
		t.Fatalf("stopRunningTasks() totalHours = %f, want 3.0", totalHours)
	}

	for i, task := range tasklist {
		if !task.Completed {
			t.Fatalf("tasklist[%d] should be completed", i)
		}

		hours, parseErr := strconv.ParseFloat(task.AdditionalTags["hours"], 64)
		if parseErr != nil {
			t.Fatalf("tasklist[%d] invalid hours tag: %v", i, parseErr)
		}
		if hours <= 0 {
			t.Fatalf("tasklist[%d] hours should be > 0", i)
		}
	}
}
