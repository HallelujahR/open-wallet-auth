import { Button, Card, Form, Input, Popconfirm, Switch, Table, Typography, message } from "antd";
import { useEffect, useState } from "react";
import { adminApi } from "../../../api/admin";
import { StatusBadge } from "../../../components/StatusBadge";
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
          { title: "会话ID", dataIndex: "id", ellipsis: true },
          { title: "用户", dataIndex: "user_id", ellipsis: true },
          { title: "应用标识", dataIndex: "client_id" },
          { title: "IP", dataIndex: "ip" },
          { title: "状态", render: (_, row) => <StatusBadge success={row.active} /> },
          { title: "过期时间", dataIndex: "expires_at" },
          { title: "最近使用", dataIndex: "last_used_at" },
          { title: "操作", render: (_, row) => <Popconfirm title="确认吊销会话？" onConfirm={() => adminApi.revokeSession(row.id).then(load)}><Button size="small" danger disabled={!row.active}>吊销</Button></Popconfirm> },
        ]} />
      </Card>
    </div>
  );
}
