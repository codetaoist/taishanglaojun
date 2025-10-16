// 测试JWT的Node.js脚本
const https = require('https');
const http = require('http');

// 创建一个简单的HTTP请求函数
function makeRequest(options, data) {
    return new Promise((resolve, reject) => {
        const req = http.request(options, (res) => {
            let body = '';
            res.on('data', (chunk) => body += chunk);
            res.on('end', () => {
                try {
                    resolve(JSON.parse(body));
                } catch (e) {
                    resolve(body);
                }
            });
        });
        
        req.on('error', reject);
        
        if (data) {
            req.write(data);
        }
        req.end();
    });
}

// Base64 URL解码
function base64UrlDecode(str) {
    str = str.replace(/-/g, '+').replace(/_/g, '/');
    while (str.length % 4) {
        str += '=';
    }
    return Buffer.from(str, 'base64').toString('utf8');
}

async function testJWT() {
    try {
        console.log('1. 登录获取token...');
        
        // 登录请求
        const loginOptions = {
            hostname: 'localhost',
            port: 8080,
            path: '/api/v1/auth/login',
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            }
        };
        
        const loginData = JSON.stringify({
            username: 'admin',
            password: 'admin123'
        });
        
        const loginResponse = await makeRequest(loginOptions, loginData);
        console.log('登录响应:', loginResponse);
        
        if (!loginResponse.data || !loginResponse.data.token) {
            console.log('未获取到token');
            return;
        }
        
        const token = loginResponse.data.token;
        console.log('Token获取成功');
        
        // 解析JWT
        console.log('\n2. 解析JWT payload...');
        const parts = token.split('.');
        if (parts.length !== 3) {
            console.log('JWT格式错误');
            return;
        }
        
        const payload = JSON.parse(base64UrlDecode(parts[1]));
        console.log('JWT中的用户信息:');
        console.log('  用户ID:', payload.user_id);
        console.log('  用户名:', payload.username);
        console.log('  角色:', payload.role);
        
        // 获取用户信息
        console.log('\n3. 获取API用户信息...');
        const userOptions = {
            hostname: 'localhost',
            port: 8080,
            path: '/api/v1/user/me',
            method: 'GET',
            headers: {
                'Authorization': `Bearer ${token}`
            }
        };
        
        const userInfo = await makeRequest(userOptions);
        console.log('API返回的用户信息:');
        console.log('完整响应:', JSON.stringify(userInfo, null, 2));
        
        // 处理嵌套的data结构
        const userData = userInfo.data || userInfo;
        console.log('  用户ID:', userData.user_id);
        console.log('  用户名:', userData.username);
        console.log('  角色:', userData.role);
        
        // 比较结果
        console.log('\n4. 比较结果:');
        if (payload.user_id === userData.user_id) {
            console.log('  用户ID匹配 ✓');
        } else {
            console.log('  用户ID不匹配 ✗');
            console.log('  JWT:', payload.user_id);
            console.log('  API:', userData.user_id);
        }
        
    } catch (error) {
        console.error('错误:', error.message);
    }
}

testJWT();