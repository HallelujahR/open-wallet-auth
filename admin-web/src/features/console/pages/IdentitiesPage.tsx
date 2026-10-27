import { formatTime } from "../../../utils/format";
import { PageHeader } from "../../../components/PageHeader";
import {
  Button,
  Card,
  Drawer,
  Dropdown,
  Form,
  Input,
  Modal,
  Popconfirm,
  Select,
  Space,
  Table,
  Tabs,
  Tag,
  Tooltip,
  Typography,
  Avatar,
  App,
} from "antd";
import {
  EllipsisOutlined,
  GithubOutlined,
  GoogleOutlined,
  MailOutlined,
  PhoneOutlined,
  UnlockOutlined,
  UserOutlined,
  WalletOutlined,
} from "@ant-design/icons";
import { useEffect, useMemo, useRef, useState, type ReactNode } from "react";
import { adminApi } from "../../../api/admin";
import { StatusBadge } from "../../../components/StatusBadge";
import type {
  IdentityDetail,
  IdentityUser,
  OAuthAccount,
  WalletBinding,
} from "../../../types/api";

// IdentitiesPage presents identity users and their bound login factors.
// IdentitiesPage 展示身份用户及其绑定的登录方式，方便管理端快速判断账号来源。
export function IdentitiesPage() {
  const { message, modal } = App.useApp();
  const [items, setItems] = useState<IdentityUser[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [keyword, setKeyword] = useState("");
  const [status, setStatus] = useState<string>();
  const [detail, setDetail] = useState<IdentityDetail | null>(null);
  const [loading, setLoading] = useState(false);
  const [passwordTarget, setPasswordTarget] = useState<IdentityUser | null>(
    null,
  );
  const [passwordSaving, setPasswordSaving] = useState(false);
  const [clearingLimitUserID, setClearingLimitUserID] = useState("");
  const [passwordForm] = Form.useForm<{
    password: string;
    confirm_password: string;
  }>();

  const requestRef = useRef(0);
  useEffect(
    () => () => {
      requestRef.current++;
    },
    [],
  );
  const load = () => {
    const request = ++requestRef.current;
    setLoading(true);
    adminApi
      .listUsers({ keyword, status, page, page_size: 20 })
      .then((result) => {
        if (request !== requestRef.current) return;
        setItems(result.items);
        setTotal(result.total);
      })
      .catch((err) => {
        if (request === requestRef.current) message.error(err.message);
      })
      .finally(() => {
        if (request === requestRef.current) setLoading(false);
      });
  };

  useEffect(load, [page]);

  const openDetail = (userId: string) => {
    adminApi
      .getUser(userId)
      .then(setDetail)
      .catch((err) => message.error(err.message));
  };

  const updateStatus = (userId: string, nextStatus: string) => {
    return adminApi
      .updateUserStatus(userId, nextStatus)
      .then(() => {
        message.success("状态已更新");
        load();
        if (detail?.user.id === userId) openDetail(userId);
      })
      .catch((err) => message.error(err.message));
  };

  const openPasswordModal = (user: IdentityUser) => {
    setPasswordTarget(user);
    passwordForm.resetFields();
  };

  const submitPassword = async () => {
    if (!passwordTarget) return;
    let values;
    try {
      values = await passwordForm.validateFields();
    } catch {
      return;
    }
    setPasswordSaving(true);
    adminApi
      .setUserPassword(passwordTarget.id, values.password)
      .then(() => {
        message.success("密码已更新，用户现有会话已吊销");
        setPasswordTarget(null);
        passwordForm.resetFields();
        load();
        if (detail?.user.id === passwordTarget.id)
          openDetail(passwordTarget.id);
      })
      .catch((err) => message.error(err.message))
      .finally(() => setPasswordSaving(false));
  };

  const resetLoginRateLimit = (user: IdentityUser) => {
    if (!user.email) {
      message.warning("该用户没有邮箱登录方式");
      return;
    }
    setClearingLimitUserID(user.id);
    return adminApi
      .resetLoginRateLimit(user.id)
      .then((result) => {
        if (result.cleared > 0) {
          message.success(`已清除 ${result.cleared} 个应用下的登录限制`);
          return;
        }
        message.info("该账号当前没有生效中的登录限制");
      })
      .catch((err) => message.error(err.message))
      .finally(() => setClearingLimitUserID(""));
  };

  const detailMethods = useMemo(
    () =>
      detail
        ? identityMethods(detail.user, detail.wallets, detail.accounts)
        : [],
    [detail],
  );

  return (
    <div>
      <PageHeader
        title="身份用户"
        description="查看用户身份、登录方式与账号状态。"
      />
      <Card className="identity-card">
        <Form
          layout="inline"
          className="toolbar"
          onFinish={() => {
            if (page === 1) load();
            else setPage(1);
          }}
        >
          <Form.Item>
            <Input.Search
              placeholder="搜索用户名、邮箱、手机号"
              value={keyword}
              onChange={(e) => setKeyword(e.target.value)}
              onSearch={() => {
                if (page === 1) load();
                else setPage(1);
              }}
            />
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
          <Button type="primary" htmlType="submit">
            查询
          </Button>
        </Form>
        <Table
          scroll={{ x: "max-content" }}
          rowKey="id"
          loading={loading}
          dataSource={items}
          pagination={{
            total,
            current: page,
            pageSize: 20,
            showSizeChanger: false,
            showTotal: (total) => `共 ${total} 条`,
            onChange: setPage,
          }}
          columns={[
            {
              title: "用户",
              render: (_, row) => (
                <div className="identity-user-cell">
                  <Avatar src={row.avatar} icon={<UserOutlined />} />
                  <div>
                    <Typography.Text strong>
                      {row.username || "未命名用户"}
                    </Typography.Text>
                    <div className="identity-meta">
                      <Typography.Text copyable={{ text: row.id }} ellipsis>
                        {shortID(row.id)}
                      </Typography.Text>
                    </div>
                  </div>
                </div>
              ),
            },
            {
              title: "联系方式",
              render: (_, row) => (
                <div className="identity-contact-cell">
                  <span>
                    <MailOutlined /> {row.email || "-"}
                  </span>
                  <span>
                    <PhoneOutlined /> {row.phone || "-"}
                  </span>
                </div>
              ),
            },
            {
              title: "登录方式",
              render: (_, row) => (
                <MethodTags
                  methods={
                    row.login_methods ||
                    identityMethods(
                      row,
                      row.wallets || [],
                      row.oauth_accounts || [],
                    )
                  }
                />
              ),
            },
            {
              title: "钱包地址",
              render: (_, row) => <WalletSummary wallets={row.wallets || []} />,
            },
            {
              title: "状态",
              render: (_, row) => <StatusBadge status={row.status} />,
            },
            {
              title: "最近登录",
              render: (_, row) => formatTime(row.last_login_at),
            },
            {
              title: "操作",
              fixed: "right",
              width: 112,
              render: (_, row) => (
                <Space size={4}>
                  <Button
                    type="link"
                    size="small"
                    onClick={() => openDetail(row.id)}
                  >
                    详情
                  </Button>
                  <Dropdown
                    trigger={["click"]}
                    menu={{
                      items: [
                        {
                          key: "password",
                          label: "修改密码",
                          onClick: () => openPasswordModal(row),
                        },
                        {
                          key: "limit",
                          label: "解除登录限流",
                          disabled:
                            !row.email || clearingLimitUserID === row.id,
                          onClick: () =>
                            modal.confirm({
                              title: "清除该账号的登录限制？",
                              content:
                                "将清除该邮箱在所有接入应用下的登录计数。",
                              onOk: () => resetLoginRateLimit(row),
                            }),
                        },
                        { type: "divider" },
                        {
                          key: "status",
                          label:
                            row.status === "active" ? "禁用账号" : "启用账号",
                          danger: row.status === "active",
                          onClick: () =>
                            modal.confirm({
                              title:
                                row.status === "active"
                                  ? "确认禁用该身份？"
                                  : "确认启用该身份？",
                              onOk: () =>
                                updateStatus(
                                  row.id,
                                  row.status === "active"
                                    ? "suspended"
                                    : "active",
                                ),
                            }),
                        },
                      ],
                    }}
                  >
                    <Button
                      type="text"
                      size="small"
                      aria-label={`管理 ${row.username || row.id}`}
                      icon={<EllipsisOutlined />}
                    />
                  </Dropdown>
                </Space>
              ),
            },
          ]}
        />
      </Card>
      <Drawer
        width={880}
        title="身份详情"
        open={!!detail}
        onClose={() => setDetail(null)}
      >
        {detail ? (
          <div className="identity-detail">
            <div className="identity-detail-hero">
              <Avatar
                size={56}
                src={detail.user.avatar}
                icon={<UserOutlined />}
              />
              <div>
                <Typography.Title level={4}>
                  {detail.user.username || "未命名用户"}
                </Typography.Title>
                <div className="identity-detail-id">
                  <Typography.Text copyable={{ text: detail.user.id }}>
                    {detail.user.id}
                  </Typography.Text>
                </div>
                <div className="identity-methods">
                  <MethodTags methods={detailMethods} />
                </div>
              </div>
              <Space>
                <Button
                  size="small"
                  onClick={() => openPasswordModal(detail.user)}
                >
                  修改密码
                </Button>
                <Popconfirm
                  title="清除该账号的登录限制？"
                  description="将清除该邮箱在所有接入应用下的登录计数。"
                  onConfirm={() => resetLoginRateLimit(detail.user)}
                >
                  <Button
                    size="small"
                    icon={<UnlockOutlined />}
                    loading={clearingLimitUserID === detail.user.id}
                    disabled={!detail.user.email}
                  >
                    解除限流
                  </Button>
                </Popconfirm>
                <StatusBadge status={detail.user.status} />
              </Space>
            </div>

            <Tabs
              items={[
                {
                  key: "profile",
                  label: "基础资料",
                  children: (
                    <div className="identity-info-grid">
                      <InfoTile
                        label="邮箱"
                        value={detail.user.email || "-"}
                        icon={<MailOutlined />}
                      />
                      <InfoTile
                        label="手机号"
                        value={detail.user.phone || "-"}
                        icon={<PhoneOutlined />}
                      />
                      <InfoTile
                        label="最近登录"
                        value={formatTime(detail.user.last_login_at)}
                      />
                      <InfoTile
                        label="创建时间"
                        value={formatTime(detail.user.created_at)}
                      />
                    </div>
                  ),
                },
                {
                  key: "clients",
                  label: `应用(${detail.clients.length})`,
                  children: (
                    <Table
                      scroll={{ x: "max-content" }}
                      rowKey="client_id"
                      size="small"
                      pagination={false}
                      dataSource={detail.clients}
                      columns={[
                        { title: "应用标识", dataIndex: "client_id" },
                        {
                          title: "首次登录",
                          dataIndex: "first_login_at",
                          render: formatTime,
                        },
                        {
                          title: "最近登录",
                          dataIndex: "last_login_at",
                          render: formatTime,
                        },
                        { title: "次数", dataIndex: "login_count" },
                        {
                          title: "状态",
                          render: (_, row) => (
                            <StatusBadge status={row.status} />
                          ),
                        },
                      ]}
                    />
                  ),
                },
                {
                  key: "wallets",
                  label: `钱包(${detail.wallets.length})`,
                  children: (
                    <Table
                      scroll={{ x: "max-content" }}
                      rowKey="id"
                      size="small"
                      pagination={false}
                      dataSource={detail.wallets}
                      columns={[
                        { title: "链", dataIndex: "chain_type" },
                        {
                          title: "地址",
                          render: (_, row) => (
                            <Typography.Text copyable={{ text: row.address }}>
                              {shortAddress(row.address)}
                            </Typography.Text>
                          ),
                        },
                        { title: "验证时间", dataIndex: "verified_at" },
                        {
                          title: "操作",
                          render: (_, row) => (
                            <Popconfirm
                              title="确认解绑钱包？"
                              onConfirm={() =>
                                adminApi
                                  .unbindWallet(detail.user.id, row.id)
                                  .then(() => openDetail(detail.user.id))
                              }
                            >
                              <Button size="small" danger>
                                解绑
                              </Button>
                            </Popconfirm>
                          ),
                        },
                      ]}
                    />
                  ),
                },
                {
                  key: "oauth",
                  label: `第三方账号(${detail.accounts.length})`,
                  children: (
                    <Table
                      scroll={{ x: "max-content" }}
                      rowKey="id"
                      size="small"
                      pagination={false}
                      dataSource={detail.accounts}
                      columns={[
                        {
                          title: "服务商",
                          render: (_, row) => (
                            <MethodTag method={row.provider} />
                          ),
                        },
                        { title: "邮箱", dataIndex: "provider_email" },
                        { title: "用户名", dataIndex: "provider_username" },
                        {
                          title: "操作",
                          render: (_, row) => (
                            <Popconfirm
                              title="确认解绑第三方账号？"
                              onConfirm={() =>
                                adminApi
                                  .unbindOAuthAccount(detail.user.id, row.id)
                                  .then(() => openDetail(detail.user.id))
                              }
                            >
                              <Button size="small" danger>
                                解绑
                              </Button>
                            </Popconfirm>
                          ),
                        },
                      ]}
                    />
                  ),
                },
                {
                  key: "sessions",
                  label: `会话(${detail.sessions.length})`,
                  children: (
                    <Table
                      scroll={{ x: "max-content" }}
                      rowKey="id"
                      size="small"
                      pagination={false}
                      dataSource={detail.sessions}
                      columns={[
                        { title: "应用标识", dataIndex: "client_id" },
                        { title: "IP", dataIndex: "ip" },
                        {
                          title: "状态",
                          render: (_, row) => (
                            <Tag color={row.active ? "success" : "default"}>
                              {row.active
                                ? "有效"
                                : row.revoked_at
                                  ? "已吊销"
                                  : "已过期"}
                            </Tag>
                          ),
                        },
                        {
                          title: "过期时间",
                          dataIndex: "expires_at",
                          render: formatTime,
                        },
                      ]}
                    />
                  ),
                },
              ]}
            />
          </div>
        ) : null}
      </Drawer>
      <Modal
        title="修改身份用户密码"
        open={!!passwordTarget}
        okText="确认修改"
        cancelText="取消"
        confirmLoading={passwordSaving}
        onOk={submitPassword}
        onCancel={() => setPasswordTarget(null)}
        destroyOnClose
      >
        <Typography.Paragraph type="secondary">
          将为{" "}
          {passwordTarget?.username ||
            passwordTarget?.email ||
            passwordTarget?.id}{" "}
          设置新的统一认证密码，并吊销该用户现有登录会话。
        </Typography.Paragraph>
        <Form form={passwordForm} layout="vertical">
          <Form.Item
            label="新密码"
            name="password"
            rules={[
              { required: true, message: "请输入新密码" },
              { min: 8, message: "密码至少 8 位" },
            ]}
          >
            <Input.Password
              autoComplete="new-password"
              placeholder="请输入至少 8 位的新密码"
            />
          </Form.Item>
          <Form.Item
            label="确认新密码"
            name="confirm_password"
            dependencies={["password"]}
            rules={[
              { required: true, message: "请再次输入新密码" },
              ({ getFieldValue }) => ({
                validator(_, value) {
                  if (!value || getFieldValue("password") === value)
                    return Promise.resolve();
                  return Promise.reject(new Error("两次输入的密码不一致"));
                },
              }),
            ]}
          >
            <Input.Password
              autoComplete="new-password"
              placeholder="请再次输入新密码"
            />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
}

function InfoTile({
  label,
  value,
  icon,
}: {
  label: string;
  value: string;
  icon?: ReactNode;
}) {
  return (
    <div className="identity-info-tile">
      <span>
        {icon}
        {label}
      </span>
      <strong>{value}</strong>
    </div>
  );
}

function MethodTags({ methods }: { methods: string[] }) {
  if (methods.length === 0)
    return <Typography.Text type="secondary">未绑定</Typography.Text>;
  return (
    <div className="identity-method-tags">
      {methods.map((method) => (
        <MethodTag key={method} method={method} />
      ))}
    </div>
  );
}

function MethodTag({ method }: { method: string }) {
  const normalized = method.toLowerCase();
  const icon =
    normalized === "email" ? (
      <MailOutlined />
    ) : normalized === "phone" ? (
      <PhoneOutlined />
    ) : normalized === "wallet" ? (
      <WalletOutlined />
    ) : normalized === "github" ? (
      <GithubOutlined />
    ) : normalized === "google" ? (
      <GoogleOutlined />
    ) : undefined;
  return (
    <Tag color={methodColor(normalized)} icon={icon}>
      {methodLabel(normalized)}
    </Tag>
  );
}

function WalletSummary({ wallets }: { wallets: WalletBinding[] }) {
  if (wallets.length === 0)
    return <Typography.Text type="secondary">未绑定</Typography.Text>;
  const primary = wallets[0];
  return (
    <Tooltip
      title={wallets
        .map((wallet) => `${wallet.chain_type}: ${wallet.address}`)
        .join("\n")}
    >
      <div className="wallet-summary">
        <WalletOutlined />
        <Typography.Text copyable={{ text: primary.address }}>
          {shortAddress(primary.address)}
        </Typography.Text>
        {wallets.length > 1 && <Tag>+{wallets.length - 1}</Tag>}
      </div>
    </Tooltip>
  );
}

function identityMethods(
  user: IdentityUser,
  wallets: WalletBinding[],
  accounts: OAuthAccount[],
) {
  const methods: string[] = [];
  if (user.email) methods.push("email");
  if (user.phone) methods.push("phone");
  if (wallets.length > 0) methods.push("wallet");
  accounts.forEach((account) => methods.push(account.provider));
  return Array.from(new Set(methods));
}

function methodLabel(method: string) {
  const labels: Record<string, string> = {
    email: "邮箱",
    phone: "手机",
    wallet: "钱包",
    github: "GitHub",
    google: "Google",
  };
  return labels[method] || method;
}

function methodColor(method: string) {
  const colors: Record<string, string> = {
    email: "blue",
    phone: "cyan",
    wallet: "purple",
    github: "default",
    google: "volcano",
  };
  return colors[method] || "geekblue";
}

function shortAddress(address: string) {
  if (!address || address.length <= 18) return address;
  return `${address.slice(0, 8)}...${address.slice(-6)}`;
}

function shortID(id: string) {
  if (!id || id.length <= 18) return id;
  return `${id.slice(0, 8)}...${id.slice(-6)}`;
}
