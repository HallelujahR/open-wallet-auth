import { formatTime } from "../../../utils/format";
import { loginMethodLabel, securityEventLabel } from "../../../utils/format";
import { PageHeader } from "../../../components/PageHeader";
import {
  ApiOutlined,
  AppstoreOutlined,
  SafetyOutlined,
  TeamOutlined,
} from "@ant-design/icons";
import { Alert, Button, Card, Col, Row, Skeleton, Table } from "antd";
import { useEffect, useRef, useState } from "react";
import * as echarts from "echarts/core";
import { BarChart, PieChart } from "echarts/charts";
import {
  GridComponent,
  LegendComponent,
  TooltipComponent,
} from "echarts/components";
import { CanvasRenderer } from "echarts/renderers";
echarts.use([
  BarChart,
  PieChart,
  GridComponent,
  LegendComponent,
  TooltipComponent,
  CanvasRenderer,
]);
import { adminApi } from "../../../api/admin";
import { StatCard } from "../../../components/StatCard";
import { StatusBadge } from "../../../components/StatusBadge";
import type {
  DashboardSummary,
  HealthStatus,
  LoginLog,
  SecurityEvent,
} from "../../../types/api";

export function DashboardPage() {
  const [summary, setSummary] = useState<DashboardSummary | null>(null);
  const [loading, setLoading] = useState(true);
  const [retry, setRetry] = useState(0);
  const barRef = useRef<HTMLDivElement>(null);
  const pieRef = useRef<HTMLDivElement>(null);
  const [logs, setLogs] = useState<LoginLog[]>([]);
  const [events, setEvents] = useState<SecurityEvent[]>([]);
  const [health, setHealth] = useState<HealthStatus | null>(null);
  const [error, setError] = useState("");

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    setError("");
    Promise.all([
      adminApi.dashboard(),
      adminApi.listLoginLogs({ page: 1, page_size: 8 }),
      adminApi.listSecurityEvents({ page: 1, page_size: 6 }),
      adminApi.health(),
    ])
      .then(([stats, logResult, eventResult, healthResult]) => {
        if (cancelled) return;
        setSummary(stats);
        setLogs(logResult.items);
        setEvents(eventResult.items);
        setHealth(healthResult);
      })
      .catch((err) => {
        if (!cancelled) setError(err.message);
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [retry]);

  useEffect(() => {
    const barElement = barRef.current;
    const pieElement = pieRef.current;
    if (!barElement || !pieElement || !summary || loading) return;

    const daily = buildDailyLoginSeries(summary);
    const methodSummary = buildLoginMethodSummary(summary);
    const bar = echarts.init(barElement);
    bar.setOption({
      tooltip: {},
      legend: { data: ["登录成功", "登录失败"] },
      grid: { left: 36, right: 18, top: 36, bottom: 28 },
      xAxis: { type: "category", data: daily.labels },
      yAxis: { type: "value" },
      series: [
        {
          name: "登录成功",
          type: "bar",
          data: daily.success,
          color: "#6c8aef",
          barMaxWidth: 24,
          itemStyle: { borderRadius: [4, 4, 0, 0] },
        },
        {
          name: "登录失败",
          type: "bar",
          data: daily.failed,
          color: "#e6b39c",
          barMaxWidth: 24,
          itemStyle: { borderRadius: [4, 4, 0, 0] },
        },
      ],
    });
    const pie = echarts.init(pieElement);
    pie.setOption({
      color: ["#6c8aef", "#7fc6b2", "#e6b39c", "#b6a1df", "#91b6d7"],
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
    const observer = new ResizeObserver(resize);
    observer.observe(barElement);
    observer.observe(pieElement);
    return () => {
      observer.disconnect();
      bar.dispose();
      pie.dispose();
    };
  }, [summary, loading]);

  return (
    <div>
      <PageHeader
        title="管理概览"
        description="认证服务的运行情况，一目了然。"
      />
      {error ? (
        <Alert
          type="error"
          message={error}
          showIcon
          className="page-alert"
          action={
            <Button onClick={() => setRetry((value) => value + 1)}>重试</Button>
          }
        />
      ) : null}
      {loading ? (
        <Card>
          <Skeleton active paragraph={{ rows: 10 }} />
        </Card>
      ) : (
        <>
          <Row gutter={[16, 16]}>
            <Col xs={24} md={12} xl={6}>
              <StatCard
                title="接入应用"
                value={summary?.clients ?? "—"}
                description="已接入的业务应用"
                icon={<AppstoreOutlined />}
              />
            </Col>
            <Col xs={24} md={12} xl={6}>
              <StatCard
                title="启用身份"
                value={summary?.active_users ?? "—"}
                description="全部启用状态的身份"
                icon={<TeamOutlined />}
              />
            </Col>
            <Col xs={24} md={12} xl={6}>
              <StatCard
                title="活跃会话"
                value={summary?.active_sessions ?? "—"}
                description="未过期且未吊销"
                icon={<SafetyOutlined />}
              />
            </Col>
            <Col xs={24} md={12} xl={6}>
              <StatCard
                title="服务状态"
                value={health?.status === "ok" ? "运行正常" : "未知"}
                description="认证服务健康检查"
                icon={<ApiOutlined />}
              />
            </Col>
            <Col xs={24} xl={16}>
              <Card
                title="近 7 日登录趋势"
                extra={<span className="chart-note">按本地时区统计</span>}
              >
                <div ref={barRef} className="chart" />
              </Card>
            </Col>
            <Col xs={24} xl={8}>
              <Card
                title="登录方式分布"
                extra={<span className="chart-note">近 7 日</span>}
              >
                <div ref={pieRef} className="chart" />
              </Card>
            </Col>
            <Col xs={24}>
              <Card title="最近登录审计">
                <Table
                  scroll={{ x: "max-content" }}
                  rowKey="id"
                  size="small"
                  pagination={false}
                  dataSource={logs.slice(0, 8)}
                  columns={[
                    {
                      title: "时间",
                      dataIndex: "created_at",
                      render: formatTime,
                    },
                    { title: "用户", dataIndex: "user_id" },
                    { title: "系统", dataIndex: "client_id" },
                    {
                      title: "方式",
                      render: (_, row) => loginMethodLabel(row.login_method),
                    },
                    {
                      title: "状态",
                      render: (_, row) => <StatusBadge success={row.success} />,
                    },
                    { title: "IP", dataIndex: "ip" },
                  ]}
                />
              </Card>
            </Col>
            <Col xs={24}>
              <Card title="最近安全操作">
                <Table
                  scroll={{ x: "max-content" }}
                  rowKey="id"
                  size="small"
                  pagination={false}
                  dataSource={events.slice(0, 6)}
                  columns={[
                    {
                      title: "时间",
                      dataIndex: "created_at",
                      render: formatTime,
                    },
                    { title: "用户", dataIndex: "user_id" },
                    {
                      title: "事件",
                      render: (_, row) => securityEventLabel(row.event_type),
                    },
                    { title: "目标", dataIndex: "target_id" },
                    {
                      title: "状态",
                      render: (_, row) => <StatusBadge success={row.success} />,
                    },
                  ]}
                />
              </Card>
            </Col>
          </Row>
        </>
      )}
    </div>
  );
}

function buildDailyLoginSeries(summary: DashboardSummary) {
  return {
    labels: summary.days.map((day) => day.slice(5).replace("-", "/")),
    success: summary.days.map((day) =>
      summary.login_buckets
        .filter((item) => item.day === day && item.success)
        .reduce((sum, item) => sum + item.count, 0),
    ),
    failed: summary.days.map((day) =>
      summary.login_buckets
        .filter((item) => item.day === day && !item.success)
        .reduce((sum, item) => sum + item.count, 0),
    ),
  };
}
function buildLoginMethodSummary(summary: DashboardSummary) {
  const counts = new Map<string, number>();
  summary.login_buckets.forEach((item) =>
    counts.set(item.method, (counts.get(item.method) || 0) + item.count),
  );
  return Array.from(counts, ([method, value]) => ({
    name: loginMethodLabel(method),
    value,
  }));
}
