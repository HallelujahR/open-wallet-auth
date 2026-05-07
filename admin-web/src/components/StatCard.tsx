import type { ReactNode } from "react";

type Props = {
  title: string;
  value: string | number;
  description?: string;
  icon?: ReactNode;
};

export function StatCard({ title, value, description, icon }: Props) {
  return (
    <div className="stat-card">
      <div>
        <p className="stat-title">{title}</p>
        <strong className="stat-value">{value}</strong>
        {description ? <span className="stat-desc">{description}</span> : null}
      </div>
      {icon ? <div className="stat-icon">{icon}</div> : null}
    </div>
  );
}
