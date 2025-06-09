package main

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/delving/hub3/ikuzo/domain/domainpb"
)

// examineIndexMessageSource inspects the message source for debugging
func examineIndexMessageSource(mesg *domainpb.IndexMessage) {
	// Directly examine the source bytes
	sourceBytes := mesg.GetSource()
	if sourceBytes == nil || len(sourceBytes) == 0 {
		slog.Error("IndexMessage Source is nil or empty")
		return
	}

	// Log the first bytes to help with debugging
	previewLen := 50
	if len(sourceBytes) < previewLen {
		previewLen = len(sourceBytes)
	}
	
	slog.Debug("IndexMessage Source preview", 
		"length", len(sourceBytes),
		"starts_with", string(sourceBytes[:previewLen]))
	
	// Directly try to unmarshal the source bytes
	var sourceObj interface{}
	if err := json.Unmarshal(sourceBytes, &sourceObj); err != nil {
		slog.Error("Failed to unmarshal Source bytes as JSON", "error", err)
		return
	}
	
	// Check what type the source decoded to
	switch objTyped := sourceObj.(type) {
	case map[string]interface{}:
		// Extract source keys
		keys := getMapKeys(objTyped)
		slog.Debug("Source parsed as object", "keys", keys)
		
		// Look for resources field
		if resources, ok := objTyped["resources"].([]interface{}); ok {
			slog.Debug("Source contains resources array", "count", len(resources))
			
			if len(resources) > 0 {
				// Check first resource
				if firstResource, ok := resources[0].(map[string]interface{}); ok {
					resourceKeys := getMapKeys(firstResource)
					slog.Debug("First resource", "keys", resourceKeys)
					
					// Check for context fields
					examineContextFieldDetailed(firstResource, "context")
					examineContextFieldDetailed(firstResource, "graphExternalContext")
				}
			}
		} else {
			slog.Debug("Resources field not found or not an array")
		}
		
	case string:
		// Source is a string, not expected
		previewLen := 30
		if len(objTyped) < previewLen {
			previewLen = len(objTyped)
		}
		slog.Error("Source decoded as string, expected object", 
			"length", len(objTyped),
			"preview", objTyped[:previewLen])
	
	default:
		// Unexpected type
		slog.Error("Source decoded as unexpected type", 
			"type", fmt.Sprintf("%T", sourceObj))
	}
}

// examineContextFieldDetailed is a more detailed version of examineContextField
func examineContextFieldDetailed(resource map[string]interface{}, fieldName string) {
	field := resource[fieldName]
	if field == nil {
		slog.Debug(fmt.Sprintf("%s field is nil", fieldName))
		return
	}
	
	switch fieldTyped := field.(type) {
	case []interface{}:
		slog.Debug(fmt.Sprintf("%s is array", fieldName), "count", len(fieldTyped))
		
		if len(fieldTyped) > 0 {
			// Examine first item
			firstItem := fieldTyped[0]
			slog.Debug(fmt.Sprintf("First %s item", fieldName), 
				"type", fmt.Sprintf("%T", firstItem))
				
			// If it's a map, look at its keys
			if itemMap, ok := firstItem.(map[string]interface{}); ok {
				keys := getMapKeys(itemMap)
				slog.Debug(fmt.Sprintf("First %s item keys", fieldName), "keys", keys)
				
				// Check for level field
				if level, ok := itemMap["Level"]; ok {
					slog.Debug(fmt.Sprintf("%s has Level field", fieldName), "value", level)
				} else {
					slog.Debug(fmt.Sprintf("%s missing Level field", fieldName))
				}
			}
		}
		
	default:
		slog.Debug(fmt.Sprintf("%s unexpected type", fieldName), 
			"type", fmt.Sprintf("%T", field))
	}
}