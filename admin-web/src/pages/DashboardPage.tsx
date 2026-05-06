import { ApiOutlined, AppstoreOutlined, SafetyOutlined, TeamOutlined } from "@ant-design/icons";
import { Alert, Card, Col, Row, Table, Typography } from "antd";
import { useEffect, useMemo, useState } from "react";
import * as echarts from "echarts";
import { adminApi } from "../api/admin";
import { StatCard } from "../components/StatCard";
import { StatusBadge } from "../components/StatusBadge";
import type { Client, HealthStatus, IdentityUser, LoginLog, SecurityEvent, Session } from "../types/api";

export function DashboardPage() {
  const [users, setUsers] = useState<IdentityUser[]>([]);
  const [clients, setClients] = useState<Client[]>([]);
  const [sessions, setSessions] = useState<Session[]>([]);
  const [logs, setLogs] = useState<LoginLog[]>([]);
  const [events, setEvents] = useState<SecurityEvent[]>([]);
  const [health, setHealth] = useState<HealthStatus | null>(null);
  const [error, setError] = useState("");

  useEffect(() => {
    Promise.all([
      adminApi.listUsers({ page: 1, page_size: 10 }),
      adminApi.listClients(),
      adminApi.listSessions({ active_only: true }),
      adminApi.listLoginLogs({ page: 1, page_size: 20 }),
      adminApi.listSecurityEvents({ page: 1, page_size: 10 }),
      adminApi.health(),
    ])
      .then(([userResult, clientResult, sessionResult, logResult, eventResult, healthResult]) => {
        setUsers(userResult.items);
        setClients(clientResult);
        setSessions(sessionResult.items);
        setLogs(logResult.items);
        setEvents(eventResult.items);
        setHealth(healthResult);
      })
      .catch((err) => setError(err.message));
  }, []);

  useEffect(() => {
    const barElement = document.getElementById("login-bar");
    const pieElement = document.getElementById("login-method-pie");
    if (!barElement || !pieElement) return;

    const daily = buildDailyLoginSeries(logs);
    const methodSummary = buildLoginMethodSummary(logs);
    const bar = echarts.init(barElement);
    bar.setOption({
      tooltip: {},
      legend: { data: ["登录成功", "登录失败"] },
      grid: { left: 36, right: 18, top: 36, bottom: 28 },
      xAxis: { type: "category", data: daily.labels },
      yAxis: { type: "value" },
      series: [
        { name: "登录成功", type: "bar", data: daily.success, color: "#123967" },
        { name: "登录失败", type: "bar", data: daily.failed, color: "#8aa0ba" },
      ],
    });
    const pie = echarts.init(pieElement);
    pie.setOption({
      tooltip: { trigger: "item" },
      series: [
        {
          type: "pie",
          radius: ["45%", "72%"],
          data: methodSummary,
        },
      ],
    });
    const resize = () => {
      bar.resize();
      pie.resize();
    };
    window.addEventListener("resize", resize);
    return () => {
      window.removeEventListener("resize", resize);
      bar.dispose();
      pie.dispose();
    };
  }, [logs]);

  const activeUsers = useMemo(() => users.filter((item) => item.status === "active").length, [users]);

  return (
    <div>
      <Typography.Title level={3}>认证管理概览</Typography.Title>
      {error ? <Alert type="error" message={error} showIcon className="page-alert" /> : null}
      <Row gutter={[16, 16]}>
        <Col xs={24} md={12} xl={6}>
          <StatCard title="接入应用" value={clients.length} description="来自应用管理接口" icon={<AppstoreOutlined />} />
        </Col>
        <Col xs={24} md={12} xl={6}>
          <StatCard title="启用身份" value={activeUsers} description={`本页加载 ${users.length} 个用户`} icon={<TeamOutlined />} />
        </Col>
        <Col xs={24} md={12} xl={6}>
          <StatCard title="活跃会话" value={sessions.length} description="真实刷新令牌会话" icon={<SafetyOutlined />} />
        </Col>
        <Col xs={24} md={12} xl={6}>
          <StatCard title="服务状态" value={health?.status || "-"} description={health?.service || "等待健康检查"} icon={<ApiOutlined />} />
        </Col>
        <Col xs={24} xl={16}>
          <Card title="近 7 日登录趋势">
            <div id="login-bar" className="chart" />
          </Card>
        </Col>
        <Col xs={24} xl={8}>
          <Card title="登录方式分布">
            <div id="login-method-pie" className="chart" />
          </Card>
        </Col>
        <Col xs={24}>
          <Card title="最近登录审计">
            <Table
              rowKey="id"
              size="small"
              pagination={false}
              dataSource={logs.slice(0, 8)}
              columns={[
                { title: "时间", dataIndex: "created_at" },
                { title: "用户", dataIndex: "user_id" },
                { title: "系统", dataIndex: "client_id" },
                { title: "方式", render: (_, row) => loginMethodLabel(row.login_method) },
                { title: "状态", render: (_, row) => <StatusBadge success={row.success} /> },
                { title: "IP", dataIndex: "ip" },
              ]}
            />
          </Card>
        </Col>
        <Col xs={24}>
          <Card title="最近安全操作">
            <Table
              rowKey="id"
              size="small"
              pagination={false}
              dataSource={events.slice(0, 6)}
              columns={[
                { title: "时间", dataIndex: "created_at" },
                { title: "用户", dataIndex: "user_id" },
                { title: "事件", render: (_, row) => securityEventLabel(row.event_type) },
                { title: "目标", dataIndex: "target_id" },
                { title: "状态", render: (_, row) => <StatusBadge success={row.success} /> },
              ]}
            />
          </Card>
        </Col>
      </Row>
    </div>
  );
}

type DailyLoginSeries = {
  labels: string[];
  success: number[];
  failed: number[];
};

function buildDailyLoginSeries(logs: LoginLog[]): DailyLoginSeries {
  const today = new Date();
  const days = Array.from({ length: 7 }, (_, index) => {
    const date = new Date(today);
    date.setDate(today.getDate() - (6 - index));
    return date;
  });
  const keys = days.map((date) => date.toISOString().slice(0, 10));
  const buckets = new Map(keys.map((key) => [key, { success: 0, failed: 0 }]));

  logs.forEach((log) => {
    const key = new Date(log.created_at).toISOString().slice(0, 10);
    const bucket = buckets.get(key);
    if (!bucket) return;
    if (log.success) {
      bucket.success += 1;
    } else {
      bucket.failed += 1;
    }
  });

  return {
    labels: days.map((date) => `${date.getMonth() + 1}/${date.getDate()}`),
    success: keys.map((key) => buckets.get(key)?.success || 0),
    failed: keys.map((key) => buckets.get(key)?.failed || 0),
  };
}

function buildLoginMethodSummary(logs: LoginLog[]) {
  const counts = logs.reduce<Record<string, number>>((acc, log) => {
    acc[log.login_method] = (acc[log.login_method] || 0) + 1;
    return acc;
  }, {});
  return Object.entries(counts).map(([method, value]) => ({
    name: loginMethodLabel(method),
    value,
  }));
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
