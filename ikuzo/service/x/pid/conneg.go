package pid

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
)

type contextKey string

const SelectedProfileKey contextKey = "selected_profile"

// ProfileInfo represents a profile with its URI and optional token
type ProfileInfo struct {
	URI   string
	Token string // Optional token identifier
}

// ProfileRequest represents a requested profile with quality value
type ProfileRequest struct {
	URI     string
	Token   string
	Quality float64
}

// SetProfileHeaders handles profile content negotiation per W3C spec and sets response headers
func SetProfileHeaders(r *http.Request, w http.ResponseWriter, response *StoreResponse) *http.Request {
	// Get available profiles from CurrentPID
	availableProfiles := []*ProfileInfo{}
	if response.CurrentPID != nil {
		availableProfiles = response.GetSupportedProfiles()
	}

	// Parse profile requests from QSA first (takes precedence), then Accept-Profile header
	requestedProfiles := parseProfileRequests(r, availableProfiles)

	// Select the best matching profile
	selectedProfile := selectProfile(requestedProfiles, availableProfiles)

	// Set response headers per W3C spec (Link headers, not Content-Profile)
	setResponseHeaders(w, availableProfiles, selectedProfile)

	// Add selected profile to request context
	var profileURI string
	if selectedProfile != nil {
		profileURI = selectedProfile.URI
	}
	slog.Info("selected profile", "profile", selectedProfile, "uri", profileURI)
	ctx := context.WithValue(r.Context(), SelectedProfileKey, profileURI)
	return r.WithContext(ctx)
}

// parseProfileRequests parses profile requests from QSA and Accept-Profile header
func parseProfileRequests(r *http.Request, availableProfiles []*ProfileInfo) []*ProfileRequest {
	var requests []*ProfileRequest

	// 1. Check Query String Arguments first (higher precedence per spec)
	if profileParam := r.URL.Query().Get("_profile"); profileParam != "" {
		// Handle multiple profiles in QSA (comma-separated)
		profiles := strings.SplitSeq(profileParam, ",")
		for profile := range profiles {
			profile = strings.TrimSpace(profile)
			if profile != "" {
				// Try to resolve token to URI if it's a token
				uri := resolveTokenToURI(profile, availableProfiles)
				if uri == "" {
					uri = profile // Assume it's already a URI
				}
				requests = append(requests, &ProfileRequest{
					URI:     uri,
					Token:   profile,
					Quality: 1.0, // QSA has highest quality
				})
			}
		}
	}

	// 2. Parse Accept-Profile header (lower precedence)
	acceptProfile := r.Header.Get("Accept-Profile")
	if acceptProfile != "" {
		headerRequests := parseAcceptProfileHeader(acceptProfile, availableProfiles)
		requests = append(requests, headerRequests...)
	}

	return requests
}

// resolveTokenToURI resolves a token to its corresponding URI
func resolveTokenToURI(token string, availableProfiles []*ProfileInfo) string {
	for _, profile := range availableProfiles {
		if profile.Token == token {
			return profile.URI
		}
	}
	return ""
}

// parseAcceptProfileHeader parses the Accept-Profile header per W3C spec
func parseAcceptProfileHeader(acceptHeader string, availableProfiles []*ProfileInfo) []*ProfileRequest {
	if acceptHeader == "" {
		return []*ProfileRequest{}
	}

	var requests []*ProfileRequest

	// Split by comma for multiple profiles
	parts := strings.SplitSeq(acceptHeader, ",")
	for part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Parse profile and quality value
		var profileStr string
		quality := 1.0

		// Check for quality value (;q=0.8)
		if idx := strings.Index(part, ";"); idx != -1 {
			profileStr = strings.TrimSpace(part[:idx])
			qPart := strings.TrimSpace(part[idx+1:])
			if strings.HasPrefix(qPart, "q=") {
				if q, err := strconv.ParseFloat(qPart[2:], 64); err == nil && q >= 0 && q <= 1 {
					quality = q
				}
			}
		} else {
			profileStr = part
		}

		// Remove angled brackets per W3C spec: <uri>
		profileStr = strings.Trim(profileStr, "<>")

		if profileStr != "" {
			// Try to resolve as token first, then treat as URI
			uri := resolveTokenToURI(profileStr, availableProfiles)
			if uri == "" {
				uri = profileStr // Assume it's a URI
			}

			requests = append(requests, &ProfileRequest{
				URI:     uri,
				Token:   profileStr,
				Quality: quality,
			})
		}
	}

	return requests
}

// selectProfile selects the best matching profile per W3C spec algorithm
func selectProfile(requested []*ProfileRequest, available []*ProfileInfo) *ProfileInfo {
	if len(available) == 0 {
		return nil
	}

	// If no profiles requested, return first available (default)
	if len(requested) == 0 {
		return available[0]
	}

	// Find exact matches first, ordered by quality
	var bestMatch *ProfileInfo
	var bestQuality float64 = -1

	for _, req := range requested {
		for _, avail := range available {
			// Check for exact URI match
			if req.URI == avail.URI && req.Quality > bestQuality {
				bestMatch = avail
				bestQuality = req.Quality
			}
		}
	}

	if bestMatch != nil {
		return bestMatch
	}

	// No exact match found, return first available as fallback
	return available[0]
}

// setResponseHeaders sets the appropriate response headers per W3C spec
func setResponseHeaders(w http.ResponseWriter, available []*ProfileInfo, selected *ProfileInfo) {
	// Set Link header with rel="profile" for selected profile (W3C spec requirement)
	if selected != nil {
		linkHeader := fmt.Sprintf("<%s>; rel=\"profile\"", selected.URI)
		w.Header().Set("Link", linkHeader)
	}

	// Set Vary header to indicate profile-based content negotiation
	// (Not explicitly required by spec, but good practice for caching)
	w.Header().Set("Vary", "Accept-Profile")

	// Optionally set additional Link headers for profile discovery
	// This supports the "list profiles" functionality mentioned in the spec
	if len(available) > 1 {
		var allLinks []string
		for _, profile := range available {
			if profile != selected {
				link := fmt.Sprintf("<%s>; rel=\"alternate\"; type=\"application/json\"", profile.URI)
				if profile.Token != "" {
					link += fmt.Sprintf("; token=\"%s\"", profile.Token)
				}
				allLinks = append(allLinks, link)
			}
		}

		if len(allLinks) > 0 {
			existing := w.Header().Get("Link")
			if existing != "" {
				w.Header().Set("Link", existing+", "+strings.Join(allLinks, ", "))
			} else {
				w.Header().Set("Link", strings.Join(allLinks, ", "))
			}
		}
	}
}

// GetSelectedProfile retrieves the selected profile URI from request context
func GetSelectedProfile(ctx context.Context) string {
	if profile, ok := ctx.Value(SelectedProfileKey).(string); ok {
		return profile
	}
	return ""
}

// ListAvailableProfiles returns all available profiles for profile discovery
func ListAvailableProfiles(response *StoreResponse) []*ProfileInfo {
	if response.CurrentPID != nil {
		return response.GetSupportedProfiles()
	}
	return []*ProfileInfo{}
}
