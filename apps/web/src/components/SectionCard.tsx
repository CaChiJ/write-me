"use client";

interface SectionCardProps {
  title?: string;
  description?: string;
  actions?: React.ReactNode;
  children: React.ReactNode;
}

export function SectionCard({ title, description, actions, children }: SectionCardProps) {
  return (
    <section className="section-card">
      {(title || description || actions) && (
        <header className="section-card__header">
          <div>
            {title ? <h2>{title}</h2> : null}
            {description ? <p>{description}</p> : null}
          </div>
          {actions ? <div className="section-card__actions">{actions}</div> : null}
        </header>
      )}
      <div className="section-card__body">{children}</div>
    </section>
  );
}
