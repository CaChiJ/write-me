"use client";

import { FormEvent, useEffect, useState } from "react";

import { AppShell } from "@/components/AppShell";
import { SectionCard } from "@/components/SectionCard";
import { StatusBadge } from "@/components/StatusBadge";
import { apiFetch, ApiError } from "@/lib/api/client";
import type { ProviderType, Settings } from "@/lib/api/types";

export function SettingsPage() {
  const [settings, setSettings] = useState<Settings | null>(null);
  const [providerHealth, setProviderHealth] = useState<Partial<Record<ProviderType, boolean>>>({});
  const [error, setError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);

  async function load() {
    try {
      const payload = await apiFetch<{ settings: Settings; providerHealth: Partial<Record<ProviderType, boolean>> }>("/api/settings");
      setSettings(payload.settings);
      setProviderHealth(payload.providerHealth);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "설정을 불러오지 못했습니다.");
    }
  }

  useEffect(() => {
    void load();
  }, []);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!settings) return;
    setSaving(true);
    setError(null);
    try {
      const payload = await apiFetch<{ settings: Settings }>("/api/settings", {
        method: "PUT",
        body: JSON.stringify({ defaultProvider: settings.defaultProvider })
      });
      setSettings(payload.settings);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : "설정 저장에 실패했습니다.");
    } finally {
      setSaving(false);
    }
  }

  return (
    <AppShell title="설정" description="기본 AI provider와 브리지 연결 상태를 관리합니다.">
      <div className="settings-grid">
        <SectionCard title="기본 AI provider" description="새 세션 생성 시 기본으로 선택되는 provider입니다.">
          <form className="settings-form" onSubmit={handleSubmit}>
            <div className="form-field">
              <label htmlFor="default-provider">기본 provider</label>
              <select
                id="default-provider"
                value={settings?.defaultProvider ?? "CODEX_CLI"}
                onChange={(event) =>
                  setSettings((current) =>
                    current
                      ? {
                          ...current,
                          defaultProvider: event.target.value as ProviderType
                        }
                      : current
                  )
                }
              >
                <option value="CODEX_CLI">Codex CLI</option>
                <option value="GEMINI_CLI">Gemini CLI</option>
                <option value="CLAUDE_CLI">Claude Code CLI</option>
              </select>
            </div>
            {error ? <div className="error-state">{error}</div> : null}
            <button type="submit" className="primary-button" disabled={saving}>
              {saving ? "저장 중..." : "설정 저장"}
            </button>
          </form>
        </SectionCard>

        <SectionCard title="브리지 및 provider 상태" description="호스트에서 실행한 CLIProxyAPI와 연결된 provider 모델 별칭 노출 여부입니다.">
          <div className="stack">
            {(["CODEX_CLI", "GEMINI_CLI", "CLAUDE_CLI"] as ProviderType[]).map((provider) => (
              <div key={provider} className="list-row">
                <div className="row-inline">
                  <h3>{provider}</h3>
                  <StatusBadge tone={providerHealth[provider] ? "success" : "warning"}>
                    {providerHealth[provider] ? "사용 가능" : "미연결"}
                  </StatusBadge>
                </div>
                <p className="helper-text">브리지 health는 `scripts/bridge/doctor.sh`에서도 별도 확인할 수 있습니다.</p>
              </div>
            ))}
          </div>
        </SectionCard>
      </div>
    </AppShell>
  );
}
