"use client";

import Link from "next/link";
import { useRouter } from "next/navigation";
import { useEffect, useMemo, useState } from "react";

import { AppShell } from "@/components/AppShell";
import { SectionCard } from "@/components/SectionCard";
import { StatusBadge } from "@/components/StatusBadge";
import { apiFetch, ApiError } from "@/lib/api/client";
import type { Session, Suggestion } from "@/lib/api/types";

interface ComparePageProps {
  sessionId: string;
  initialQuestionId?: string;
}

export function ComparePage({ sessionId, initialQuestionId = "" }: ComparePageProps) {
  const router = useRouter();
  const [session, setSession] = useState<Session | null>(null);
  const [questionId, setQuestionId] = useState(initialQuestionId);
  const [suggestion, setSuggestion] = useState<Suggestion | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [applying, setApplying] = useState(false);

  useEffect(() => {
    apiFetch<{ session: Session }>(`/api/sessions/${sessionId}/workspace`)
      .then((payload) => {
        setSession(payload.session);
        setQuestionId((current) => current || payload.session.drafts[0]?.questionId || "");
      })
      .catch((err) => setError(err instanceof ApiError ? err.message : "세션을 불러오지 못했습니다."));
  }, [sessionId]);

  useEffect(() => {
    if (!questionId) return;
    apiFetch<{ suggestion: Suggestion | null }>(`/api/sessions/${sessionId}/compare?questionId=${encodeURIComponent(questionId)}`)
      .then((payload) => setSuggestion(payload.suggestion))
      .catch((err) => setError(err instanceof ApiError ? err.message : "비교 데이터를 불러오지 못했습니다."));
  }, [questionId, sessionId]);

  const activeDraft = useMemo(() => session?.drafts.find((draft) => draft.questionId === questionId) ?? null, [questionId, session]);

  async function handleApply() {
    if (!suggestion) return;
    setApplying(true);
    try {
      await apiFetch(`/api/sessions/${sessionId}/suggestions/${suggestion.id}/apply`, { method: "POST" });
      router.push(`/sessions/${sessionId}`);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "수정안 적용에 실패했습니다.");
    } finally {
      setApplying(false);
    }
  }

  return (
    <AppShell
      title="비교 보기"
      description="기존안과 수정안을 좌우 비교한 뒤 적용 여부를 결정합니다."
      actions={<Link href={`/sessions/${sessionId}`} className="secondary-button">에디터로 돌아가기</Link>}
    >
      {error ? <div className="error-state">{error}</div> : null}
      <div className="tab-strip">
        {session?.drafts.map((draft, index) => (
          <button key={draft.questionId} type="button" className={questionId === draft.questionId ? "tab-button is-active" : "tab-button"} onClick={() => setQuestionId(draft.questionId)}>
            {index + 1}. {draft.title}
          </button>
        ))}
      </div>

      {!suggestion || !activeDraft ? (
        <div className="empty-state">이 문항에는 아직 비교할 수정안이 없습니다.</div>
      ) : (
        <SectionCard
          title={activeDraft.title}
          description={activeDraft.promptText}
          actions={<StatusBadge tone="info">{suggestion.source}</StatusBadge>}
        >
          <div className="compare-grid">
            <div>
              <h3>기존안</h3>
              <div className="diff-pane">
                {buildLineDiff(activeDraft.plainText, suggestion.suggestedPlainText, "original").map((line, index) => (
                  <div key={`original-${index}`} className={line.tone ? `diff-line--${line.tone}` : undefined}>
                    {line.text || " "}
                  </div>
                ))}
              </div>
            </div>
            <div>
              <h3>수정안</h3>
              <div className="diff-pane">
                {buildLineDiff(activeDraft.plainText, suggestion.suggestedPlainText, "suggested").map((line, index) => (
                  <div key={`suggested-${index}`} className={line.tone ? `diff-line--${line.tone}` : undefined}>
                    {line.text || " "}
                  </div>
                ))}
              </div>
            </div>
          </div>
          <div className="sticky-actions">
            <button type="button" className="secondary-button" onClick={() => router.push(`/sessions/${sessionId}`)}>
              에디터로 돌아가기
            </button>
            <button type="button" className="primary-button" onClick={() => void handleApply()} disabled={applying}>
              {applying ? "적용 중..." : "수정안 적용"}
            </button>
          </div>
        </SectionCard>
      )}
    </AppShell>
  );
}

function buildLineDiff(originalText: string, suggestedText: string, side: "original" | "suggested") {
  const originalLines = originalText.split("\n");
  const suggestedLines = suggestedText.split("\n");
  const maxLength = Math.max(originalLines.length, suggestedLines.length);
  return Array.from({ length: maxLength }, (_, index) => {
    const originalLine = originalLines[index] ?? "";
    const suggestedLine = suggestedLines[index] ?? "";
    if (side === "original") {
      return {
        text: originalLine,
        tone: originalLine !== suggestedLine ? "removed" : ""
      };
    }
    return {
      text: suggestedLine,
      tone: originalLine !== suggestedLine ? "added" : ""
    };
  });
}
