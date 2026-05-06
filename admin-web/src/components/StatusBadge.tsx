import { Tag } from "antd";

type Props = {
  status?: string;
  success?: boolean;
};

export function StatusBadge({ status, success }: Props) {
  if (typeof success === "boolean") {
    return <Tag color={success ? "success" : "error"}>{success ? "成功" : "失败"}</Tag>;
  }
  const normalized = status || "unknown";
  const color = normalized === "active" ? "success" : normalized === "suspended" ? "warning" : "default";
  const labelMap: Record<string, string> = {
    active: "启用",
    suspended: "禁用",
    deleted: "删除",
    enabled: "启用",
    disabled: "停用",
  };
  return <Tag color={color}>{labelMap[normalized] || normalized}</Tag>;
}
