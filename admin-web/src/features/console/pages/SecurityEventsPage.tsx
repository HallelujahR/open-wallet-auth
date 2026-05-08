import { Button, Card, Form, Input, Select, Table, Typography, message } from "antd";
import { useEffect, useState } from "react";
import { adminApi } from "../../../api/admin";
import { StatusBadge } from "../../../components/StatusBadge";
import type { SecurityEvent } from "../../../types/api";

export function SecurityEventsPage() {
  const [items, setItems] = useState<SecurityEvent[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [userId, setUserId] = useState("");
  const [eventType, setEventType] = useState<string>();

  const load = () => {
    adminApi
      .listSecurityEvents({ user_id: userId || undefined, event_type: eventType, page, page_size: 20 })
      .then((result) => {
        setItems(result.items);
        setTotal(result.total);
      })
      .catch((err) => message.error(err.message));
  };

  useEffect(load, [page]);

  return (
    <div>
      <Typography.Title level={3}>安全操作</Typography.Title>
      <Card>
        <Form layout="inline" className="toolbar" onFinish={() => { setPage(1); load(); }}>
          <Input placeholder="用户ID" value={userId} onChange={(e) => setUserId(e.target.value)} />
          <Select
            allowClear
            placeholder="事件类型"
            value={eventType}
            onChange={setEventType}
            style={{ width: 220 }}
            options={[
              "change_password",
              "reset_password",
              "bind_email",
              "bind_phone",
              "bind_wallet",
              "unbind_email",
              "unbind_phone",
              "unbind_wallet",
              "unbind_oauth",
            ].map((value) => ({ value, label: securityEventLabel(value) }))}
          />
          <Button type="primary" htmlType="submit">查询</Button>
        </Form>
        <Table rowKey="id" dataSource={items} pagination={{ total, current: page, pageSize: 20, onChange: setPage }} columns={[
          { title: "时间", dataIndex: "created_at" },
          { title: "用户", dataIndex: "user_id", ellipsis: true },
          { title: "事件", render: (_, row) => securityEventLabel(row.event_type) },
          { title: "目标类型", dataIndex: "target_type" },
          { title: "目标", dataIndex: "target_id", ellipsis: true },
          { title: "状态", render: (_, row) => <StatusBadge success={row.success} /> },
          { title: "IP", dataIndex: "ip" },
        ]} />
      </Card>
    </div>
  );
}

function securityEventLabel(eventType: string) {
  const labels: Record<string, string> = {
    change_password: "修改密码",
    reset_password: "重置密码",
    bind_email: "绑定邮箱",
    bind_phone: "绑定手机",
    bind_wallet: "绑定钱包",
    unbind_email: "解绑邮箱",
    unbind_phone: "解绑手机",
    unbind_wallet: "解绑钱包",
    unbind_oauth: "解绑第三方账号",
  };
  return labels[eventType] || eventType || "-";
}
