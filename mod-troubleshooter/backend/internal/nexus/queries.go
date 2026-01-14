package nexus

// GraphQL queries for Nexus Mods API

// CollectionQuery fetches collection metadata including mod list.
const CollectionQuery = `
query Collection($slug: String!) {
  collection(slug: $slug) {
    id
    slug
    name
    summary
    description
    endorsements
    totalDownloads
    user {
      name
      avatar
      memberId
    }
    game {
      id
      domainName
      name
    }
    tileImage {
      url
    }
    latestPublishedRevision {
      revisionNumber
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
            summary
            version
            author
            pictureUrl
            modCategory {
              name
            }
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
`

// CollectionRevisionsQuery fetches revision history for a collection.
const CollectionRevisionsQuery = `
query CollectionRevisions($domainName: String, $slug: String!) {
  collection(domainName: $domainName, slug: $slug) {
    revisions {
      revisionNumber
      createdAt
      revisionStatus
      totalSize
    }
  }
}
`

// CollectionRevisionModsQuery fetches mod files for a specific revision.
const CollectionRevisionModsQuery = `
query CollectionRevisionMods($revision: Int, $slug: String!) {
  collectionRevision(revision: $revision, slug: $slug) {
    revisionNumber
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
          summary
          pictureUrl
          game {
            domainName
          }
        }
      }
    }
  }
}
`
