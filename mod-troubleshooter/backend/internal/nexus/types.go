package nexus

import "time"

// API endpoints
const (
	GraphQLEndpoint = "https://api.nexusmods.com/v2/graphql"
	RESTAPIBase     = "https://api.nexusmods.com/v1"
)

// Collection represents a Nexus Mods collection.
type Collection struct {
	ID             string           `json:"id"`
	Slug           string           `json:"slug"`
	Name           string           `json:"name"`
	Summary        string           `json:"summary"`
	Description    string           `json:"description"`
	Endorsements   int              `json:"endorsements"`
	TotalDownloads int              `json:"totalDownloads"`
	User           User             `json:"user"`
	Game           Game             `json:"game"`
	TileImage      *Image           `json:"tileImage"`
	Revisions      []Revision       `json:"revisions,omitempty"`
	LatestRevision *RevisionDetails `json:"latestPublishedRevision,omitempty"`
}

// User represents a Nexus Mods user.
type User struct {
	Name     string `json:"name"`
	Avatar   string `json:"avatar"`
	MemberID int    `json:"memberId"`
}

// Game represents a game on Nexus Mods.
type Game struct {
	ID         int    `json:"id"`
	DomainName string `json:"domainName"`
	Name       string `json:"name"`
}

// Image represents an image with URL.
type Image struct {
	URL string `json:"url"`
}

// Revision represents a collection revision summary.
type Revision struct {
	RevisionNumber  int       `json:"revisionNumber"`
	CreatedAt       time.Time `json:"createdAt"`
	RevisionStatus  string    `json:"revisionStatus"`
	TotalSize       int64     `json:"totalSize"`
	CollectionNotes string    `json:"collectionNotes,omitempty"`
}

// RevisionDetails contains full revision information including mods.
type RevisionDetails struct {
	RevisionNumber    int                `json:"revisionNumber"`
	ModFiles          []ModFileReference `json:"modFiles"`
	ExternalResources []ExternalResource `json:"externalResources,omitempty"`
}

// ModFileReference is a reference to a mod file within a collection.
type ModFileReference struct {
	FileID   int      `json:"fileId"`
	Optional bool     `json:"optional"`
	File     *ModFile `json:"file"`
}

// ModFile represents a downloadable mod file.
type ModFile struct {
	FileID  int    `json:"fileId"`
	Name    string `json:"name"`
	Size    int64  `json:"size"`
	Version string `json:"version"`
	Mod     *Mod   `json:"mod"`
}

// Mod represents a mod on Nexus Mods.
type Mod struct {
	ModID       int          `json:"modId"`
	Name        string       `json:"name"`
	Summary     string       `json:"summary"`
	Version     string       `json:"version"`
	Author      string       `json:"author"`
	PictureURL  string       `json:"pictureUrl"`
	ModCategory *ModCategory `json:"modCategory"`
	Game        *Game        `json:"game"`
}

// ModCategory represents a mod category.
type ModCategory struct {
	Name string `json:"name"`
}

// ExternalResource represents an external resource in a collection.
type ExternalResource struct {
	Name         string `json:"name"`
	ResourceType string `json:"resourceType"`
	ResourceURL  string `json:"resourceUrl"`
}

// GraphQL request/response types

// GraphQLRequest is the request body for GraphQL queries.
type GraphQLRequest struct {
	Query         string                 `json:"query"`
	Variables     map[string]interface{} `json:"variables,omitempty"`
	OperationName string                 `json:"operationName,omitempty"`
}

// GraphQLResponse is the generic GraphQL response wrapper.
type GraphQLResponse struct {
	Data   map[string]interface{} `json:"data,omitempty"`
	Errors []GraphQLError         `json:"errors,omitempty"`
}

// GraphQLError represents a GraphQL error.
type GraphQLError struct {
	Message    string                 `json:"message"`
	Locations  []GraphQLErrorLocation `json:"locations,omitempty"`
	Path       []interface{}          `json:"path,omitempty"`
	Extensions map[string]interface{} `json:"extensions,omitempty"`
}

// GraphQLErrorLocation indicates where in the query an error occurred.
type GraphQLErrorLocation struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

// CollectionResponse wraps the collection query response.
type CollectionResponse struct {
	Collection *Collection `json:"collection"`
}

// CollectionRevisionsResponse wraps the collection revisions query response.
type CollectionRevisionsResponse struct {
	Collection *Collection `json:"collection"`
}

// CollectionRevisionModsResponse wraps the revision mods query response.
type CollectionRevisionModsResponse struct {
	CollectionRevision *RevisionDetails `json:"collectionRevision"`
}

// RateLimitInfo contains rate limiting information from API responses.
type RateLimitInfo struct {
	HourlyLimit     int
	HourlyRemaining int
	DailyLimit      int
	DailyRemaining  int
}

// DownloadLink represents a download URL returned by the Nexus API.
type DownloadLink struct {
	Name      string `json:"name"`
	ShortName string `json:"short_name"`
	URI       string `json:"URI"`
}

// DownloadLinksResponse wraps the download links array from the REST API.
type DownloadLinksResponse []DownloadLink
