"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useCallback, useEffect, useMemo, useState } from "react";

import { AppShell } from "@/components/AppShell";
import { SectionCard } from "@/components/SectionCard";
import { StatusBadge } from "@/components/StatusBadge";
import { apiFetch, ApiError } from "@/lib/api/client";
import type { ChatMessage, ClaimTag, CountMode, DraftQuestion, ReviewReport, Session, Suggestion } from "@/lib/api/types";

const toolNames = ["다른 소재 찾기", "질문 적합성 개선", "가독성 개선", "분량 맞추기", "톤 조정"] as const;

interface SessionEditorPageProps {
  sessionId: string;
}

export function SessionEditorPage({ sessionId }: SessionEditorPageProps) {
  const router = useRouter();
  const [session, setSession] = useState<Session | null>(null);
  const [activeQuestionId, setActiveQuestionId] = useState("");
  const [editorText, setEditorText] = useState("");
  const [baselineText, setBaselineText] = useState("");
  const [saveState, setSaveState] = useState<"idle" | "dirty" | "saving" | "error">("idle");
  const [scope, setScope] = useState("WHOLE_DOCUMENT");
  const [chatInput, setChatInput] = useState("");
  const [review, setReview] = useState<ReviewReport | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [chatError, setChatError] = useState<string | null>(null);
  const [composing, setComposing] = useState(false);
  const [busy, setBusy] = useState(false);

  const loadWorkspace = useCallback(async () => {
    try {
      const payload = await apiFetch<{ session: Session }>(`/api/sessions/${sessionId}/workspace`);
      setSession(payload.session);
      setReview(payload.session.latestReview ?? null);
      const fallbackQuestionId = payload.session.drafts[0]?.questionId ?? "";
      setActiveQuestionId((current) => current || fallbackQuestionId);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "세션을 불러오지 못했습니다.");
    } finally {
      setLoading(false);
    }
  }, [sessionId]);

  useEffect(() => {
    void loadWorkspace();
  }, [loadWorkspace]);

  const activeQuestion = useMemo(
    () => session?.drafts.find((draft) => draft.questionId === activeQuestionId) ?? session?.drafts[0] ?? null,
    [activeQuestionId, session]
  );
  const activeMessages = useMemo(
    () => session?.chatMessages.filter((message) => message.questionId === activeQuestion?.questionId) ?? [],
    [activeQuestion?.questionId, session?.chatMessages]
  );
  const activeSuggestion = activeQuestion ? session?.pendingSuggestions?.[activeQuestion.questionId] ?? null : null;
  const unresolvedCount = useMemo(
    () => session?.drafts.reduce((sum, draft) => sum + Math.max(draft.inferredCount - draft.resolvedInferredCount, 0), 0) ?? 0,
    [session?.drafts]
  );

  useEffect(() => {
    if (!activeQuestion) return;
    setEditorText(activeQuestion.plainText);
    setBaselineText(activeQuestion.plainText);
    setSaveState("idle");
  }, [activeQuestion?.plainText, activeQuestion?.questionId]);

  const persistDraft = useCallback(
    async (withReview: boolean) => {
      if (!activeQuestion || !session) return;
      if (editorText === baselineText) return;
      setSaveState("saving");
      try {
        const payload = await apiFetch<{ draft: DraftQuestion }>(
          `/api/sessions/${session.id}/questions/${activeQuestion.questionId}/draft`,
          {
            method: "PATCH",
            body: JSON.stringify({ plainText: editorText, reason: withReview ? "auto_review_cycle" : "autosave" })
          }
        );
        setSession((current) =>
          current
            ? {
                ...current,
                drafts: current.drafts.map((draft) => (draft.questionId === payload.draft.questionId ? payload.draft : draft))
              }
            : current
        );
        setBaselineText(payload.draft.plainText);
        setSaveState("idle");
        if (withReview && session.autoReview) {
          const reviewPayload = await apiFetch<{ review: ReviewReport }>(`/api/sessions/${session.id}/reviews`, {
            method: "POST",
            body: JSON.stringify({ questionId: activeQuestion.questionId })
          });
          setReview(reviewPayload.review);
        }
      } catch (err) {
        setSaveState("error");
        setError(err instanceof ApiError ? err.message : "초안 저장에 실패했습니다.");
      }
    },
    [activeQuestion, baselineText, editorText, session]
  );

  useEffect(() => {
    if (!activeQuestion || !session) return;
    if (composing) return;
    if (editorText === baselineText) return;

    setSaveState("dirty");
    const saveTimer = window.setTimeout(() => void persistDraft(false), 2000);
    const reviewTimer = session.autoReview ? window.setTimeout(() => void persistDraft(true), 3000) : undefined;
    return () => {
      window.clearTimeout(saveTimer);
      if (reviewTimer) window.clearTimeout(reviewTimer);
    };
  }, [activeQuestion, baselineText, composing, editorText, persistDraft, session]);

  async function handleUpdateSession(patch: Partial<Session>) {
    if (!session) return;
    try {
      const payload = await apiFetch<{ session: Session }>(`/api/sessions/${session.id}`, {
        method: "PATCH",
        body: JSON.stringify({
          currentProvider: patch.currentProvider ?? session.currentProvider,
          autoReview: patch.autoReview ?? session.autoReview,
          autoApply: patch.autoApply ?? session.autoApply,
          countMode: patch.countMode ?? session.countMode
        })
      });
      setSession(payload.session);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "세션 설정 변경에 실패했습니다.");
    }
  }

  async function handleGenerateDraft() {
    if (!session || !activeQuestion) return;
    setBusy(true);
    try {
      const payload = await apiFetch<{ suggestion: Suggestion }>(`/api/sessions/${session.id}/generate`, {
        method: "POST",
        body: JSON.stringify({ questionId: activeQuestion.questionId })
      });
      if (payload.suggestion) {
        setSession((current) =>
          current
            ? {
                ...current,
                pendingSuggestions: {
                  ...current.pendingSuggestions,
                  [activeQuestion.questionId]: payload.suggestion
                }
              }
            : current
        );
      }
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "AI 초안 생성에 실패했습니다.");
    } finally {
      setBusy(false);
    }
  }

  async function handleRunTool(toolName: string) {
    if (!session || !activeQuestion) return;
    setBusy(true);
    setChatError(null);
    try {
      const payload = await apiFetch<{ message: ChatMessage; suggestion?: Suggestion }>(`/api/sessions/${session.id}/tools/${encodeURIComponent(toolName)}/execute`, {
        method: "POST",
        body: JSON.stringify({ questionId: activeQuestion.questionId, scope })
      });
      setSession((current) =>
        current
          ? {
              ...current,
              chatMessages: [...current.chatMessages, payload.message],
              pendingSuggestions: payload.suggestion
                ? {
                    ...current.pendingSuggestions,
                    [activeQuestion.questionId]: payload.suggestion
                  }
                : current.pendingSuggestions
            }
          : current
      );
    } catch (err) {
      setChatError(err instanceof ApiError ? err.message : "도구 실행에 실패했습니다.");
    } finally {
      setBusy(false);
    }
  }

  async function handleSendChat() {
    if (!session || !activeQuestion || !chatInput.trim()) return;
    setBusy(true);
    setChatError(null);
    try {
      const payload = await apiFetch<{ message: ChatMessage; suggestion?: Suggestion }>(`/api/sessions/${session.id}/chat/turns`, {
        method: "POST",
        body: JSON.stringify({ questionId: activeQuestion.questionId, scope, message: chatInput })
      });
      setSession((current) =>
        current
          ? {
              ...current,
              chatMessages: [...current.chatMessages, payload.message],
              pendingSuggestions: payload.suggestion
                ? {
                    ...current.pendingSuggestions,
                    [activeQuestion.questionId]: payload.suggestion
                  }
                : current.pendingSuggestions
            }
          : current
      );
      setChatInput("");
    } catch (err) {
      setChatError(err instanceof ApiError ? err.message : "AI 채팅 요청에 실패했습니다.");
    } finally {
      setBusy(false);
    }
  }

  async function handleRunReview() {
    if (!session || !activeQuestion) return;
    setBusy(true);
    try {
      const payload = await apiFetch<{ review: ReviewReport }>(`/api/sessions/${session.id}/reviews`, {
        method: "POST",
        body: JSON.stringify({ questionId: activeQuestion.questionId })
      });
      setReview(payload.review);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "검토 실행에 실패했습니다.");
    } finally {
      setBusy(false);
    }
  }

  async function handleResolveClaim(claim: ClaimTag) {
    if (!session || !activeQuestion) return;
    try {
      const payload = await apiFetch<{ draft: DraftQuestion }>(
        `/api/sessions/${session.id}/questions/${activeQuestion.questionId}/claims/${claim.id}/resolve`,
        { method: "POST" }
      );
      setSession((current) =>
        current
          ? {
              ...current,
              drafts: current.drafts.map((draft) => (draft.questionId === payload.draft.questionId ? payload.draft : draft))
            }
          : current
      );
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "AI 추론 문장 확인 처리에 실패했습니다.");
    }
  }

  async function handleFinalize() {
    if (!session) return;
    if (unresolvedCount > 0) {
      setError("AI 추론 문장을 모두 확인한 뒤 최종 확정할 수 있습니다.");
      return;
    }
    try {
      await apiFetch(`/api/sessions/${session.id}/finalize`, { method: "POST" });
      await loadWorkspace();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "최종 확정에 실패했습니다.");
    }
  }

  if (loading) {
    return (
      <AppShell title="작성 세션" description="세션을 불러오는 중입니다.">
        <div className="page-skeleton">세션 불러오는 중...</div>
      </AppShell>
    );
  }

  if (!session || !activeQuestion) {
    return (
      <AppShell title="작성 세션" description="세션을 찾을 수 없습니다.">
        <div className="error-state">{error ?? "세션이 존재하지 않습니다."}</div>
      </AppShell>
    );
  }

  return (
    <AppShell
      title={session.title}
      description={`${session.applicationSpec.companyName} · ${session.applicationSpec.roleName}`}
      actions={
        <div className="inline-actions">
          <StatusBadge tone={session.status === "FINALIZED" ? "success" : "neutral"}>{session.status === "FINALIZED" ? "최종 확정" : "작성 중"}</StatusBadge>
          <StatusBadge tone={saveState === "error" ? "danger" : saveState === "saving" ? "info" : "neutral"}>
            {saveState === "error" ? "저장 실패" : saveState === "saving" ? "저장 중" : saveState === "dirty" ? "수정됨" : "저장됨"}
          </StatusBadge>
        </div>
      }
    >
      {error ? <div className="error-state">{error}</div> : null}
      <div className="workspace-bar">
        <div className="workspace-controls">
          <select value={session.currentProvider} onChange={(event) => void handleUpdateSession({ currentProvider: event.target.value as Session["currentProvider"] })}>
            <option value="CODEX_CLI">Codex CLI</option>
            <option value="GEMINI_CLI">Gemini CLI</option>
            <option value="CLAUDE_CLI">Claude Code CLI</option>
          </select>
          <select value={session.countMode} onChange={(event) => void handleUpdateSession({ countMode: event.target.value as CountMode })}>
            <option value="INCLUDE_SPACES">공백 포함</option>
            <option value="EXCLUDE_SPACES">공백 제외</option>
          </select>
          <label className="row-inline">
            <input type="checkbox" checked={session.autoReview} onChange={(event) => void handleUpdateSession({ autoReview: event.target.checked })} />
            자동 검토
          </label>
          <label className="row-inline">
            <input type="checkbox" checked={session.autoApply} onChange={(event) => void handleUpdateSession({ autoApply: event.target.checked })} />
            자동 반영
          </label>
          <select value={scope} onChange={(event) => setScope(event.target.value)}>
            <option value="WHOLE_DOCUMENT">전체 문서</option>
            <option value="SELECTION_ONLY">선택 문단</option>
          </select>
        </div>
        <div className="inline-actions">
          {session.providerHealth[session.currentProvider] ? <StatusBadge tone="success">bridge 연결됨</StatusBadge> : <StatusBadge tone="warning">bridge 미연결</StatusBadge>}
          {unresolvedCount > 0 ? <StatusBadge tone="warning">미확정 추론 {unresolvedCount}</StatusBadge> : <StatusBadge tone="success">추론 확인 완료</StatusBadge>}
        </div>
      </div>

      <div className="tab-strip">
        {session.drafts.map((draft, index) => (
          <button
            key={draft.questionId}
            type="button"
            className={activeQuestion.questionId === draft.questionId ? "tab-button is-active" : "tab-button"}
            onClick={() => setActiveQuestionId(draft.questionId)}
          >
            {index + 1}. {draft.title}
            {draft.inferredCount > draft.resolvedInferredCount ? ` (${draft.inferredCount - draft.resolvedInferredCount})` : ""}
          </button>
        ))}
      </div>

      <div className="split-grid">
        <SectionCard title="자료 패널" description="선택된 자료의 텍스트 추출 결과를 확인합니다.">
          <div className="stack">
            {session.assets.length ? (
              session.assets.map((asset) => (
                <article key={asset.id} className="asset-card">
                  <div className="row-inline">
                    <strong>{asset.title}</strong>
                    <StatusBadge tone={asset.extractionStatus === "READY" ? "success" : "warning"}>{asset.extractionStatus}</StatusBadge>
                  </div>
                  <p className="helper-text">{asset.extractedText.slice(0, 180) || "추출 텍스트 없음"}</p>
                </article>
              ))
            ) : (
              <div className="empty-state">연결된 자료가 없습니다.</div>
            )}
          </div>
        </SectionCard>

        <div className="workspace-panel">
          <SectionCard
            title={activeQuestion.title}
            description={activeQuestion.promptText}
            actions={
              <div className="inline-actions">
                <StatusBadge tone={activeQuestion.charLimit > 0 && countText(editorText, session.countMode) > activeQuestion.charLimit ? "warning" : "neutral"}>
                  {countText(editorText, session.countMode)} / {activeQuestion.charLimit || "-"}자
                </StatusBadge>
                <button type="button" className="secondary-button" onClick={() => void handleGenerateDraft()} disabled={busy}>
                  AI 초안 생성
                </button>
              </div>
            }
          >
            <div className="stack">
              <textarea
                className="editor-area"
                value={editorText}
                onChange={(event) => setEditorText(event.target.value)}
                onBlur={() => void persistDraft(false)}
                onCompositionStart={() => setComposing(true)}
                onCompositionEnd={() => setComposing(false)}
              />
              <div className="editor-meta">
                <span className="helper-text">마지막 저장 기준: {new Date(activeQuestion.updatedAt).toLocaleString("ko-KR")}</span>
                <button type="button" className="secondary-button" onClick={() => void handleRunReview()} disabled={busy}>
                  빠른 검토 실행
                </button>
              </div>
              {collectUnresolvedClaims(activeQuestion).length ? (
                <div className="inferred-list">
                  {collectUnresolvedClaims(activeQuestion).map((claim) => (
                    <div key={claim.id} className="inferred-item">
                      <div className="row-inline">
                        <StatusBadge tone="warning">AI 추론 추가</StatusBadge>
                        <strong>{claim.excerpt}</strong>
                      </div>
                      {claim.reason ? <p className="helper-text">{claim.reason}</p> : null}
                      <div className="inline-actions">
                        <button type="button" className="secondary-button" onClick={() => void handleResolveClaim(claim)}>
                          사실 확인 완료
                        </button>
                      </div>
                    </div>
                  ))}
                </div>
              ) : (
                <div className="empty-state">현재 문항에는 미확정 AI 추론 문장이 없습니다.</div>
              )}
            </div>
          </SectionCard>

          {review ? (
            <SectionCard title="빠른 검토" description="현재 문항 기준 자동/수동 검토 결과입니다.">
              <div className="review-list">
                <div className="row-inline">
                  <StatusBadge tone={review.readyToSubmit ? "success" : "warning"}>{review.readyToSubmit ? "제출 준비 가능" : "추가 수정 필요"}</StatusBadge>
                  <StatusBadge tone="info">질문 적합성 {review.priorityScores.question_fit ?? 0}</StatusBadge>
                  <StatusBadge tone="info">근거성 {review.priorityScores.evidence ?? 0}</StatusBadge>
                  <StatusBadge tone="info">가독성 {review.priorityScores.readability ?? 0}</StatusBadge>
                </div>
                {review.blockingItems.length ? (
                  <div className="stack">
                    {review.blockingItems.map((item) => (
                      <div key={item} className="review-item">
                        {item}
                      </div>
                    ))}
                  </div>
                ) : (
                  <div className="empty-state">현재 문항에는 차단 이슈가 없습니다.</div>
                )}
              </div>
            </SectionCard>
          ) : null}
        </div>

        <SectionCard title="도구 및 AI 채팅" description="도구를 누르면 프롬프트가 채팅 흐름에 바로 연결됩니다.">
          <div className="stack">
            <div className="toolbar-pills">
              {toolNames.map((toolName) => (
                <button key={toolName} type="button" className="tool-pill" onClick={() => void handleRunTool(toolName)} disabled={busy}>
                  {toolName}
                </button>
              ))}
            </div>
            {activeSuggestion ? (
              <article className="suggestion-card">
                <div className="row-inline">
                  <strong>수정 제안 준비됨</strong>
                  <StatusBadge tone="info">{activeSuggestion.source}</StatusBadge>
                </div>
                <p className="helper-text">{activeSuggestion.rationale.join(" · ") || "수정 이유가 제공되지 않았습니다."}</p>
                <Link href={`/sessions/${session.id}/compare?questionId=${activeQuestion.questionId}`} className="secondary-button">
                  비교 보기
                </Link>
              </article>
            ) : null}
            <div className="message-list">
              {activeMessages.length ? (
                activeMessages.map((message) => (
                  <div key={message.id} className={`message-bubble ${message.role === "assistant" ? "message-bubble--assistant" : "message-bubble--user"}`}>
                    <strong>{message.role === "assistant" ? "AI" : "나"}</strong>
                    <p>{message.content}</p>
                  </div>
                ))
              ) : (
                <div className="empty-state">아직 대화가 없습니다.</div>
              )}
            </div>
            <textarea value={chatInput} onChange={(event) => setChatInput(event.target.value)} placeholder="AI에게 수정 방향이나 조언을 요청하세요." />
            {chatError ? <div className="error-state">{chatError}</div> : null}
            <button type="button" className="primary-button" onClick={() => void handleSendChat()} disabled={busy || !chatInput.trim()}>
              전송
            </button>
          </div>
        </SectionCard>
      </div>

      <div className="sticky-actions">
        <Link href={`/sessions/${session.id}/review`} className="secondary-button">
          검토 열기
        </Link>
        <Link
          href={activeSuggestion ? `/sessions/${session.id}/compare?questionId=${activeQuestion.questionId}` : "#"}
          className="secondary-button"
          onClick={(event) => {
            if (!activeSuggestion) {
              event.preventDefault();
            }
          }}
        >
          비교 보기
        </Link>
        <button type="button" className="primary-button" disabled={unresolvedCount > 0} onClick={() => void handleFinalize()}>
          최종 확정
        </button>
      </div>
    </AppShell>
  );
}

function collectUnresolvedClaims(draft: DraftQuestion): ClaimTag[] {
  return draft.document.blocks.flatMap((block) => block.claimTags).filter((claim) => claim.status === "INFERRED" && !claim.resolved);
}

function countText(text: string, mode: CountMode): number {
  if (mode === "EXCLUDE_SPACES") {
    return text.replace(/\s+/g, "").length;
  }
  return text.length;
}
