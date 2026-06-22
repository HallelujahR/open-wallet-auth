import {
  Alert,
  Button,
  Card,
  Drawer,
  Form,
  Input,
  Modal,
  Popconfirm,
  Select,
  Space,
  Switch,
  Table,
  Tag,
  Typography,
  message,
} from "antd";
import { useEffect, useState } from "react";
import { adminApi } from "../../../api/admin";
import { StatusBadge } from "../../../components/StatusBadge";
import type { Client, ClientMember, ClientMemberInput, IdentityUser } from "../../../types/api";

// ApplicationsPage manages OAuth/OIDC client applications and per-client allow-list access.
// ApplicationsPage 管理接入应用及应用级登录白名单，白名单开启后只有成员可以进入对应业务系统。
export function ApplicationsPage() {
  const [items, setItems] = useState<Client[]>([]);
  const [open, setOpen] = useState(false);
  const [selectedClient, setSelectedClient] = useState<Client | null>(null);
  const [members, setMembers] = useState<ClientMember[]>([]);
  const [memberLoading, setMemberLoading] = useState(false);
  const [userOptions, setUserOptions] = useState<IdentityUser[]>([]);
  const [editingMember, setEditingMember] = useState<ClientMember | null>(null);
  const [form] = Form.useForm();
  const [memberForm] = Form.useForm();
  const [editForm] = Form.useForm();

  const load = () => adminApi.listClients().then(setItems).catch((err) => message.error(err.message));

  useEffect(() => {
    load();
  }, []);

  const loadMembers = (client: Client) => {
    setMemberLoading(true);
    adminApi
      .listClientMembers(client.client_id)
      .then(setMembers)
      .catch((err) => message.error(err.message))
      .finally(() => setMemberLoading(false));
  };

  const create = async () => {
    try {
      const values = await form.validateFields();
      await adminApi.createClient({
        client_id: values.client_id,
        name: values.name,
        jwt_audience: values.jwt_audience,
        allowed_origins: splitLines(values.allowed_origins),
        allowed_redirect_uris: splitLines(values.allowed_redirect_uris),
        whitelist_enabled: !!values.whitelist_enabled,
      });
      message.success("应用已创建");
      setOpen(false);
      form.resetFields();
      load();
    } catch (err: any) {
      if (err?.errorFields) return;
      message.error(err.message || "创建失败");
    }
  };

  const toggleWhitelist = async (client: Client, enabled: boolean) => {
    try {
      await adminApi.updateClientAccessPolicy(client.client_id, enabled);
      message.success(enabled ? "登录白名单已启用" : "登录白名单已关闭");
      load();
      if (selectedClient?.client_id === client.client_id) {
        setSelectedClient({ ...client, whitelist_enabled: enabled });
      }
    } catch (err: any) {
      message.error(err.message || "更新失败");
    }
  };

  const openMembers = (client: Client) => {
    setSelectedClient(client);
    memberForm.resetFields();
    searchUsers("");
    loadMembers(client);
  };

  const searchUsers = (keyword: string) => {
    adminApi
      .listUsers({ keyword, status: "active", page: 1, page_size: 20 })
      .then((result) => setUserOptions(result.items))
      .catch((err) => message.error(err.message));
  };

  const addMember = async () => {
    if (!selectedClient) return;
    try {
      const values = await memberForm.validateFields();
      await adminApi.addClientMember(selectedClient.client_id, memberInput(values));
      message.success("成员已保存");
      memberForm.resetFields(["user_id", "remark"]);
      loadMembers(selectedClient);
    } catch (err: any) {
      if (err?.errorFields) return;
      message.error(err.message || "保存失败");
    }
  };

  const startEditMember = (member: ClientMember) => {
    setEditingMember(member);
    editForm.setFieldsValue({
      role: member.role || "member",
      permissions: (member.permissions || []).join("\n"),
      status: member.status || "active",
      remark: member.remark,
    });
  };

  const updateMember = async () => {
    if (!selectedClient || !editingMember) return;
    try {
      const values = await editForm.validateFields();
      await adminApi.updateClientMember(selectedClient.client_id, editingMember.id, memberInput(values));
      message.success("成员已更新");
      setEditingMember(null);
      loadMembers(selectedClient);
    } catch (err: any) {
      if (err?.errorFields) return;
      message.error(err.message || "更新失败");
    }
  };

  const deleteMember = async (member: ClientMember) => {
    if (!selectedClient) return;
    try {
      await adminApi.deleteClientMember(selectedClient.client_id, member.id);
      message.success("成员已移除");
      loadMembers(selectedClient);
    } catch (err: any) {
      message.error(err.message || "移除失败");
    }
  };

  return (
    <div>
      <div className="page-title-row">
        <Typography.Title level={3}>接入应用</Typography.Title>
        <Button type="primary" onClick={() => setOpen(true)}>创建应用</Button>
      </div>
      <Card>
        <Table
          rowKey="id"
          dataSource={items}
          columns={[
            { title: "应用标识", dataIndex: "client_id" },
            { title: "名称", dataIndex: "name" },
            { title: "令牌受众", dataIndex: "jwt_audience" },
            { title: "状态", render: (_, row) => <StatusBadge status={row.status} /> },
            {
              title: "登录白名单",
              render: (_, row) => (
                <Switch
                  checked={row.whitelist_enabled}
                  checkedChildren="开启"
                  unCheckedChildren="关闭"
                  onChange={(checked) => toggleWhitelist(row, checked)}
                />
              ),
            },
            { title: "允许来源", render: (_, row) => <Typography.Text>{row.allowed_origins.join(", ") || "-"}</Typography.Text> },
            { title: "创建时间", dataIndex: "created_at" },
            { title: "操作", render: (_, row) => <Button size="small" onClick={() => openMembers(row)}>成员</Button> },
          ]}
        />
      </Card>

      <Modal title="创建接入应用" open={open} onOk={create} onCancel={() => setOpen(false)} okText="创建" cancelText="取消">
        <Form form={form} layout="vertical">
          <Form.Item name="client_id" label="应用标识 Client ID" rules={[{ required: true }]}>
            <Input placeholder="blockx" />
          </Form.Item>
          <Form.Item name="name" label="应用名称" rules={[{ required: true }]}>
            <Input placeholder="Example App" />
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
          <Form.Item name="whitelist_enabled" label="登录白名单" valuePropName="checked">
            <Switch checkedChildren="开启" unCheckedChildren="关闭" />
          </Form.Item>
        </Form>
      </Modal>

      <Drawer
        width={920}
        title={selectedClient ? `${selectedClient.name} 成员白名单` : "成员白名单"}
        open={!!selectedClient}
        onClose={() => setSelectedClient(null)}
      >
        {selectedClient ? (
          <div className="application-member-drawer">
            <Alert
              type={selectedClient.whitelist_enabled ? "warning" : "info"}
              showIcon
              message={selectedClient.whitelist_enabled ? "该应用已启用登录白名单" : "该应用未启用登录白名单"}
              description={selectedClient.whitelist_enabled ? "只有下方状态为启用的成员可以登录该业务系统。" : "关闭时不限制统一身份用户登录；成员列表可提前维护。"}
            />

            <Card size="small" title="添加成员" className="application-member-form">
              <Form form={memberForm} layout="vertical">
                <div className="application-member-grid">
                  <Form.Item name="user_id" label="用户" rules={[{ required: true, message: "请选择用户" }]}>
                    <Select
                      showSearch
                      filterOption={false}
                      placeholder="搜索用户名、邮箱、手机号"
                      onSearch={searchUsers}
                      options={userOptions.map((user) => ({
                        value: user.id,
                        label: userLabel(user),
                      }))}
                    />
                  </Form.Item>
                  <Form.Item name="role" label="角色" initialValue="member">
                    <Input placeholder="member" />
                  </Form.Item>
                  <Form.Item name="status" label="状态" initialValue="active">
                    <Select options={memberStatusOptions} />
                  </Form.Item>
                  <Form.Item name="permissions" label="权限标识（一行一个）">
                    <Input.TextArea rows={2} placeholder="case:read" />
                  </Form.Item>
                  <Form.Item name="remark" label="备注">
                    <Input placeholder="授权原因或适用范围" />
                  </Form.Item>
                  <Form.Item label=" ">
                    <Button type="primary" onClick={addMember}>保存成员</Button>
                  </Form.Item>
                </div>
              </Form>
            </Card>

            <Table
              rowKey="id"
              size="small"
              loading={memberLoading}
              dataSource={members}
              columns={[
                {
                  title: "用户",
                  render: (_, row) => (
                    <div>
                      <Typography.Text strong>{row.username || row.email || row.phone || row.user_id}</Typography.Text>
                      <div className="muted-text">{row.email || row.phone || row.user_id}</div>
                    </div>
                  ),
                },
                { title: "角色", dataIndex: "role" },
                {
                  title: "权限",
                  render: (_, row) => (
                    <Space wrap>
                      {(row.permissions || []).length ? row.permissions.map((item) => <Tag key={item}>{item}</Tag>) : "-"}
                    </Space>
                  ),
                },
                { title: "状态", render: (_, row) => <StatusBadge status={row.status} /> },
                { title: "备注", dataIndex: "remark" },
                { title: "更新时间", dataIndex: "updated_at" },
                {
                  title: "操作",
                  render: (_, row) => (
                    <Space>
                      <Button size="small" onClick={() => startEditMember(row)}>编辑</Button>
                      <Popconfirm title="确认移除该成员？" onConfirm={() => deleteMember(row)}>
                        <Button size="small" danger>移除</Button>
                      </Popconfirm>
                    </Space>
                  ),
                },
              ]}
            />
          </div>
        ) : null}
      </Drawer>

      <Modal title="编辑成员权限" open={!!editingMember} onOk={updateMember} onCancel={() => setEditingMember(null)} okText="保存" cancelText="取消">
        <Form form={editForm} layout="vertical">
          <Form.Item name="role" label="角色" rules={[{ required: true }]}>
            <Input placeholder="member" />
          </Form.Item>
          <Form.Item name="status" label="状态" rules={[{ required: true }]}>
            <Select options={memberStatusOptions} />
          </Form.Item>
          <Form.Item name="permissions" label="权限标识（一行一个）">
            <Input.TextArea rows={3} placeholder="case:read" />
          </Form.Item>
          <Form.Item name="remark" label="备注">
            <Input placeholder="授权原因或适用范围" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
}

const memberStatusOptions = [
  { value: "active", label: "启用" },
  { value: "disabled", label: "停用" },
];

function memberInput(values: Record<string, unknown>): ClientMemberInput {
  return {
    user_id: typeof values.user_id === "string" ? values.user_id : undefined,
    role: String(values.role || "member"),
    status: String(values.status || "active"),
    permissions: splitLines(typeof values.permissions === "string" ? values.permissions : ""),
    remark: typeof values.remark === "string" ? values.remark : undefined,
  };
}

function userLabel(user: IdentityUser) {
  const contact = user.email || user.phone || user.id;
  return `${user.username || "未命名用户"} · ${contact}`;
}

function splitLines(value?: string) {
  return (value || "")
    .split(/\n|,/)
    .map((item) => item.trim())
    .filter(Boolean);
}
