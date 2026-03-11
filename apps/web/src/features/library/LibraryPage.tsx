"use client";

import { ChangeEvent, useEffect, useState } from "react";

import { AppShell } from "@/components/AppShell";
import { SectionCard } from "@/components/SectionCard";
import { StatusBadge } from "@/components/StatusBadge";
import { apiFetch, ApiError } from "@/lib/api/client";
import type { Asset } from "@/lib/api/types";

export function LibraryPage() {
  const [assets, setAssets] = useState<Asset[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [uploading, setUploading] = useState(false);
  const [assetType, setAssetType] = useState("RESUME");
  const [title, setTitle] = useState("");

  async function loadAssets() {
    try {
      const payload = await apiFetch<{ assets: Asset[] }>("/api/assets");
      setAssets(payload.assets);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "자료를 불러오지 못했습니다.");
    }
  }

  useEffect(() => {
    void loadAssets();
  }, []);

  async function handleFileChange(event: ChangeEvent<HTMLInputElement>) {
    const file = event.target.files?.[0];
    if (!file) return;
    const formData = new FormData();
    formData.append("file", file);
    formData.append("assetType", assetType);
    formData.append("title", title || file.name);

    setUploading(true);
    setError(null);
    try {
      await apiFetch("/api/assets", { method: "POST", body: formData });
      setTitle("");
      event.target.value = "";
      await loadAssets();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "업로드에 실패했습니다.");
    } finally {
      setUploading(false);
    }
  }

  async function handleDelete(assetId: string) {
    if (!window.confirm("이 자료를 즉시 삭제할까요?")) return;
    try {
      await apiFetch(`/api/assets/${assetId}`, { method: "DELETE" });
      await loadAssets();
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "자료 삭제에 실패했습니다.");
    }
  }

  return (
    <AppShell title="자료함" description="이력서, 포트폴리오, 기존 자기소개서를 업로드하고 세션에 연결합니다.">
      <div className="library-grid">
        <SectionCard title="자료 업로드" description="PDF, DOCX, MD, TXT만 지원합니다. 스캔 PDF는 텍스트 추출이 되지 않을 수 있습니다.">
          <div className="form-grid">
            <div className="form-field">
              <label htmlFor="asset-title">제목</label>
              <input id="asset-title" value={title} onChange={(event) => setTitle(event.target.value)} placeholder="예: 2026 신입 개발자 이력서" />
            </div>
            <div className="form-field">
              <label htmlFor="asset-type">자료 유형</label>
              <select id="asset-type" value={assetType} onChange={(event) => setAssetType(event.target.value)}>
                <option value="RESUME">이력서</option>
                <option value="PORTFOLIO">포트폴리오</option>
                <option value="COVER_LETTER">기존 자기소개서</option>
                <option value="NOTE">메모</option>
              </select>
            </div>
            <div className="form-field">
              <label htmlFor="asset-file">파일 선택</label>
              <input id="asset-file" type="file" accept=".pdf,.docx,.md,.txt" onChange={handleFileChange} disabled={uploading} />
            </div>
            {uploading ? <div className="helper-text">업로드 중...</div> : null}
            {error ? <div className="error-state">{error}</div> : null}
          </div>
        </SectionCard>

        <SectionCard title="보관 중인 자료" description="추출 상태와 텍스트 미리보기를 함께 제공합니다.">
          <div className="stack">
            {assets.length ? (
              assets.map((asset) => (
                <article key={asset.id} className="asset-card">
                  <div className="row-inline">
                    <h3>{asset.title}</h3>
                    <StatusBadge tone={asset.extractionStatus === "READY" ? "success" : asset.extractionStatus === "SCANNED_PDF_UNSUPPORTED" ? "warning" : "danger"}>
                      {asset.extractionStatus}
                    </StatusBadge>
                  </div>
                  <p className="helper-text">
                    {asset.assetType} · {asset.fileName}
                  </p>
                  <p>{asset.extractedText.slice(0, 320) || "텍스트 미리보기를 불러오지 못했습니다."}</p>
                  <div className="inline-actions">
                    <button type="button" className="danger-button" onClick={() => handleDelete(asset.id)}>
                      삭제
                    </button>
                  </div>
                </article>
              ))
            ) : (
              <div className="empty-state">아직 업로드한 자료가 없습니다.</div>
            )}
          </div>
        </SectionCard>
      </div>
    </AppShell>
  );
}
