"use client";

import Link from "next/link";
import { useEffect, useState } from "react";

import { AppShell } from "@/components/AppShell";
import { SectionCard } from "@/components/SectionCard";
import { StatusBadge } from "@/components/StatusBadge";
import { apiFetch, ApiError } from "@/lib/api/client";
import type { DashboardSummary } from "@/lib/api/types";

export function DashboardPage() {
  const [summary, setSummary] = useState<DashboardSummary | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    apiFetch<DashboardSummary>("/api/dashboard")
      .then(setSummary)
      .catch((err) => {
        setError(err instanceof ApiError ? err.message : "대시보드를 불러오지 못했습니다.");
      });
  }, []);

  return (
    <AppShell title="대시보드" description="새 작성 시작점과 최근 작업, 자료 현황을 한 번에 확인합니다.">
      {error ? <div className="error-state">{error}</div> : null}

      <div className="dashboard-grid">
        <SectionCard title="새 작성 시작" description="지원 페이지, 기존 초안, 보유 자료 중 어떤 출발점이든 같은 세션 구조로 연결합니다.">
          <div className="start-cards">
            <Link href="/sessions/new" className="start-card">
              <h3>지원 페이지 붙여넣고 시작</h3>
              <p>공고 텍스트를 붙여넣고 문항/글자 수를 자동 구조화합니다.</p>
            </Link>
            <Link href="/sessions/new?mode=draft" className="start-card">
              <h3>기존 초안 붙여넣고 시작</h3>
              <p>이미 작성한 초안을 기반으로 질문 적합성과 가독성을 편집합니다.</p>
            </Link>
            <Link href="/sessions/new?mode=assets" className="start-card">
              <h3>자료 선택 후 시작</h3>
              <p>이력서, 포트폴리오, 기존 자소서를 연결해 자유 형식 초안을 생성합니다.</p>
            </Link>
          </div>
        </SectionCard>

        <SectionCard title="현황" description="MVP 기준의 핵심 운영 지표입니다.">
          <div className="metric-grid">
            <div className="metric-card">
              <span>자료 개수</span>
              <strong>{summary?.assetCount ?? "-"}</strong>
            </div>
            <div className="metric-card">
              <span>제출 준비 완료</span>
              <strong>{summary?.readyCount ?? "-"}</strong>
            </div>
            <div className="metric-card">
              <span>최근 세션</span>
              <strong>{summary?.recentSessions.length ?? "-"}</strong>
            </div>
          </div>
        </SectionCard>

        <SectionCard title="최근 작업" description="세션 상태와 미확정 AI 추론 잔여 개수를 함께 보여줍니다.">
          <div className="card-list">
            {summary?.recentSessions.length ? (
              summary.recentSessions.map((session) => (
                <Link key={session.id} href={`/sessions/${session.id}`} className="list-row">
                  <div className="row-inline">
                    <h3>{session.title}</h3>
                    <StatusBadge tone={session.status === "FINALIZED" ? "success" : "neutral"}>{session.status === "FINALIZED" ? "최종 확정" : "작성 중"}</StatusBadge>
                    {session.unresolvedCount > 0 ? <StatusBadge tone="warning">미확정 {session.unresolvedCount}</StatusBadge> : null}
                  </div>
                  <p>
                    {session.companyName} · {session.roleName}
                  </p>
                  <p className="helper-text">{new Date(session.updatedAt).toLocaleString("ko-KR")} 업데이트</p>
                </Link>
              ))
            ) : (
              <div className="empty-state">아직 생성한 세션이 없습니다. 새 작성 시작 카드를 눌러 첫 세션을 만드세요.</div>
            )}
          </div>
        </SectionCard>
      </div>
    </AppShell>
  );
}
