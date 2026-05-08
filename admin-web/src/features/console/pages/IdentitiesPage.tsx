import { Button, Card, Drawer, Form, Input, Popconfirm, Select, Space, Table, Tabs, Typography, message } from "antd";
import { useEffect, useState } from "react";
import { adminApi } from "../../../api/admin";
import { StatusBadge } from "../../../components/StatusBadge";
import type { IdentityDetail, IdentityUser } from "../../../types/api";

export function IdentitiesPage() {
  const [items, setItems] = useState<IdentityUser[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [keyword, setKeyword] = useState("");
  const [status, setStatus] = useState<string>();
  const [detail, setDetail] = useState<IdentityDetail | null>(null);
  const [loading, setLoading] = useState(false);

  const load = () => {
    setLoading(true);
    adminApi
      .listUsers({ keyword, status, page, page_size: 20 })
      .then((result) => {
        setItems(result.items);
        setTotal(result.total);
      })
      .catch((err) => message.error(err.message))
      .finally(() => setLoading(false));
  };

  useEffect(load, [page]);

  const openDetail = (userId: string) => {
    adminApi.getUser(userId).then(setDetail).catch((err) => message.error(err.message));
  };

  const updateStatus = (userId: string, nextStatus: string) => {
    adminApi
      .updateUserStatus(userId, nextStatus)
      .then(() => {
        message.success("状态已更新");
        load();
        if (detail?.user.id === userId) openDetail(userId);
      })
      .catch((err) => message.error(err.message));
  };

  return (
    <div>
      <Typography.Title level={3}>身份用户</Typography.Title>
      <Card>
        <Form layout="inline" className="toolbar" onFinish={() => { setPage(1); load(); }}>
          <Form.Item>
            <Input.Search placeholder="搜索用户、邮箱、手机号" value={keyword} onChange={(e) => setKeyword(e.target.value)} />
          </Form.Item>
          <Form.Item>
            <Select
              allowClear
              placeholder="状态"
              value={status}
              onChange={setStatus}
              options={[
                { value: "active", label: "启用" },
                { value: "suspended", label: "禁用" },
                { value: "deleted", label: "删除" },
              ]}
              style={{ width: 140 }}
            />
          </Form.Item>
          <Button type="primary" htmlType="submit">查询</Button>
        </Form>
        <Table
          rowKey="id"
          loading={loading}
          dataSource={items}
          pagination={{ total, current: page, pageSize: 20, onChange: setPage }}
          columns={[
            { title: "用户ID", dataIndex: "id", ellipsis: true },
            { title: "用户名", dataIndex: "username" },
            { title: "邮箱", dataIndex: "email" },
            { title: "手机号", dataIndex: "phone" },
            { title: "状态", render: (_, row) => <StatusBadge status={row.status} /> },
            { title: "最近登录", dataIndex: "last_login_at" },
            {
              title: "操作",
              render: (_, row) => (
                <Space>
                  <Button size="small" onClick={() => openDetail(row.id)}>详情</Button>
                  <Popconfirm title="确认禁用该身份？" onConfirm={() => updateStatus(row.id, "suspended")}>
                    <Button size="small" danger disabled={row.status === "suspended"}>禁用</Button>
                  </Popconfirm>
                  <Button size="small" disabled={row.status === "active"} onClick={() => updateStatus(row.id, "active")}>启用</Button>
                </Space>
              ),
            },
          ]}
        />
      </Card>
      <Drawer width={760} title="身份详情" open={!!detail} onClose={() => setDetail(null)}>
        {detail ? (
          <Tabs
            items={[
              {
                key: "profile",
                label: "基础资料",
                children: (
                  <div className="detail-list">
                    <p><b>ID：</b>{detail.user.id}</p>
                    <p><b>用户名：</b>{detail.user.username}</p>
                    <p><b>邮箱：</b>{detail.user.email || "-"}</p>
                    <p><b>手机号：</b>{detail.user.phone || "-"}</p>
                    <p><b>状态：</b><StatusBadge status={detail.user.status} /></p>
                  </div>
                ),
              },
              {
                key: "clients",
                label: `应用(${detail.clients.length})`,
                children: <Table rowKey="client_id" size="small" pagination={false} dataSource={detail.clients} columns={[
                  { title: "应用标识", dataIndex: "client_id" },
                  { title: "首次登录", dataIndex: "first_login_at" },
                  { title: "最近登录", dataIndex: "last_login_at" },
                  { title: "次数", dataIndex: "login_count" },
                ]} />,
              },
              {
                key: "wallets",
                label: `钱包(${detail.wallets.length})`,
                children: <Table rowKey="id" size="small" pagination={false} dataSource={detail.wallets} columns={[
                  { title: "链", dataIndex: "chain_type" },
                  { title: "地址", dataIndex: "address", ellipsis: true },
                  { title: "验证时间", dataIndex: "verified_at" },
                  { title: "操作", render: (_, row) => <Popconfirm title="确认解绑钱包？" onConfirm={() => adminApi.unbindWallet(detail.user.id, row.id).then(() => openDetail(detail.user.id))}><Button size="small" danger>解绑</Button></Popconfirm> },
                ]} />,
              },
              {
                key: "oauth",
                label: `第三方账号(${detail.accounts.length})`,
                children: <Table rowKey="id" size="small" pagination={false} dataSource={detail.accounts} columns={[
                  { title: "服务商", dataIndex: "provider" },
                  { title: "邮箱", dataIndex: "provider_email" },
                  { title: "用户名", dataIndex: "provider_username" },
                  { title: "操作", render: (_, row) => <Popconfirm title="确认解绑第三方账号？" onConfirm={() => adminApi.unbindOAuthAccount(detail.user.id, row.id).then(() => openDetail(detail.user.id))}><Button size="small" danger>解绑</Button></Popconfirm> },
                ]} />,
              },
              {
                key: "sessions",
                label: `会话(${detail.sessions.length})`,
                children: <Table rowKey="id" size="small" pagination={false} dataSource={detail.sessions} columns={[
                  { title: "应用标识", dataIndex: "client_id" },
                  { title: "IP", dataIndex: "ip" },
                  { title: "活跃", render: (_, row) => <StatusBadge success={row.active} /> },
                  { title: "过期时间", dataIndex: "expires_at" },
                ]} />,
              },
            ]}
          />
        ) : null}
      </Drawer>
    </div>
  );
}
