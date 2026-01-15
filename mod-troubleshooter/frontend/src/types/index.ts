export type GameId = 'skyrim' | 'stardew' | 'cyberpunk';

export interface GameConfig {
  id: GameId;
  label: string;
  filename: string;
}

export const GAMES: GameConfig[] = [
  { id: 'skyrim', label: 'Skyrim Special Edition', filename: 'skyrim.json' },
  { id: 'stardew', label: 'Stardew Valley', filename: 'stardew.json' },
  { id: 'cyberpunk', label: 'Cyberpunk 2077', filename: 'cyberpunk.json' },
];

/** Simplified mod for viewer mode */
export interface ViewerMod {
  modId?: number;
  fileId?: number;
  name: string;
  summary?: string;
  description?: string;
  pictureUrl?: string;
  nexusUrl?: string;
  version?: string;
  category?: string;
  optional?: boolean;
  uploader?: {
    name: string;
    avatar?: string;
  };
  author?: string;
  collectionName?: string;
}

/** Simplified collection for viewer mode */
export interface ViewerCollection {
  id: string;
  name: string;
  slug?: string;
  summary?: string;
  author?: {
    name: string;
  };
  tileImage?: {
    url: string;
  };
  modCount: number;
  mods: ViewerMod[];
}

/** Collections data structure for JSON loading */
export interface CollectionsData {
  collections: ViewerCollection[];
  totalMods: number;
  fetchedAt?: string;
}

export type {
  ApiResponse,
  User,
  Game,
  Image,
  ModCategory,
  Mod,
  ModFile,
  ModFileReference,
  ExternalResource,
  Revision,
  RevisionDetails,
  Collection,
  // FOMOD types
  Dependency,
  GroupType,
  PluginType,
  FileState,
  DependencyOperator,
  FomodInfo,
  HeaderImage,
  FileInstall,
  FolderInstall,
  FileList,
  FileDependency,
  FlagDependency,
  VersionDependency,
  ConditionFlag,
  DependencyPattern,
  DependencyPluginType,
  TypeDescriptor,
  Plugin,
  OptionGroup,
  InstallStep,
  ConditionalInstallItem,
  ModuleConfig,
  FomodData,
  FomodAnalyzeResponse,
  // Load Order types
  LoadOrderPluginType,
  PluginFlags,
  IssueType,
  IssueSeverity,
  LoadOrderIssue,
  LoadOrderPluginInfo,
  LoadOrderStats,
  LoadOrderAnalysisResult,
  LoadOrderAnalyzeResponse,
  // Conflict Detection types
  FileType,
  ConflictType,
  ConflictSeverity,
  ModFileConflict,
  Conflict,
  ConflictStats,
  ModConflictSummary,
  ConflictAnalysisResult,
  ConflictAnalyzeResponse,
  // Quota types
  Quota,
} from './api.ts';

export {
  ApiResponseSchema,
  UserSchema,
  GameSchema,
  ImageSchema,
  ModCategorySchema,
  ModSchema,
  ModFileSchema,
  ModFileReferenceSchema,
  ExternalResourceSchema,
  RevisionSchema,
  RevisionDetailsSchema,
  CollectionSchema,
  // FOMOD schemas
  GroupTypeSchema,
  PluginTypeSchema,
  FileStateSchema,
  DependencyOperatorSchema,
  FomodInfoSchema,
  HeaderImageSchema,
  FileInstallSchema,
  FolderInstallSchema,
  FileListSchema,
  FileDependencySchema,
  FlagDependencySchema,
  VersionDependencySchema,
  DependencySchema,
  ConditionFlagSchema,
  DependencyPatternSchema,
  DependencyPluginTypeSchema,
  TypeDescriptorSchema,
  PluginSchema,
  OptionGroupSchema,
  InstallStepSchema,
  ConditionalInstallItemSchema,
  ModuleConfigSchema,
  FomodDataSchema,
  FomodAnalyzeResponseSchema,
  // Load Order schemas
  LoadOrderPluginTypeSchema,
  PluginFlagsSchema,
  IssueTypeSchema,
  IssueSeveritySchema,
  LoadOrderIssueSchema,
  LoadOrderPluginInfoSchema,
  LoadOrderStatsSchema,
  LoadOrderAnalysisResultSchema,
  LoadOrderAnalyzeResponseSchema,
  // Conflict Detection schemas
  FileTypeSchema,
  ConflictTypeSchema,
  ConflictSeveritySchema,
  ModFileSchema2,
  ConflictSchema,
  ConflictStatsSchema,
  ModConflictSummarySchema,
  ConflictAnalysisResultSchema,
  ConflictAnalyzeResponseSchema,
  // Quota schema
  QuotaSchema,
} from './api.ts';
