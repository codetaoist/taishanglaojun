#!/usr/bin/env node

// Windows 端口 5173 自动清理并启动 Vite 开发服务
// - 若 5173 被占用：自动终止占用进程
// - 然后启动 Vite dev（严格端口）

import { spawn, execSync } from 'child_process';
import path from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const PORT = 5173;

function killPortOnWindows(port) {
  try {
    // 使用 PowerShell 获取监听该端口的进程 PID
    const cmdList = `powershell -NoProfile -Command "Get-NetTCPConnection -LocalPort ${port} -State Listen -ErrorAction SilentlyContinue | Select-Object -ExpandProperty OwningProcess | Sort-Object -Unique"`;
    const output = execSync(cmdList, { encoding: 'utf8' }).trim();
    if (!output) {
      console.log(`✅ 端口 ${port} 未被占用`);
      return;
    }

    const pids = output
      .split(/\r?\n/)
      .map(s => s.trim())
      .filter(Boolean);

    console.log(`⚠️ 检测到端口 ${port} 占用，进程: ${pids.join(', ')}`);

    // 逐个强制结束进程
    for (const pid of pids) {
      try {
        execSync(`taskkill /F /PID ${pid}`);
        console.log(`🔪 已结束占用进程 PID=${pid}`);
      } catch (e) {
        console.warn(`⚠️ 无法结束 PID=${pid}，尝试 PowerShell 停止`);
        try {
          execSync(`powershell -NoProfile -Command "Stop-Process -Id ${pid} -Force"`);
          console.log(`🔪 已停止占用进程 PID=${pid}（PowerShell）`);
        } catch (e2) {
          console.error(`❌ 结束进程失败 PID=${pid}:`, e2.message);
        }
      }
    }
    // 等待片刻确保端口释放
    execSync('powershell -NoProfile -Command "Start-Sleep -Seconds 1"');
  } catch (err) {
    console.warn(`⚠️ 检查/清理端口 ${port} 失败:`, err.message);
    // 兼容性回退：使用 netstat 查找占用端口的进程并结束
    try {
      const netstat = execSync(`cmd /c netstat -ano | findstr :${port}`, { encoding: 'utf8' });
      const lines = netstat.split(/\r?\n/).map(l => l.trim()).filter(Boolean);
      const pids = Array.from(new Set(lines.map(l => {
        const parts = l.split(/\s+/);
        return parts[parts.length - 1];
      }).filter(pid => /^(\d+)$/.test(pid))));
      if (pids.length > 0) {
        console.log(`⚠️ 通过 netstat 检测到端口 ${port} 占用，进程: ${pids.join(', ')}`);
        for (const pid of pids) {
          try {
            execSync(`taskkill /F /PID ${pid}`);
            console.log(`🔪 已结束占用进程 PID=${pid}`);
          } catch (e) {
            console.error(`❌ 结束进程失败 PID=${pid}:`, e.message);
          }
        }
        execSync('powershell -NoProfile -Command "Start-Sleep -Seconds 1"');
      } else {
        console.log(`✅ netstat 未发现端口 ${port} 占用`);
      }
    } catch (e2) {
      console.warn('⚠️ netstat 回退检查失败：', e2.message);
    }
  }
}

function startViteDev() {
  console.log('🚀 启动 Vite 开发服务器 (端口 5173, strictPort=true)');
  // 直接通过 Node 可执行文件运行 vite 的 JS 入口，避免 .cmd/spawn 兼容性问题
  const viteJs = path.join(__dirname, '../node_modules/vite/bin/vite.js');
  const child = spawn(process.execPath, [viteJs], {
    stdio: 'inherit',
    cwd: path.join(__dirname, '..'),
    shell: false,
  });

  child.on('exit', (code) => {
    console.log(`📦 Vite 进程退出，code=${code}`);
  });
}

// 仅在 Windows 环境执行专用清理逻辑
if (process.platform === 'win32') {
  killPortOnWindows(PORT);
  startViteDev();
} else {
  // 非 Windows 简单启动（端口与 strictPort 已在 vite.config.ts 配置）
  startViteDev();
}