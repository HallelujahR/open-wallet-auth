import { Navigate, Route, Routes } from "react-router-dom";
import { lazy, Suspense } from "react";
import { Spin } from "antd";
const ApplicationsPage = lazy(() =>
  import("./features/console/pages/ApplicationsPage").then((m) => ({
    default: m.ApplicationsPage,
  })),
);
const AuditLogsPage = lazy(() =>
  import("./features/console/pages/AuditLogsPage").then((m) => ({
    default: m.AuditLogsPage,
  })),
);
const DashboardPage = lazy(() =>
  import("./features/console/pages/DashboardPage").then((m) => ({
    default: m.DashboardPage,
  })),
);
const IdentitiesPage = lazy(() =>
  import("./features/console/pages/IdentitiesPage").then((m) => ({
    default: m.IdentitiesPage,
  })),
);
const AdminLoginPage = lazy(() =>
  import("./features/console/pages/LoginPage").then((m) => ({
    default: m.AdminLoginPage,
  })),
);
const SecurityEventsPage = lazy(() =>
  import("./features/console/pages/SecurityEventsPage").then((m) => ({
    default: m.SecurityEventsPage,
  })),
);
const SessionsPage = lazy(() =>
  import("./features/console/pages/SessionsPage").then((m) => ({
    default: m.SessionsPage,
  })),
);
const SettingsPage = lazy(() =>
  import("./features/console/pages/SettingsPage").then((m) => ({
    default: m.SettingsPage,
  })),
);
const UnifiedLoginPage = lazy(() =>
  import("./features/hosted-login/UnifiedLoginPage").then((m) => ({
    default: m.UnifiedLoginPage,
  })),
);
const AdminLayout = lazy(() =>
  import("./layouts/AdminLayout").then((m) => ({ default: m.AdminLayout })),
);
import { AuthLayout } from "./layouts/AuthLayout";
import { getAdminSession } from "./store/authStore";

function RequireAdmin({ children }: { children: JSX.Element }) {
  if (!getAdminSession()) {
    return <Navigate to="/console/login" replace />;
  }
  return children;
}

export function App() {
  return (
    <Suspense
      fallback={
        <div className="page-loading">
          <Spin size="large" tip="正在加载" />
        </div>
      }
    >
      <Routes>
        <Route path="/login" element={<UnifiedLoginPage />} />
        <Route element={<AuthLayout />}>
          <Route path="/console/login" element={<AdminLoginPage />} />
        </Route>
        <Route
          path="/console"
          element={
            <RequireAdmin>
              <AdminLayout />
            </RequireAdmin>
          }
        >
          <Route index element={<DashboardPage />} />
          <Route path="applications" element={<ApplicationsPage />} />
          <Route path="identities" element={<IdentitiesPage />} />
          <Route path="sessions" element={<SessionsPage />} />
          <Route path="audit-logs" element={<AuditLogsPage />} />
          <Route path="security-events" element={<SecurityEventsPage />} />
          <Route path="settings" element={<SettingsPage />} />
        </Route>
        <Route path="/" element={<Navigate to="/console" replace />} />
        <Route
          path="/admin/login"
          element={<Navigate to="/console/login" replace />}
        />
        <Route
          path="/applications"
          element={<Navigate to="/console/applications" replace />}
        />
        <Route
          path="/identities"
          element={<Navigate to="/console/identities" replace />}
        />
        <Route
          path="/sessions"
          element={<Navigate to="/console/sessions" replace />}
        />
        <Route
          path="/audit-logs"
          element={<Navigate to="/console/audit-logs" replace />}
        />
        <Route
          path="/security-events"
          element={<Navigate to="/console/security-events" replace />}
        />
        <Route
          path="/settings"
          element={<Navigate to="/console/settings" replace />}
        />
        <Route path="*" element={<Navigate to="/console" replace />} />
      </Routes>
    </Suspense>
  );
}
