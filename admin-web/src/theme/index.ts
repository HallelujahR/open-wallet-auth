import type { ThemeConfig } from "antd";

export const appTheme: ThemeConfig = {
  token: {
    colorPrimary: "#3864e8",
    colorInfo: "#3864e8",
    colorSuccess: "#16876b",
    colorWarning: "#b87916",
    colorError: "#d74751",
    colorText: "#202838",
    colorTextSecondary: "#738096",
    colorBgLayout: "#f6f8fb",
    colorBorder: "#dfe5ee",
    colorBorderSecondary: "#edf0f5",
    borderRadius: 10,
    fontSize: 14,
    controlHeight: 38,
    fontFamily:
      'Inter, -apple-system, BlinkMacSystemFont, "Segoe UI", "PingFang SC", "Microsoft YaHei", sans-serif',
    boxShadow: "0 8px 32px rgba(31, 45, 75, 0.07)",
    boxShadowSecondary: "0 4px 20px rgba(31, 45, 75, 0.06)",
  },
  components: {
    Layout: { headerBg: "#ffffff", siderBg: "#ffffff", bodyBg: "#f6f8fb" },
    Menu: {
      itemBg: "transparent",
      itemSelectedBg: "#edf2ff",
      itemSelectedColor: "#3159d4",
      itemHeight: 44,
      itemBorderRadius: 8,
      iconSize: 18,
    },
    Button: { primaryShadow: "none", defaultShadow: "none", fontWeight: 500 },
    Card: { headerFontSize: 15, headerHeight: 56, paddingLG: 24 },
    Table: {
      headerBg: "#f8fafc",
      headerColor: "#647187",
      headerSplitColor: "transparent",
      cellPaddingBlock: 15,
      rowHoverBg: "#f8faff",
    },
    Tabs: { titleFontSize: 14, horizontalItemGutter: 24 },
    Typography: { titleMarginTop: 0, titleMarginBottom: 12 },
  },
};
