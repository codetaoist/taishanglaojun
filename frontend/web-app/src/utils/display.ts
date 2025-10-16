// 统一展示工具：日期、时间与标签颜色
// 避免各页面重复实现，确保一致的格式与显示效果

export function formatDate(value?: string | number | Date): string {
  if (!value) return '';
  try {
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) return String(value);
    return new Intl.DateTimeFormat('zh-CN', {
      year: 'numeric', month: '2-digit', day: '2-digit'
    }).format(date);
  } catch {
    return String(value);
  }
}

export function formatDateTime(value?: string | number | Date): string {
  if (!value) return '';
  try {
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) return String(value);
    return new Intl.DateTimeFormat('zh-CN', {
      year: '2-digit', month: '2-digit', day: '2-digit',
      hour: '2-digit', minute: '2-digit'
    }).format(date);
  } catch {
    return String(value);
  }
}

// 统一状态颜色
export function getStatusColor(status?: string): string {
  const s = (status || '').toLowerCase();
  switch (s) {
    case 'active':
    case 'enabled':
      return 'green';
    case 'inactive':
    case 'disabled':
      return 'orange';
    case 'banned':
    case 'error':
      return 'red';
    default:
      return 'default';
  }
}

// 统一角色颜色
export function getRoleColor(role?: string): string {
  const r = (role || '').toLowerCase();
  switch (r) {
    case 'admin':
    case 'administrator':
    case 'super_admin':
      return 'red';
    case 'moderator':
      return 'orange';
    default:
      return 'blue';
  }
}