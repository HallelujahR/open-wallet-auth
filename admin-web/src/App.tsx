import { Navigate, Route, Routes } from "react-router-dom";
import { AdminLayout } from "./layouts/AdminLayout";
import { AuthLayout } from "./layouts/AuthLayout";
import { ApplicationsPage } from "./pages/ApplicationsPage";
import { AuditLogsPage } from "./pages/AuditLogsPage";
import { DashboardPage } from "./pages/DashboardPage";
import { IdentitiesPage } from "./pages/IdentitiesPage";
import { AdminLoginPage } from "./pages/LoginPage";
import { SecurityEventsPage } from "./pages/SecurityEventsPage";
import { SessionsPage } from "./pages/SessionsPage";
import { SettingsPage } from "./pages/SettingsPage";
import { UnifiedLoginPage } from "./pages/UnifiedLoginPage";
import { getAdminSession } from "./store/authStore";

function RequireAdmin({ children }: { children: JSX.Element }) {
  if (!getAdminSession()) {
    return <Navigate to="/console/login" replace />;
  }
  return children;
}

export function App() {
  return (
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
      <Route path="/admin/login" element={<Navigate to="/console/login" replace />} />
      <Route path="/applications" element={<Navigate to="/console/applications" replace />} />
      <Route path="/identities" element={<Navigate to="/console/identities" replace />} />
      <Route path="/sessions" element={<Navigate to="/console/sessions" replace />} />
      <Route path="/audit-logs" element={<Navigate to="/console/audit-logs" replace />} />
      <Route path="/security-events" element={<Navigate to="/console/security-events" replace />} />
      <Route path="/settings" element={<Navigate to="/console/settings" replace />} />
      <Route path="*" element={<Navigate to="/console" replace />} />
    </Routes>
  );
}
