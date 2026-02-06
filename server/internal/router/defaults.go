package router

import (
	"context"

	"mindfs/server/internal/fs"
)

// DefaultViewPayload builds a minimal view payload for the main canvas.
func DefaultViewPayload(entries []fs.Entry) map[string]any {
	return map[string]any{
		"schema": "default.list",
		"items":  entries,
	}
}

// DefaultViewTree builds a json-render tree for the main canvas.
func DefaultViewTree(entries []fs.Entry) map[string]any {
	return map[string]any{
		"root": "root",
		"elements": map[string]any{
			"root": map[string]any{
				"key":  "root",
				"type": "DefaultListView",
				"props": map[string]any{
					"entries": entries,
				},
			},
		},
	}
}

// DefaultActionResponse returns a baseline action response with a full view payload.
func DefaultActionResponse(entries []fs.Entry) ActionResponse {
	return ActionResponse{
		Status:  "ok",
		Handled: false,
		View: map[string]any{
			"type":    "full",
			"payload": DefaultViewPayload(entries),
		},
	}
}

// DefaultHandler creates a fallback handler that returns the default view.
func DefaultHandler(listing func() ([]fs.Entry, error)) ActionHandler {
	return func(_ context.Context, _ ActionRequest) (ActionResponse, error) {
		entries, err := listing()
		if err != nil {
			return ActionResponse{
				Status:  "error",
				Handled: false,
				Error: map[string]any{
					"code":    "listing_failed",
					"message": err.Error(),
				},
			}, nil
		}
		return DefaultActionResponse(entries), nil
	}
}
