"use client";

import { FormEvent, useState } from "react";
import { useRouter } from "next/navigation";

import { apiFetch, ApiError } from "@/lib/api/client";

export function LoginPage() {
  const router = useRouter();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setSubmitting(true);
    setError(null);
    try {
      await apiFetch("/api/auth/login", {
        method: "POST",
        body: JSON.stringify({ email, password })
      });
      router.replace("/dashboard");
    } catch (err) {
      if (err instanceof ApiError) {
        setError(err.message);
      } else {
        setError("로그인 중 오류가 발생했습니다.");
      }
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <div className="login-shell">
      <section className="login-brand">
        <p className="page-header__eyebrow">Korean-first editorial workspace</p>
        <h1>자기소개서 작성의 근거와 편집 이력을 한 화면에서 관리합니다.</h1>
        <p className="page-header__description">
          지원 페이지를 붙여넣고, 자료를 연결한 뒤, AI 제안을 비교 적용하고, 미확정 AI 추론 문장을 확인한 뒤 최종 확정할 수 있습니다.
        </p>
        <div className="stack">
          <div className="metric-card">
            <strong>1 세션</strong>
            <span>1 지원 공고 기준으로 작성 흐름을 관리</span>
          </div>
          <div className="metric-card">
            <strong>INFERRED</strong>
            <span>자료에 없는 보강 문장은 별도 표시 후 확인 필수</span>
          </div>
        </div>
      </section>
      <section className="login-panel">
        <form className="login-form section-card" onSubmit={handleSubmit}>
          <div className="section-card__header">
            <div>
              <h2>로그인</h2>
              <p>단일 관리자 계정으로 접속합니다.</p>
            </div>
          </div>
          <div className="section-card__body">
            <div className="form-field">
              <label htmlFor="email">이메일</label>
              <input id="email" type="email" value={email} onChange={(event) => setEmail(event.target.value)} required />
            </div>
            <div className="form-field">
              <label htmlFor="password">비밀번호</label>
              <input id="password" type="password" value={password} onChange={(event) => setPassword(event.target.value)} required />
            </div>
            {error ? <div className="error-state">{error}</div> : null}
            <button type="submit" className="primary-button" disabled={submitting}>
              {submitting ? "로그인 중..." : "로그인"}
            </button>
            <p className="helper-text">기본 계정은 `.env`의 `ADMIN_EMAIL`, `ADMIN_PASSWORD`로 초기화됩니다.</p>
          </div>
        </form>
      </section>
    </div>
  );
}
