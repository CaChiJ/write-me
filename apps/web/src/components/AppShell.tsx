"use client";

import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { useEffect, useMemo, useState } from "react";

import { apiFetch, ApiError, isUnauthorized } from "@/lib/api/client";
import type { User } from "@/lib/api/types";

interface AppShellProps {
  title: string;
  description?: string;
  actions?: React.ReactNode;
  children: React.ReactNode;
}

const navItems = [
  { href: "/dashboard", label: "대시보드" },
  { href: "/sessions/new", label: "새 작성" },
  { href: "/library", label: "자료함" },
  { href: "/settings", label: "설정" }
];

export function AppShell({ title, description, actions, children }: AppShellProps) {
  const pathname = usePathname();
  const router = useRouter();
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let active = true;
    apiFetch<{ user: User }>("/api/me")
      .then((payload) => {
        if (!active) return;
        setUser(payload.user);
      })
      .catch((error) => {
        if (!active) return;
        if (isUnauthorized(error)) {
          router.replace("/login");
          return;
        }
        console.error(error);
      })
      .finally(() => {
        if (active) {
          setLoading(false);
        }
      });

    return () => {
      active = false;
    };
  }, [router]);

  const shellClassName = useMemo(() => (pathname?.startsWith("/sessions/") ? "app-shell app-shell--workspace" : "app-shell"), [pathname]);

  async function handleLogout() {
    try {
      await apiFetch("/api/auth/logout", { method: "POST" });
    } catch (error) {
      if (!(error instanceof ApiError)) {
        console.error(error);
      }
    } finally {
      router.replace("/login");
    }
  }

  return (
    <div className={shellClassName}>
      <aside className="nav-rail">
        <div className="nav-rail__brand">
          <span className="nav-rail__logo">W</span>
          <span className="nav-rail__name">Write Me</span>
        </div>
        <nav className="nav-rail__links">
          {navItems.map((item) => (
            <Link key={item.href} href={item.href} className={pathname === item.href ? "nav-link is-active" : "nav-link"}>
              {item.label}
            </Link>
          ))}
        </nav>
        <div className="nav-rail__footer">
          <span className="nav-rail__user">{user?.email ?? "로그인 확인 중"}</span>
          <button type="button" className="ghost-button" onClick={handleLogout}>
            로그아웃
          </button>
        </div>
      </aside>
      <main className="app-main">
        <header className="page-header">
          <div>
            <p className="page-header__eyebrow">Self-hosted resume writing workspace</p>
            <h1>{title}</h1>
            {description ? <p className="page-header__description">{description}</p> : null}
          </div>
          {actions ? <div className="page-header__actions">{actions}</div> : null}
        </header>
        <div className="page-content">{loading ? <div className="page-skeleton">불러오는 중...</div> : children}</div>
      </main>
    </div>
  );
}
