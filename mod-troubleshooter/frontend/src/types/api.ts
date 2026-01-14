import { z } from 'zod';

/** API response envelope schema - matches backend Response struct */
export const ApiResponseSchema = <T extends z.ZodTypeAny>(dataSchema: T) =>
  z.object({
    data: dataSchema.optional(),
    error: z.string().optional(),
    message: z.string().optional(),
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

// ============================================
// FOMOD Types
// ============================================

/** FOMOD group type - selection behavior */
export const GroupTypeSchema = z.enum([
  'SelectAtLeastOne',
  'SelectAtMostOne',
  'SelectExactlyOne',
  'SelectAll',
  'SelectAny',
]);

/** FOMOD plugin type - installation recommendation */
export const PluginTypeSchema = z.enum([
  'Required',
  'Optional',
  'Recommended',
  'NotUsable',
  'CouldBeUsable',
]);

/** FOMOD file state for dependencies */
export const FileStateSchema = z.enum(['Missing', 'Inactive', 'Active']);

/** FOMOD dependency operator */
export const DependencyOperatorSchema = z.enum(['And', 'Or']);

/** FOMOD info from info.xml */
export const FomodInfoSchema = z.object({
  name: z.string().optional(),
  author: z.string().optional(),
  version: z.string().optional(),
  description: z.string().optional(),
  website: z.string().optional(),
  id: z.string().optional(),
});

/** Header image configuration */
export const HeaderImageSchema = z.object({
  path: z.string(),
  showFade: z.boolean(),
  height: z.number().optional(),
});

/** File to install */
export const FileInstallSchema = z.object({
  source: z.string(),
  destination: z.string().optional(),
  priority: z.number().optional(),
  alwaysInstall: z.boolean().optional(),
  installIfUsable: z.boolean().optional(),
});

/** Folder to install */
export const FolderInstallSchema = z.object({
  source: z.string(),
  destination: z.string().optional(),
  priority: z.number().optional(),
  alwaysInstall: z.boolean().optional(),
  installIfUsable: z.boolean().optional(),
});

/** File list containing files and folders */
export const FileListSchema = z.object({
  files: z.array(FileInstallSchema).optional(),
  folders: z.array(FolderInstallSchema).optional(),
});

/** File dependency condition */
export const FileDependencySchema = z.object({
  file: z.string(),
  state: FileStateSchema,
});

/** Flag dependency condition */
export const FlagDependencySchema = z.object({
  flag: z.string(),
  value: z.string(),
});

/** Version dependency condition */
export const VersionDependencySchema = z.object({
  version: z.string(),
});

/** Dependency - recursive structure for conditions */
export const DependencySchema: z.ZodType<Dependency> = z.lazy(() =>
  z.object({
    operator: DependencyOperatorSchema.optional(),
    children: z.array(DependencySchema).optional(),
    fileDependency: FileDependencySchema.optional(),
    flagDependency: FlagDependencySchema.optional(),
    gameDependency: VersionDependencySchema.optional(),
    fommDependency: VersionDependencySchema.optional(),
  }),
);

/** Condition flag set when plugin is selected */
export const ConditionFlagSchema = z.object({
  name: z.string(),
  value: z.string(),
});

/** Dependency pattern mapping condition to plugin type */
export const DependencyPatternSchema = z.object({
  dependencies: DependencySchema.optional(),
  type: PluginTypeSchema,
});

/** Dependency-based plugin type */
export const DependencyPluginTypeSchema = z.object({
  defaultType: PluginTypeSchema,
  patterns: z.array(DependencyPatternSchema).optional(),
});

/** Type descriptor for plugin */
export const TypeDescriptorSchema = z.object({
  type: PluginTypeSchema.optional(),
  dependencyType: DependencyPluginTypeSchema.optional(),
});

/** Plugin - selectable option in installer */
export const PluginSchema = z.object({
  name: z.string(),
  description: z.string().optional(),
  image: z.string().optional(),
  files: FileListSchema.optional(),
  conditionFlags: z.array(ConditionFlagSchema).optional(),
  typeDescriptor: TypeDescriptorSchema.optional(),
});

/** Option group with selection constraints */
export const OptionGroupSchema = z.object({
  name: z.string(),
  type: GroupTypeSchema,
  plugins: z.array(PluginSchema).optional(),
});

/** Installation step */
export const InstallStepSchema = z.object({
  name: z.string(),
  visible: DependencySchema.optional(),
  optionGroups: z.array(OptionGroupSchema).optional(),
});

/** Conditional file installation item */
export const ConditionalInstallItemSchema = z.object({
  dependencies: DependencySchema.optional(),
  files: FileListSchema.optional(),
});

/** Module configuration from ModuleConfig.xml */
export const ModuleConfigSchema = z.object({
  moduleName: z.string(),
  moduleImage: HeaderImageSchema.optional(),
  moduleDependencies: DependencySchema.optional(),
  requiredInstallFiles: FileListSchema.optional(),
  installSteps: z.array(InstallStepSchema).optional(),
  conditionalFileInstalls: z.array(ConditionalInstallItemSchema).optional(),
});

/** Complete FOMOD data from both XML files */
export const FomodDataSchema = z.object({
  info: FomodInfoSchema.optional(),
  config: ModuleConfigSchema,
});

/** FOMOD analysis response from API */
export const FomodAnalyzeResponseSchema = z.object({
  game: z.string(),
  modId: z.number(),
  fileId: z.number(),
  hasFomod: z.boolean(),
  data: FomodDataSchema.optional(),
  cached: z.boolean(),
});

/** Dependency type - recursive interface */
export interface Dependency {
  operator?: 'And' | 'Or';
  children?: Dependency[];
  fileDependency?: { file: string; state: 'Missing' | 'Inactive' | 'Active' };
  flagDependency?: { flag: string; value: string };
  gameDependency?: { version: string };
  fommDependency?: { version: string };
}

/** FOMOD type exports */
export type GroupType = z.infer<typeof GroupTypeSchema>;
export type PluginType = z.infer<typeof PluginTypeSchema>;
export type FileState = z.infer<typeof FileStateSchema>;
export type DependencyOperator = z.infer<typeof DependencyOperatorSchema>;
export type FomodInfo = z.infer<typeof FomodInfoSchema>;
export type HeaderImage = z.infer<typeof HeaderImageSchema>;
export type FileInstall = z.infer<typeof FileInstallSchema>;
export type FolderInstall = z.infer<typeof FolderInstallSchema>;
export type FileList = z.infer<typeof FileListSchema>;
export type FileDependency = z.infer<typeof FileDependencySchema>;
export type FlagDependency = z.infer<typeof FlagDependencySchema>;
export type VersionDependency = z.infer<typeof VersionDependencySchema>;
export type ConditionFlag = z.infer<typeof ConditionFlagSchema>;
export type DependencyPattern = z.infer<typeof DependencyPatternSchema>;
export type DependencyPluginType = z.infer<typeof DependencyPluginTypeSchema>;
export type TypeDescriptor = z.infer<typeof TypeDescriptorSchema>;
export type Plugin = z.infer<typeof PluginSchema>;
export type OptionGroup = z.infer<typeof OptionGroupSchema>;
export type InstallStep = z.infer<typeof InstallStepSchema>;
export type ConditionalInstallItem = z.infer<typeof ConditionalInstallItemSchema>;
export type ModuleConfig = z.infer<typeof ModuleConfigSchema>;
export type FomodData = z.infer<typeof FomodDataSchema>;
export type FomodAnalyzeResponse = z.infer<typeof FomodAnalyzeResponseSchema>;
