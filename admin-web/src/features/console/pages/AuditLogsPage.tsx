import { formatTime } from "../../../utils/format";
import { PageHeader } from "../../../components/PageHeader";
import { Button, Card, Form, Input, Table, App } from "antd";
import { useEffect, useRef, useState } from "react";
import { adminApi } from "../../../api/admin";
import { StatusBadge } from "../../../components/StatusBadge";
import type { LoginLog } from "../../../types/api";

export function AuditLogsPage() {
  const { message } = App.useApp();
  const [loading, setLoading] = useState(false);
  const [items, setItems] = useState<LoginLog[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [userId, setUserId] = useState("");
  const [clientId, setClientId] = useState("");

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
      .listLoginLogs({
        user_id: userId || undefined,
        client_id: clientId || undefined,
        page,
        page_size: 20,
      })
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

  return (
    <div>
      <PageHeader
        title="登录审计"
        description="追踪每一次登录，了解认证结果与来源。"
      />
      <Card>
        <Form
          layout="inline"
          className="toolbar"
          onFinish={() => {
            if (page === 1) load();
            else setPage(1);
          }}
        >
          <Input
            placeholder="用户ID"
            value={userId}
            onChange={(e) => setUserId(e.target.value)}
          />
          <Input
            placeholder="应用标识"
            value={clientId}
            onChange={(e) => setClientId(e.target.value)}
          />
          <Button type="primary" htmlType="submit">
            查询
          </Button>
        </Form>
        <Table
          loading={loading}
          scroll={{ x: "max-content" }}
          rowKey="id"
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
            { title: "时间", dataIndex: "created_at", render: formatTime },
            { title: "用户", dataIndex: "user_id", ellipsis: true },
            { title: "应用标识", dataIndex: "client_id" },
            {
              title: "方式",
              render: (_, row) => loginMethodLabel(row.login_method),
            },
            {
              title: "状态",
              render: (_, row) => <StatusBadge success={row.success} />,
            },
            { title: "失败原因", dataIndex: "failure_code" },
            { title: "IP", dataIndex: "ip" },
          ]}
        />
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
