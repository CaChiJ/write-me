"use client";

import Link from "next/link";
import { useEffect, useMemo, useState } from "react";

import { AppShell } from "@/components/AppShell";
import { SectionCard } from "@/components/SectionCard";
import { StatusBadge } from "@/components/StatusBadge";
import { apiFetch, ApiError } from "@/lib/api/client";
import type { ReviewReport, Session } from "@/lib/api/types";

interface ReviewPageProps {
  sessionId: string;
}

export function ReviewPage({ sessionId }: ReviewPageProps) {
  const [session, setSession] = useState<Session | null>(null);
  const [activeQuestionId, setActiveQuestionId] = useState("");
  const [review, setReview] = useState<ReviewReport | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    apiFetch<{ session: Session }>(`/api/sessions/${sessionId}/workspace`)
      .then((payload) => {
        setSession(payload.session);
        setActiveQuestionId(payload.session.drafts[0]?.questionId ?? "");
        setReview(payload.session.latestReview ?? null);
      })
      .catch((err) => setError(err instanceof ApiError ? err.message : "세션을 불러오지 못했습니다."))
      .finally(() => setLoading(false));
  }, [sessionId]);

  useEffect(() => {
    if (!session || review) return;
    void runReview();
  }, [review, session]);

  const activeDraft = useMemo(() => session?.drafts.find((draft) => draft.questionId === activeQuestionId) ?? null, [activeQuestionId, session]);

  async function runReview() {
    if (!session || !activeQuestionId) return;
    try {
      const payload = await apiFetch<{ review: ReviewReport }>(`/api/sessions/${session.id}/reviews`, {
        method: "POST",
        body: JSON.stringify({ questionId: activeQuestionId })
      });
      setReview(payload.review);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "검토 실행에 실패했습니다.");
    }
  }

  return (
    <AppShell
      title="검토 보기"
      description="질문 적합성, 근거성, 가독성을 기준으로 현재 초안을 점검합니다."
      actions={<Link href={`/sessions/${sessionId}`} className="secondary-button">에디터로 돌아가기</Link>}
    >
      {loading ? <div className="page-skeleton">검토 데이터를 불러오는 중...</div> : null}
      {error ? <div className="error-state">{error}</div> : null}

      <div className="tab-strip">
        {session?.drafts.map((draft, index) => (
          <button key={draft.questionId} type="button" className={draft.questionId === activeQuestionId ? "tab-button is-active" : "tab-button"} onClick={() => setActiveQuestionId(draft.questionId)}>
            {index + 1}. {draft.title}
          </button>
        ))}
      </div>

      <SectionCard
        title={activeDraft?.title ?? "검토 결과"}
        description={activeDraft?.promptText ?? "문항을 선택하세요."}
        actions={
          <div className="inline-actions">
            {review ? <StatusBadge tone={review.readyToSubmit ? "success" : "warning"}>{review.readyToSubmit ? "제출 준비 가능" : "추가 수정 필요"}</StatusBadge> : null}
            <button type="button" className="secondary-button" onClick={() => void runReview()}>
              다시 검토
            </button>
          </div>
        }
      >
        {!review ? (
          <div className="empty-state">검토 결과가 아직 없습니다.</div>
        ) : (
          <div className="stack">
            <div className="row-inline">
              <StatusBadge tone="info">질문 적합성 {review.priorityScores.question_fit ?? 0}</StatusBadge>
              <StatusBadge tone="info">근거성 {review.priorityScores.evidence ?? 0}</StatusBadge>
              <StatusBadge tone="info">가독성 {review.priorityScores.readability ?? 0}</StatusBadge>
            </div>

            <div className="two-column">
              <div className="stack">
                <h3>차단 이슈</h3>
                {review.blockingItems.length ? (
                  review.blockingItems.map((item) => (
                    <div key={item} className="review-item">
                      {item}
                    </div>
                  ))
                ) : (
                  <div className="empty-state">차단 이슈가 없습니다.</div>
                )}
              </div>
              <div className="stack">
                <h3>우선 수정 액션</h3>
                {review.topActions.length ? (
                  review.topActions.map((item) => (
                    <div key={item} className="review-item">
                      {item}
                    </div>
                  ))
                ) : (
                  <div className="empty-state">즉시 권장하는 수정이 없습니다.</div>
                )}
              </div>
            </div>

            <div className="stack">
              <h3>문항별 소견</h3>
              {review.questionFindings.length ? (
                review.questionFindings.map((finding, index) => (
                  <div key={`${finding.questionId ?? index}`} className="review-item">
                    <pre>{JSON.stringify(finding, null, 2)}</pre>
                  </div>
                ))
              ) : (
                <div className="empty-state">문항별 소견이 없습니다.</div>
              )}
            </div>

            <div className="stack">
              <h3>미확정 AI 추론</h3>
              {review.unresolvedClaims.length ? (
                review.unresolvedClaims.map((claim) => (
                  <div key={claim.id} className="review-item">
                    <strong>{claim.excerpt}</strong>
                    {claim.reason ? <p className="helper-text">{claim.reason}</p> : null}
                  </div>
                ))
              ) : (
                <div className="empty-state">미확정 AI 추론 문장이 없습니다.</div>
              )}
            </div>
          </div>
        )}
      </SectionCard>
    </AppShell>
  );
}
