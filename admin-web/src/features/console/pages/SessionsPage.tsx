import { Button, Card, Form, Input, Popconfirm, Switch, Table, Tag, Tooltip, Typography, message } from "antd";
import { useEffect, useState } from "react";
import { adminApi } from "../../../api/admin";
import type { Session } from "../../../types/api";

export function SessionsPage() {
  const [items, setItems] = useState<Session[]>([]);
  const [userId, setUserId] = useState("");
  const [clientId, setClientId] = useState("");
  const [activeOnly, setActiveOnly] = useState(true);

  const load = () => {
    adminApi
      .listSessions({ user_id: userId || undefined, client_id: clientId || undefined, active_only: activeOnly })
      .then((result) => setItems(result.items))
      .catch((err) => message.error(err.message));
  };

  useEffect(load, []);

  return (
    <div>
      <Typography.Title level={3}>登录会话</Typography.Title>
      <Card>
        <Form layout="inline" className="toolbar" onFinish={load}>
          <Input placeholder="用户ID" value={userId} onChange={(e) => setUserId(e.target.value)} />
          <Input placeholder="应用标识" value={clientId} onChange={(e) => setClientId(e.target.value)} />
          <Switch checkedChildren="仅活跃" unCheckedChildren="全部" checked={activeOnly} onChange={setActiveOnly} />
          <Button type="primary" htmlType="submit">查询</Button>
        </Form>
        <Table rowKey="id" dataSource={items} columns={[
          {
            title: "会话",
            render: (_, row) => (
              <div className="session-cell">
                <Typography.Text copyable={{ text: row.id }} ellipsis>{shortID(row.id)}</Typography.Text>
                <div className="session-tags">
                  {row.current && <Tag color="blue">当前浏览器</Tag>}
                  <Tag color={row.active ? "success" : "default"}>{row.active ? "有效" : "已吊销"}</Tag>
                </div>
              </div>
            ),
          },
          { title: "用户", dataIndex: "user_id", ellipsis: true },
          { title: "应用", dataIndex: "client_id" },
          { title: "IP", dataIndex: "ip" },
          {
            title: "设备",
            render: (_, row) => (
              <Tooltip title={row.user_agent || "无 User-Agent"}>
                <span>{deviceLabel(row.user_agent)}</span>
              </Tooltip>
            ),
          },
          { title: "创建时间", dataIndex: "created_at" },
          { title: "过期时间", dataIndex: "expires_at" },
          { title: "最近使用", render: (_, row) => row.last_used_at || "-" },
          {
            title: "操作",
            render: (_, row) => (
              <Popconfirm
                title="确认吊销这个登录会话？"
                description={`应用：${row.client_id}，设备：${deviceLabel(row.user_agent)}。吊销后该设备需要重新登录。`}
                onConfirm={() => adminApi.revokeSession(row.id).then(load)}
              >
                <Button size="small" danger disabled={!row.active}>吊销</Button>
              </Popconfirm>
            ),
          },
        ]} />
      </Card>
    </div>
  );
}

// shortID keeps the table readable while copyable text preserves the full id.
// shortID 缩短表格中的会话 ID，完整值仍可复制。
function shortID(id: string) {
  if (!id || id.length <= 16) return id;
  return `${id.slice(0, 8)}...${id.slice(-6)}`;
}

// deviceLabel gives admins a practical device/browser hint from User-Agent.
// deviceLabel 从 User-Agent 提取便于管理员判断的设备和浏览器摘要。
function deviceLabel(userAgent?: string) {
  if (!userAgent) return "未知设备";
  const os = /Windows/i.test(userAgent)
    ? "Windows"
    : /Mac OS X|Macintosh/i.test(userAgent)
      ? "macOS"
      : /iPhone|iPad/i.test(userAgent)
        ? "iOS"
        : /Android/i.test(userAgent)
          ? "Android"
          : /Linux/i.test(userAgent)
            ? "Linux"
            : "未知系统";
  const browser = /Edg\//i.test(userAgent)
    ? "Edge"
    : /Chrome\//i.test(userAgent)
      ? "Chrome"
      : /Safari\//i.test(userAgent)
        ? "Safari"
        : /Firefox\//i.test(userAgent)
          ? "Firefox"
          : "浏览器";
  return `${os} / ${browser}`;
}
