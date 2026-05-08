import { Button, Card, Form, Input, Table, Typography, message } from "antd";
import { useEffect, useState } from "react";
import { adminApi } from "../../../api/admin";
import { StatusBadge } from "../../../components/StatusBadge";
import type { LoginLog } from "../../../types/api";

export function AuditLogsPage() {
  const [items, setItems] = useState<LoginLog[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [userId, setUserId] = useState("");
  const [clientId, setClientId] = useState("");

  const load = () => {
    adminApi
      .listLoginLogs({ user_id: userId || undefined, client_id: clientId || undefined, page, page_size: 20 })
      .then((result) => {
        setItems(result.items);
        setTotal(result.total);
      })
      .catch((err) => message.error(err.message));
  };

  useEffect(load, [page]);

  return (
    <div>
      <Typography.Title level={3}>登录审计</Typography.Title>
      <Card>
        <Form layout="inline" className="toolbar" onFinish={() => { setPage(1); load(); }}>
          <Input placeholder="用户ID" value={userId} onChange={(e) => setUserId(e.target.value)} />
          <Input placeholder="应用标识" value={clientId} onChange={(e) => setClientId(e.target.value)} />
          <Button type="primary" htmlType="submit">查询</Button>
        </Form>
        <Table rowKey="id" dataSource={items} pagination={{ total, current: page, pageSize: 20, onChange: setPage }} columns={[
          { title: "时间", dataIndex: "created_at" },
          { title: "用户", dataIndex: "user_id", ellipsis: true },
          { title: "应用标识", dataIndex: "client_id" },
          { title: "方式", render: (_, row) => loginMethodLabel(row.login_method) },
          { title: "状态", render: (_, row) => <StatusBadge success={row.success} /> },
          { title: "失败原因", dataIndex: "failure_code" },
          { title: "IP", dataIndex: "ip" },
        ]} />
      </Card>
    </div>
  );
}

function loginMethodLabel(method: string) {
  const labels: Record<string, string> = {
    password: "邮箱密码",
    wallet: "钱包",
    oauth: "第三方账号",
    phone: "手机验证码",
    refresh: "刷新令牌",
  };
  return labels[method] || method || "-";
}
