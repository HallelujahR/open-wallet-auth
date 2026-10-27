export function formatTime(value?: string) {
  if (!value) return "—";
  const date = new Date(value);
  return Number.isNaN(date.getTime())
    ? "—"
    : date.toLocaleString("zh-CN", { hour12: false });
}
export function loginMethodLabel(method: string) {
  return (
    (
      {
        password: "邮箱密码",
        wallet: "钱包",
        oauth: "第三方账号",
        phone: "手机验证码",
        refresh: "会话续签",
      } as Record<string, string>
    )[method] ||
    method ||
    "—"
  );
}
export function securityEventLabel(event: string) {
  return (
    (
      {
        change_password: "修改密码",
        reset_password: "重置密码",
        bind_email: "绑定邮箱",
        bind_phone: "绑定手机",
        bind_wallet: "绑定钱包",
        unbind_email: "解绑邮箱",
        unbind_phone: "解绑手机",
        unbind_wallet: "解绑钱包",
        unbind_oauth: "解绑第三方账号",
      } as Record<string, string>
    )[event] ||
    event ||
    "—"
  );
}
