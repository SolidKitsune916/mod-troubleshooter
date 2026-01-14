import { z } from 'zod';

/** API response envelope schema */
export const ApiResponseSchema = <T extends z.ZodTypeAny>(dataSchema: T) =>
  z.object({
    success: z.boolean(),
    data: dataSchema.optional(),
    error: z.string().optional(),
  });

/** Nexus user schema */
export const UserSchema = z.object({
  name: z.string(),
  avatar: z.string(),
  memberId: z.number(),
});

/** Nexus game schema */
export const GameSchema = z.object({
  id: z.number(),
  domainName: z.string(),
  name: z.string(),
});

/** Image schema */
export const ImageSchema = z.object({
  url: z.string(),
});

/** Mod category schema */
export const ModCategorySchema = z.object({
  name: z.string(),
});

/** Mod schema */
export const ModSchema = z.object({
  modId: z.number(),
  name: z.string(),
  summary: z.string(),
  version: z.string(),
  author: z.string(),
  pictureUrl: z.string(),
  modCategory: ModCategorySchema.nullable().optional(),
  game: GameSchema.nullable().optional(),
});

/** Mod file schema */
export const ModFileSchema = z.object({
  fileId: z.number(),
  name: z.string(),
  size: z.number(),
  version: z.string(),
  mod: ModSchema.nullable().optional(),
});

/** Mod file reference schema */
export const ModFileReferenceSchema = z.object({
  fileId: z.number(),
  optional: z.boolean(),
  file: ModFileSchema.nullable().optional(),
});

/** External resource schema */
export const ExternalResourceSchema = z.object({
  name: z.string(),
  resourceType: z.string(),
  resourceUrl: z.string(),
});

/** Revision summary schema */
export const RevisionSchema = z.object({
  revisionNumber: z.number(),
  createdAt: z.string(),
  revisionStatus: z.string(),
  totalSize: z.number(),
  collectionNotes: z.string().optional(),
});

/** Revision details schema (with mods) */
export const RevisionDetailsSchema = z.object({
  revisionNumber: z.number(),
  modFiles: z.array(ModFileReferenceSchema),
  externalResources: z.array(ExternalResourceSchema).optional(),
});

/** Collection schema */
export const CollectionSchema = z.object({
  id: z.string(),
  slug: z.string(),
  name: z.string(),
  summary: z.string(),
  description: z.string(),
  endorsements: z.number(),
  totalDownloads: z.number(),
  user: UserSchema,
  game: GameSchema,
  tileImage: ImageSchema.nullable().optional(),
  revisions: z.array(RevisionSchema).optional(),
  latestPublishedRevision: RevisionDetailsSchema.nullable().optional(),
});

/** Type exports inferred from schemas */
export type ApiResponse<T> = z.infer<ReturnType<typeof ApiResponseSchema<z.ZodType<T>>>>;
export type User = z.infer<typeof UserSchema>;
export type Game = z.infer<typeof GameSchema>;
export type Image = z.infer<typeof ImageSchema>;
export type ModCategory = z.infer<typeof ModCategorySchema>;
export type Mod = z.infer<typeof ModSchema>;
export type ModFile = z.infer<typeof ModFileSchema>;
export type ModFileReference = z.infer<typeof ModFileReferenceSchema>;
export type ExternalResource = z.infer<typeof ExternalResourceSchema>;
export type Revision = z.infer<typeof RevisionSchema>;
export type RevisionDetails = z.infer<typeof RevisionDetailsSchema>;
export type Collection = z.infer<typeof CollectionSchema>;
