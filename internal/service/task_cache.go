package service

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

func taskListCacheKey(params ListTasksParams) string {
	status := ""

	if params.Status != nil {
		status = string(*params.Status)
	}

	assignee := ""

	if params.Assignee != nil {
		assignee = *params.Assignee
	}

	raw := fmt.Sprintf(
		"%s|%s|%d|%d",
		status,
		assignee,
		params.Page,
		params.PageSize,
	)

	hash := sha256.Sum256([]byte(raw))

	return "task_manager:tasks:list:" +
		hex.EncodeToString(hash[:])
}
