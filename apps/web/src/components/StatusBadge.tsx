"use client";

interface StatusBadgeProps {
  tone?: "neutral" | "warning" | "success" | "danger" | "info";
  children: React.ReactNode;
}

export function StatusBadge({ tone = "neutral", children }: StatusBadgeProps) {
  return <span className={`status-badge status-badge--${tone}`}>{children}</span>;
}
