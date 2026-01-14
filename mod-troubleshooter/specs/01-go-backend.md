# Go Backend & Nexus API Integration

## Overview

Go backend service that handles Nexus Mods API communication, archive downloading, and data processing. Serves REST API to the React frontend.

## Nexus API Authentication

### API Key Setup
- User provides their Nexus API key via settings UI
- Key stored securely (environment variable or encrypted config)
- All Nexus requests include `apikey` header

### Rate Limiting
- Nexus allows ~100 requests/day for free, ~2500/day for Premium
- Implement request queuing and backoff
- Cache responses aggressively

## API Endpoints

### Collections

```
GET /api/collections/:gameId/:slug
```
Fetch collection metadata including mod list.

Response:
```json
{
  "id": "abc123",
  "name": "Ultimate Skyrim",
  "slug": "qdurkx",
  "summary": "Complete modding guide",
  "author": { "name": "AuthorName" },
  "modCount": 450,
  "revisions": [
    { "number": 15, "createdAt": "2024-01-15" }
  ],
  "mods": [
    {
      "modId": 12345,
      "fileId": 67890,
      "name": "SkyUI",
      "optional": false,
      "category": "User Interface"
    }
  ]
}
```

### FOMOD Analysis

```
POST /api/fomod/analyze
```
Download mod archive and extract FOMOD structure.

Request:
```json
{
  "modId": 12345,
  "fileId": 67890,
  "gameId": "skyrimspecialedition"
}
```

Response:
```json
{
  "modName": "Example Mod",
  "hasFomod": true,
  "info": {
    "name": "Example Mod",
    "author": "Author",
    "version": "1.0.0",
    "description": "A cool mod"
  },
  "steps": [
    {
      "name": "Choose Version",
      "groups": [
        {
          "name": "Select One",
          "type": "SelectExactlyOne",
          "plugins": [
            {
              "name": "Standard",
              "description": "Normal installation",
              "image": "fomod/standard.png",
              "files": [
                { "source": "standard/", "destination": "" }
              ],
              "typeDescriptor": "Recommended"
            }
          ]
        }
      ]
    }
  ],
  "conditionalInstalls": [],
  "requiredFiles": []
}
```

### Load Order

```
GET /api/collections/:gameId/:slug/loadorder
```
Get recommended plugin load order from collection.

Response:
```json
{
  "plugins": [
    {
      "name": "Skyrim.esm",
      "type": "master",
      "modId": null,
      "index": 0
    },
    {
      "name": "SkyUI_SE.esp",
      "type": "esp",
      "modId": 12604,
      "index": 1,
      "masters": ["Skyrim.esm"]
    }
  ],
  "warnings": [
    {
      "type": "missing_master",
      "plugin": "example.esp",
      "missingMaster": "RequiredMod.esm"
    }
  ]
}
```

### Conflict Analysis

```
POST /api/conflicts/analyze
```
Analyze file conflicts between mods.

Request:
```json
{
  "mods": [
    { "modId": 12345, "fileId": 67890 },
    { "modId": 12346, "fileId": 67891 }
  ]
}
```

Response:
```json
{
  "conflicts": [
    {
      "file": "textures/armor/steel.dds",
      "mods": [
        { "modId": 12345, "name": "Steel Armor Retexture" },
        { "modId": 12346, "name": "Complete Armor Overhaul" }
      ],
      "winner": { "modId": 12346 },
      "severity": "low"
    }
  ],
  "summary": {
    "totalConflicts": 15,
    "highSeverity": 2,
    "mediumSeverity": 5,
    "lowSeverity": 8
  }
}
```

### Settings

```
GET /api/settings
PUT /api/settings
```
Get/update user settings including API key.

## GraphQL Queries

### Collection Info
```graphql
query Collection($slug: String!) {
  collection(slug: $slug) {
    id
    slug
    name
    summary
    description
    endorsements
    totalDownloads
    user { name avatar memberId }
    game { domainName }
    tileImage { url }
    latestPublishedRevision {
      revisionNumber
      modFiles {
        fileId
        optional
        file {
          mod {
            modId
            name
            summary
            version
            author
            pictureUrl
            modCategory { name }
          }
        }
      }
      externalResources {
        name
        resourceType
        resourceUrl
      }
    }
  }
}
```

### Collection Revisions
```graphql
query CollectionRevisions($domainName: String, $slug: String!) {
  collection(domainName: $domainName, slug: $slug) {
    revisions {
      revisionNumber
      createdAt
      revisionStatus
      totalSize
      collectionChangelog { description }
    }
  }
}
```

### Mod Files (for download)
```graphql
query CollectionRevisionMods($revision: Int, $slug: String!) {
  collectionRevision(revision: $revision, slug: $slug) {
    modFiles {
      fileId
      optional
      file {
        fileId
        name
        size
        version
        mod {
          modId
          name
          author
          game { domainName }
        }
      }
    }
  }
}
```

## Download Flow

1. **Get download links** (requires Premium)
   ```
   GET https://api.nexusmods.com/v1/games/{game}/mods/{modId}/files/{fileId}/download_link.json
   ```

2. **Download archive**
   - Stream to temp file
   - Verify file integrity

3. **Extract archive**
   - Support .zip, .7z, .rar
   - Extract only `fomod/` directory for FOMOD analysis
   - Extract file manifest for conflict analysis

4. **Parse FOMOD**
   - Parse `fomod/info.xml` for metadata
   - Parse `fomod/ModuleConfig.xml` for install steps

5. **Cache results**
   - Store parsed FOMOD in SQLite
   - Cache for 7 days (configurable)

## Project Structure

```
backend/
├── cmd/
│   └── server/
│       └── main.go           # Entry point
├── internal/
│   ├── config/
│   │   └── config.go         # Configuration loading
│   ├── nexus/
│   │   ├── client.go         # Nexus API client
│   │   ├── graphql.go        # GraphQL queries
│   │   └── types.go          # API response types
│   ├── fomod/
│   │   ├── parser.go         # FOMOD XML parser
│   │   └── types.go          # FOMOD data structures
│   ├── archive/
│   │   └── extractor.go      # Archive extraction
│   ├── conflict/
│   │   └── analyzer.go       # Conflict detection
│   ├── cache/
│   │   └── sqlite.go         # SQLite caching
│   └── handlers/
│       ├── collections.go
│       ├── fomod.go
│       ├── loadorder.go
│       ├── conflicts.go
│       └── settings.go
├── pkg/
│   └── response/
│       └── response.go       # Standard response helpers
├── go.mod
├── go.sum
└── Makefile
```

## Dependencies

```go
// go.mod
require (
    github.com/rs/cors v1.10.1
    github.com/mattn/go-sqlite3 v1.14.22
    github.com/mholt/archiver/v4 v4.0.0-alpha.8  // Multi-format archive support
)
```

## Environment Variables

```bash
NEXUS_API_KEY=your-api-key-here
PORT=8080
DATA_DIR=./data
CACHE_TTL_HOURS=168  # 7 days
```

## Acceptance Criteria

- [ ] Server starts and serves health endpoint
- [ ] Can authenticate with Nexus API using provided key
- [ ] Can fetch collection metadata via GraphQL
- [ ] Can download mod archives (Premium required)
- [ ] Can extract FOMOD XML from archives
- [ ] Can parse FOMOD ModuleConfig.xml
- [ ] Responses follow standard envelope format
- [ ] CORS configured for frontend dev server
- [ ] Rate limiting prevents API abuse
- [ ] Results cached in SQLite
