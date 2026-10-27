import { SafetyCertificateOutlined } from "@ant-design/icons";
import type { ReactNode } from "react";

export function AuthShell({
  children,
  brand = "Open Wallet Auth",
  mark = "O",
  context = "统一身份认证",
  heroTitle = "一个账号，\n连接你的应用。",
  admin = false,
}: {
  children: ReactNode;
  brand?: string;
  mark?: string;
  context?: string;
  heroTitle?: string;
  admin?: boolean;
}) {
  return (
    <main className="auth-shell">
      <header className="auth-topbar">
        <div className="brand-lockup">
          <span className="brand-mark">{mark}</span>
          <span>{brand}</span>
        </div>
        <span className="auth-context">
          <SafetyCertificateOutlined />
          {context}
        </span>
      </header>
      <div className="auth-stage">
        <section className="auth-intro">
          <span className="eyebrow">
            {admin ? "IDENTITY CONSOLE" : "ONE IDENTITY. EVERY CONNECTION."}
          </span>
          <h1>
            {admin ? (
              <>
                每个身份，
                <br />
                清晰掌握。
              </>
            ) : (
              heroTitle
            )}
          </h1>
          <p>
            {admin
              ? "在一个地方管理应用接入、用户身份与登录安全。"
              : "选择你的账号，安全、便捷地继续访问。"}
          </p>
          <div className="auth-assurance">
            <SafetyCertificateOutlined />
            <span>{admin ? "统一管理 · 清晰可控" : "安全认证 · 由你选择"}</span>
          </div>
          <div className="auth-orbit" aria-hidden="true">
            <div className="orbit-node orbit-one">ID</div>
            <div className="orbit-core">{mark}</div>
            <div className="orbit-node orbit-two">✓</div>
          </div>
        </section>
        <section
          className="auth-card"
          aria-label={admin ? "管理员登录" : "账号登录"}
        >
          {children}
        </section>
      </div>
      <footer className="auth-footer">
        <span>{brand}</span>
        <span>安全连接，从身份开始</span>
      </footer>
    </main>
  );
}
