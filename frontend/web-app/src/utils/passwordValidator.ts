export interface PasswordStrength {
  score: number; // 0-4 分数
  level: 'weak' | 'fair' | 'good' | 'strong' | 'very-strong';
  feedback: string[];
  color: string;
  percentage: number;
}

export interface PasswordValidationRule {
  test: (password: string) => boolean;
  message: string;
  weight: number;
}

// 密码验证规则
export const passwordRules: PasswordValidationRule[] = [
  {
    test: (password: string) => password.length >= 8,
    message: '至少8个字符',
    weight: 1
  },
  {
    test: (password: string) => /[a-z]/.test(password),
    message: '包含小写字母',
    weight: 1
  },
  {
    test: (password: string) => /[A-Z]/.test(password),
    message: '包含大写字母',
    weight: 1
  },
  {
    test: (password: string) => /\d/.test(password),
    message: '包含数字',
    weight: 1
  },
  {
    test: (password: string) => /[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]/.test(password),
    message: '包含特殊字符',
    weight: 1
  },
  {
    test: (password: string) => password.length >= 12,
    message: '至少12个字符（推荐）',
    weight: 0.5
  },
  {
    test: (password: string) => !/(.)\1{2,}/.test(password),
    message: '避免连续重复字符',
    weight: 0.5
  },
  {
    test: (password: string) => !/^(123|abc|qwe|password|admin)/i.test(password),
    message: '避免常见密码模式',
    weight: 0.5
  }
];

// 常见弱密码列表
const commonWeakPasswords = [
  'password', '123456', '123456789', 'qwerty', 'abc123', 
  'password123', 'admin', 'letmein', 'welcome', 'monkey',
  '1234567890', 'qwertyuiop', 'asdfghjkl', 'zxcvbnm'
];

/**
 * 检查密码强度
 */
export const checkPasswordStrength = (password: string): PasswordStrength => {
  if (!password) {
    return {
      score: 0,
      level: 'weak',
      feedback: ['请输入密码'],
      color: '#ff4d4f',
      percentage: 0
    };
  }

  let score = 0;
  const feedback: string[] = [];
  const passedRules: string[] = [];
  const failedRules: string[] = [];

  // 检查每个规则
  passwordRules.forEach(rule => {
    if (rule.test(password)) {
      score += rule.weight;
      passedRules.push(rule.message);
    } else {
      failedRules.push(rule.message);
    }
  });

  // 检查是否为常见弱密码
  if (commonWeakPasswords.some(weak => password.toLowerCase().includes(weak.toLowerCase()))) {
    score -= 1;
    failedRules.push('避免使用常见密码');
  }

  // 长度奖励
  if (password.length >= 16) {
    score += 0.5;
  }

  // 确保分数在合理范围内
  score = Math.max(0, Math.min(5, score));

  // 生成反馈
  if (failedRules.length > 0) {
    feedback.push(...failedRules.slice(0, 3)); // 只显示前3个建议
  }

  if (score >= 4) {
    feedback.unshift('密码强度很好！');
  } else if (score >= 3) {
    feedback.unshift('密码强度良好');
  } else if (score >= 2) {
    feedback.unshift('密码强度一般');
  } else {
    feedback.unshift('密码强度较弱');
  }

  // 确定强度等级
  let level: PasswordStrength['level'];
  let color: string;
  
  if (score >= 4.5) {
    level = 'very-strong';
    color = '#52c41a';
  } else if (score >= 3.5) {
    level = 'strong';
    color = '#73d13d';
  } else if (score >= 2.5) {
    level = 'good';
    color = '#faad14';
  } else if (score >= 1.5) {
    level = 'fair';
    color = '#fa8c16';
  } else {
    level = 'weak';
    color = '#ff4d4f';
  }

  const percentage = Math.min(100, (score / 5) * 100);

  return {
    score,
    level,
    feedback,
    color,
    percentage
  };
};

/**
 * 验证密码是否满足最低要求
 */
export const validatePassword = (password: string): { valid: boolean; errors: string[] } => {
  const errors: string[] = [];

  if (!password) {
    errors.push('密码不能为空');
    return { valid: false, errors };
  }

  if (password.length < 8) {
    errors.push('密码至少需要8个字符');
  }

  if (!/[a-z]/.test(password)) {
    errors.push('密码必须包含小写字母');
  }

  if (!/[A-Z]/.test(password)) {
    errors.push('密码必须包含大写字母');
  }

  if (!/\d/.test(password)) {
    errors.push('密码必须包含数字');
  }

  if (!/[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]/.test(password)) {
    errors.push('密码必须包含特殊字符');
  }

  // 检查常见弱密码
  if (commonWeakPasswords.some(weak => password.toLowerCase().includes(weak.toLowerCase()))) {
    errors.push('请避免使用常见密码');
  }

  return {
    valid: errors.length === 0,
    errors
  };
};

/**
 * 生成密码强度描述文本
 */
export const getPasswordStrengthText = (level: PasswordStrength['level']): string => {
  const texts = {
    'weak': '弱',
    'fair': '一般',
    'good': '良好',
    'strong': '强',
    'very-strong': '很强'
  };
  return texts[level];
};