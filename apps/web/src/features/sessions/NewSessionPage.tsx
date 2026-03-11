"use client";

import { useRouter } from "next/navigation";
import { useEffect, useMemo, useState } from "react";

import { AppShell } from "@/components/AppShell";
import { SectionCard } from "@/components/SectionCard";
import { StatusBadge } from "@/components/StatusBadge";
import { apiFetch, ApiError } from "@/lib/api/client";
import type { ApplicationSpec, Asset, ProviderType, QuestionSpec, Settings, Session } from "@/lib/api/types";

type StartMode = "posting" | "draft" | "assets";

interface NewSessionPageProps {
  initialMode?: StartMode;
}

export function NewSessionPage({ initialMode = "posting" }: NewSessionPageProps) {
  const router = useRouter();
  const [mode, setMode] = useState<StartMode>(initialMode);
  const [provider, setProvider] = useState<ProviderType>("CODEX_CLI");
  const [postingText, setPostingText] = useState("");
  const [draftText, setDraftText] = useState("");
  const [companyName, setCompanyName] = useState("");
  const [roleName, setRoleName] = useState("");
  const [questions, setQuestions] = useState<QuestionSpec[]>([]);
  const [assets, setAssets] = useState<Asset[]>([]);
  const [selectedAssetIds, setSelectedAssetIds] = useState<string[]>([]);
  const [warnings, setWarnings] = useState<string[]>([]);
  const [parsing, setParsing] = useState(false);
  const [creating, setCreating] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    Promise.all([
      apiFetch<{ assets: Asset[] }>("/api/assets"),
      apiFetch<{ settings: Settings }>("/api/settings")
    ])
      .then(([assetPayload, settingsPayload]) => {
        setAssets(assetPayload.assets);
        setProvider(settingsPayload.settings.defaultProvider);
      })
      .catch((err) => {
        setError(err instanceof ApiError ? err.message : "초기 데이터를 불러오지 못했습니다.");
      });
  }, []);

  useEffect(() => {
    if (mode === "draft" && questions.length === 0) {
      setQuestions([
        { id: "q1", title: "기존 초안 편집", promptText: "기존 초안을 바탕으로 자기소개를 정리하세요.", charLimit: 700 }
      ]);
    }
    if (mode === "assets" && questions.length === 0) {
      setQuestions([
        { id: "q1", title: "자유 자기소개", promptText: "보유 자료를 바탕으로 자유 형식 자기소개를 작성하세요.", charLimit: 700 }
      ]);
    }
  }, [mode, questions.length]);

  const canParse = useMemo(() => mode === "posting" && postingText.trim().length >= 30, [mode, postingText]);

  async function handleParsePosting() {
    setParsing(true);
    setError(null);
    try {
      const payload = await apiFetch<{ applicationSpec: ApplicationSpec }>("/api/application-specs/parse", {
        method: "POST",
        body: JSON.stringify({ sourceText: postingText })
      });
      setCompanyName(payload.applicationSpec.companyName === "미확정" ? "" : payload.applicationSpec.companyName);
      setRoleName(payload.applicationSpec.roleName === "미확정" ? "" : payload.applicationSpec.roleName);
      setQuestions(payload.applicationSpec.questions);
      setWarnings(payload.applicationSpec.warnings);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "지원 페이지 구조화에 실패했습니다.");
    } finally {
      setParsing(false);
    }
  }

  function updateQuestion(index: number, field: keyof QuestionSpec, value: string | number) {
    setQuestions((current) =>
      current.map((question, questionIndex) => (questionIndex === index ? { ...question, [field]: value } : question))
    );
  }

  function addQuestion() {
    setQuestions((current) => [
      ...current,
      {
        id: `q${current.length + 1}`,
        title: `문항 ${current.length + 1}`,
        promptText: "",
        charLimit: 700
      }
    ]);
  }

  function toggleAsset(assetId: string) {
    setSelectedAssetIds((current) =>
      current.includes(assetId) ? current.filter((value) => value !== assetId) : [...current, assetId]
    );
  }

  async function handleCreateSession() {
    setCreating(true);
    setError(null);

    const manualSpec: ApplicationSpec = {
      id: "",
      companyName,
      roleName,
      sourceText: mode === "posting" ? postingText : draftText,
      warnings,
      questions: questions.length
        ? questions
        : [{ id: "q1", title: "자유 자기소개", promptText: "자유 형식 자기소개를 작성하세요.", charLimit: 700 }]
    };

    try {
      const payload = await apiFetch<{ session: Session }>("/api/sessions", {
        method: "POST",
        body: JSON.stringify({
          provider,
          selectedAssetIds,
          initialDraftText: mode === "draft" ? draftText : "",
          applicationSpec: manualSpec
        })
      });
      router.push(`/sessions/${payload.session.id}`);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "세션 생성에 실패했습니다.");
    } finally {
      setCreating(false);
    }
  }

  return (
    <AppShell title="새 작성 시작" description="지원 페이지, 기존 초안, 보유 자료 중 어떤 출발점이든 하나의 작성 세션으로 연결합니다.">
      <div className="stack">
        <SectionCard title="시작 방식" description="형식별 메뉴 분기 없이 시작 자료만 고릅니다.">
          <div className="toolbar-pills">
            {[
              ["posting", "지원 페이지 붙여넣기"],
              ["draft", "기존 초안 붙여넣기"],
              ["assets", "자료 선택 후 시작"]
            ].map(([value, label]) => (
              <button
                key={value}
                type="button"
                className={mode === value ? "tab-button is-active" : "tab-button"}
                onClick={() => setMode(value as StartMode)}
              >
                {label}
              </button>
            ))}
          </div>
        </SectionCard>

        <SectionCard title="기본 정보" description="회사명과 직무명은 비어 있어도 되지만, 비어 있으면 미확정 상태로 표시됩니다.">
          <div className="form-grid">
            <div className="form-field">
              <label htmlFor="provider">AI provider</label>
              <select id="provider" value={provider} onChange={(event) => setProvider(event.target.value as ProviderType)}>
                <option value="CODEX_CLI">Codex CLI</option>
                <option value="GEMINI_CLI">Gemini CLI</option>
                <option value="CLAUDE_CLI">Claude Code CLI</option>
              </select>
            </div>
            <div className="form-field">
              <label htmlFor="company-name">회사명</label>
              <input id="company-name" value={companyName} onChange={(event) => setCompanyName(event.target.value)} placeholder="예: 삼성전자" />
            </div>
            <div className="form-field">
              <label htmlFor="role-name">직무명</label>
              <input id="role-name" value={roleName} onChange={(event) => setRoleName(event.target.value)} placeholder="예: 백엔드 개발" />
            </div>
          </div>
        </SectionCard>

        {mode === "posting" ? (
          <SectionCard title="지원 페이지 텍스트" description="붙여넣은 텍스트를 기반으로 문항과 글자 수를 자동 구조화합니다.">
            <div className="stack">
              <textarea value={postingText} onChange={(event) => setPostingText(event.target.value)} placeholder="채용 페이지의 텍스트를 그대로 붙여넣으세요." />
              <div className="inline-actions">
                <button type="button" className="secondary-button" disabled={!canParse || parsing} onClick={handleParsePosting}>
                  {parsing ? "구조화 중..." : "구조화 실행"}
                </button>
              </div>
              {warnings.length ? (
                <div className="stack">
                  {warnings.map((warning) => (
                    <StatusBadge key={warning} tone="warning">
                      {warning}
                    </StatusBadge>
                  ))}
                </div>
              ) : null}
            </div>
          </SectionCard>
        ) : null}

        {mode === "draft" ? (
          <SectionCard title="기존 초안 텍스트" description="붙여넣은 초안은 첫 번째 문항에 초기값으로 들어갑니다.">
            <textarea value={draftText} onChange={(event) => setDraftText(event.target.value)} placeholder="이미 작성한 자기소개 초안을 붙여넣으세요." />
          </SectionCard>
        ) : null}

        <SectionCard title="문항 구성" description="자동 구조화 결과를 바로 수정하거나, 자유 형식 문항을 직접 추가할 수 있습니다.">
          <div className="stack">
            {questions.map((question, index) => (
              <div key={question.id} className="question-card">
                <div className="form-grid">
                  <div className="form-field">
                    <label>문항 제목</label>
                    <input value={question.title} onChange={(event) => updateQuestion(index, "title", event.target.value)} />
                  </div>
                  <div className="form-field">
                    <label>문항 본문</label>
                    <textarea value={question.promptText} onChange={(event) => updateQuestion(index, "promptText", event.target.value)} />
                  </div>
                  <div className="form-field">
                    <label>글자 수 제한</label>
                    <input
                      type="number"
                      min={1}
                      max={5000}
                      value={question.charLimit}
                      onChange={(event) => updateQuestion(index, "charLimit", Number(event.target.value))}
                    />
                  </div>
                </div>
              </div>
            ))}
            <button type="button" className="secondary-button" onClick={addQuestion}>
              문항 추가
            </button>
          </div>
        </SectionCard>

        <SectionCard title="자료 연결" description="세션에 연결된 자료는 AI 초안 생성, 편집, 검토 컨텍스트에 함께 포함됩니다.">
          <div className="stack">
            {assets.length ? (
              assets.map((asset) => (
                <label key={asset.id} className="asset-card">
                  <div className="row-inline">
                    <input type="checkbox" checked={selectedAssetIds.includes(asset.id)} onChange={() => toggleAsset(asset.id)} />
                    <div>
                      <strong>{asset.title}</strong>
                      <p className="helper-text">
                        {asset.assetType} · {asset.fileName}
                      </p>
                    </div>
                  </div>
                </label>
              ))
            ) : (
              <div className="empty-state">먼저 자료함에서 업로드한 자료가 필요합니다.</div>
            )}
          </div>
        </SectionCard>

        {error ? <div className="error-state">{error}</div> : null}

        <div className="sticky-actions">
          <button type="button" className="secondary-button" onClick={() => router.push("/dashboard")}>
            취소
          </button>
          <button type="button" className="primary-button" disabled={creating} onClick={handleCreateSession}>
            {creating ? "세션 생성 중..." : "작성 세션 만들기"}
          </button>
        </div>
      </div>
    </AppShell>
  );
}
