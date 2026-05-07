import { Button, Card, Form, Input, Modal, Space, Table, Typography, message } from "antd";
import { useEffect, useState } from "react";
import { adminApi } from "../api/admin";
import { StatusBadge } from "../components/StatusBadge";
import type { Client } from "../types/api";

export function ApplicationsPage() {
  const [items, setItems] = useState<Client[]>([]);
  const [open, setOpen] = useState(false);
  const [form] = Form.useForm();

  const load = () => adminApi.listClients().then(setItems).catch((err) => message.error(err.message));

  useEffect(() => {
    load();
  }, []);

  const create = async () => {
    const values = await form.validateFields();
    await adminApi.createClient({
      client_id: values.client_id,
      name: values.name,
      jwt_audience: values.jwt_audience,
      allowed_origins: splitLines(values.allowed_origins),
      allowed_redirect_uris: splitLines(values.allowed_redirect_uris),
    });
    message.success("应用已创建");
    setOpen(false);
    form.resetFields();
    load();
  };

  return (
    <div>
      <div className="page-title-row">
        <Typography.Title level={3}>接入应用</Typography.Title>
        <Button type="primary" onClick={() => setOpen(true)}>创建应用</Button>
      </div>
      <Card>
        <Table rowKey="id" dataSource={items} columns={[
          { title: "应用标识", dataIndex: "client_id" },
          { title: "名称", dataIndex: "name" },
          { title: "令牌受众", dataIndex: "jwt_audience" },
          { title: "状态", render: (_, row) => <StatusBadge status={row.status} /> },
          { title: "允许来源", render: (_, row) => <Space wrap>{row.allowed_origins.join(", ") || "-"}</Space> },
          { title: "创建时间", dataIndex: "created_at" },
        ]} />
      </Card>
      <Modal title="创建接入应用" open={open} onOk={create} onCancel={() => setOpen(false)} okText="创建" cancelText="取消">
        <Form form={form} layout="vertical">
          <Form.Item name="client_id" label="应用标识 Client ID" rules={[{ required: true }]}>
            <Input placeholder="blockx" />
          </Form.Item>
          <Form.Item name="name" label="应用名称" rules={[{ required: true }]}>
            <Input placeholder="BlockX 链影" />
          </Form.Item>
          <Form.Item name="jwt_audience" label="令牌受众 JWT Audience" rules={[{ required: true }]}>
            <Input placeholder="blockx-api" />
          </Form.Item>
          <Form.Item name="allowed_origins" label="允许前端来源（一行一个）">
            <Input.TextArea rows={3} placeholder="http://localhost:5173" />
          </Form.Item>
          <Form.Item name="allowed_redirect_uris" label="允许回调地址（一行一个）">
            <Input.TextArea rows={3} placeholder="http://localhost:5173/auth/callback" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
}

function splitLines(value?: string) {
  return (value || "")
    .split(/\n|,/)
    .map((item) => item.trim())
    .filter(Boolean);
}
