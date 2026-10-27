import { Typography } from "antd";
import type { ReactNode } from "react";

export function PageHeader({
  title,
  description,
  actions,
}: {
  title: string;
  description?: string;
  actions?: ReactNode;
}) {
  return (
    <div className="page-heading">
      <div>
        <Typography.Title level={2}>{title}</Typography.Title>
        {description && (
          <Typography.Text type="secondary">{description}</Typography.Text>
        )}
      </div>
      {actions && <div className="page-heading-actions">{actions}</div>}
    </div>
  );
}
