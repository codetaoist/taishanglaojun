#!/usr/bin/env node

import http from 'http';

const url = process.env.BACKEND_HEALTH_URL || 'http://localhost:8080/health';
const maxAttempts = parseInt(process.env.MAX_ATTEMPTS || '15', 10);
const intervalMs = parseInt(process.env.INTERVAL_MS || '1000', 10);

function checkHealth() {
  return new Promise((resolve) => {
    const req = http.get(url, (res) => {
      let data = '';
      res.on('data', (chunk) => (data += chunk));
      res.on('end', () => {
        const ok = res.statusCode === 200;
        resolve({ ok, status: res.statusCode, body: data });
      });
    });
    req.on('error', (err) => resolve({ ok: false, error: err.message }));
  });
}

async function run() {
  console.log(`🔎 Checking backend health at: ${url}`);
  for (let attempt = 1; attempt <= maxAttempts; attempt++) {
    const result = await checkHealth();
    if (result.ok) {
      console.log(`✅ Backend healthy (status ${result.status}).`);
      process.exit(0);
    }
    console.log(`⏳ Attempt ${attempt}/${maxAttempts} failed. Retrying in ${intervalMs}ms...`);
    await new Promise((r) => setTimeout(r, intervalMs));
  }
  console.error('❌ Backend health check failed after max attempts.');
  process.exit(1);
}

run();