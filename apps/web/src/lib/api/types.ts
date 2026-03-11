export type ProviderType = "CODEX_CLI" | "GEMINI_CLI" | "CLAUDE_CLI";
export type CountMode = "INCLUDE_SPACES" | "EXCLUDE_SPACES";
export type ClaimTagStatus = "SUPPORTED" | "INFERRED" | "RESOLVED";
export type SessionStatus = "DRAFT" | "FINALIZED";

export interface User {
  id: string;
  email: string;
  createdAt: string;
}

export interface Settings {
  defaultProvider: ProviderType;
  updatedAt: string;
}

export interface Asset {
  id: string;
  assetType: string;
  title: string;
  fileName: string;
  mimeType: string;
  extractionStatus: string;
  extractedText: string;
  createdAt: string;
}

export interface QuestionSpec {
  id: string;
  title: string;
  promptText: string;
  charLimit: number;
}

export interface ApplicationSpec {
  id: string;
  companyName: string;
  roleName: string;
  sourceText: string;
  warnings: string[];
  questions: QuestionSpec[];
  createdAt?: string;
}

export interface ClaimTag {
  id: string;
  start: number;
  end: number;
  status: ClaimTagStatus;
  excerpt: string;
  reason?: string;
  sourceAssetIds?: string[];
  resolved: boolean;
}

export interface DocumentBlock {
  id: string;
  text: string;
  claimTags: ClaimTag[];
}

export interface DocumentPayload {
  version: number;
  blocks: DocumentBlock[];
}

export interface DraftQuestion {
  id: string;
  questionId: string;
  title: string;
  promptText: string;
  charLimit: number;
  document: DocumentPayload;
  plainText: string;
  inferredCount: number;
  resolvedInferredCount: number;
  updatedAt: string;
}

export interface ChatMessage {
  id: string;
  questionId: string;
  role: string;
  messageType: string;
  content: string;
  meta: Record<string, unknown>;
  createdAt: string;
}

export interface Suggestion {
  id: string;
  questionId: string;
  source: string;
  scope: string;
  rationale: string[];
  originalDocument: DocumentPayload;
  originalPlainText: string;
  suggestedDocument: DocumentPayload;
  suggestedPlainText: string;
  status: string;
  createdAt: string;
}

export interface ReviewReport {
  id: string;
  sessionId: string;
  questionId?: string;
  priorityScores: Record<string, number>;
  blockingItems: string[];
  topActions: string[];
  questionFindings: Record<string, unknown>[];
  unresolvedClaims: ClaimTag[];
  readyToSubmit: boolean;
  raw: Record<string, unknown>;
  createdAt: string;
}

export interface Session {
  id: string;
  title: string;
  status: SessionStatus;
  currentProvider: ProviderType;
  applyMode: string;
  reviewMode: string;
  autoReview: boolean;
  autoApply: boolean;
  countMode: CountMode;
  applicationSpec: ApplicationSpec;
  assets: Asset[];
  drafts: DraftQuestion[];
  chatMessages: ChatMessage[];
  latestReview?: ReviewReport | null;
  pendingSuggestions: Record<string, Suggestion | undefined>;
  providerHealth: Partial<Record<ProviderType, boolean>>;
  createdAt: string;
  updatedAt: string;
}

export interface DashboardSummary {
  recentSessions: {
    id: string;
    title: string;
    companyName: string;
    roleName: string;
    status: SessionStatus;
    updatedAt: string;
    unresolvedCount: number;
  }[];
  assetCount: number;
  readyCount: number;
}

export interface ApiErrorPayload {
  error: {
    code: string;
    message: string;
  };
}
