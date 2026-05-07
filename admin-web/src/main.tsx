import React from "react";
import ReactDOM from "react-dom/client";
import { BrowserRouter } from "react-router-dom";
import { ConfigProvider } from "antd";
import zhCN from "antd/locale/zh_CN";
import { App } from "./App";
import "./styles/globals.css";

const routerBasename = import.meta.env.BASE_URL.replace(/\/$/, "");

ReactDOM.createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <ConfigProvider
      locale={zhCN}
      theme={{
        token: {
          colorPrimary: "#123967",
          borderRadius: 6,
          fontFamily:
            'Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif',
        },
      }}
    >
      <BrowserRouter basename={routerBasename}>
        <App />
      </BrowserRouter>
    </ConfigProvider>
  </React.StrictMode>,
);
