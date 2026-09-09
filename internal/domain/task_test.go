package domain

import "testing"

func TestTaskValidate(t *testing.T) {
	tests := []struct {
		name    string
		task    Task
		wantErr error
	}{
		{
			name: "valid pending task",
			task: Task{
				Title:  "Buy milk",
				Status: TaskStatusPending,
			},
		},
		{
			name: "valid in progress task",
			task: Task{
				Title:  "Build API",
				Status: TaskStatusInProgress,
			},
		},
		{
			name: "valid completed task",
			task: Task{
				Title:  "Write tests",
				Status: TaskStatusCompleted,
			},
		},
		{
			name: "empty title",
			task: Task{
				Title:  "",
				Status: TaskStatusPending,
			},
			wantErr: ErrInvalidTaskTitle,
		},
		{
			name: "whitespace title",
			task: Task{
				Title:  "   ",
				Status: TaskStatusPending,
			},
			wantErr: ErrInvalidTaskTitle,
		},
		{
			name: "invalid status",
			task: Task{
				Title:  "Test",
				Status: "unknown",
			},
			wantErr: ErrInvalidTaskStatus,
		},
		{
			name: "description too long",
			task: Task{
				Title:       "Test",
				Status:      TaskStatusPending,
				Description: stringPtr(longString(5001)),
			},
			wantErr: ErrInvalidDescription,
		},
		{
			name: "assignee too long",
			task: Task{
				Title:    "Test",
				Status:   TaskStatusPending,
				Assignee: stringPtr(longString(101)),
			},
			wantErr: ErrInvalidAssignee,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.task.Validate()

			if err != tt.wantErr {
				t.Fatalf(
					"expected error %v, got %v",
					tt.wantErr,
					err,
				)
			}
		})
	}
}

func TestIsValidTaskStatus(t *testing.T) {
	tests := []struct {
		status TaskStatus
		valid  bool
	}{
		{
			status: TaskStatusPending,
			valid:  true,
		},
		{
			status: TaskStatusInProgress,
			valid:  true,
		},
		{
			status: TaskStatusCompleted,
			valid:  true,
		},
		{
			status: "invalid",
			valid:  false,
		},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			if got := IsValidTaskStatus(tt.status); got != tt.valid {
				t.Fatalf(
					"expected %v, got %v",
					tt.valid,
					got,
				)
			}
		})
	}
}

func stringPtr(value string) *string {
	return &value
}

func longString(length int) string {
	result := make([]byte, length)

	for i := range result {
		result[i] = 'a'
	}

	return string(result)
}
