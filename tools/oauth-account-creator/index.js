/**
 * sub2api OAuth Account Creator
 * 简化版：登录 → 获取 Token → 上传到 sub2api
 * 
 * 参考：D:\注册机\codex-console
 */

import axios from 'axios';
import express from 'express';
import open from 'open';
import readline from 'readline';
import crypto from 'crypto';
import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

// 配置加载
const CONFIG_PATH = path.join(__dirname, 'config.yaml');

let config = {
  sub2api_url: 'https://sub.chancexj.com',
  admin_token: '',
  callback_port: 1455,
  concurrency: 3,
  priority: 50,
  rate_multiplier: 1.0
};

// API 客户端
const apiClient = axios.create({
  timeout: 30000,
  validateStatus: (status) => status >= 200 && status < 500
});

// OAuth 常量
const OAUTH_CONFIG = {
  openai: {
    clientId: 'app_EMoamEEZ73f0CkXaXp7hrann',
    authUrl: 'https://auth.openai.com/oauth/authorize',
    tokenUrl: 'https://auth.openai.com/oauth/token',
    scope: 'openid profile email offline_access',
    redirectUri: 'http://localhost:1455/auth/callback'
  },
  anthropic: {
    clientId: 'codex_cli_00000000000000000000000000',
    authUrl: 'https://console.anthropic.com/oauth/authorize',
    tokenUrl: 'https://console.anthropic.com/api/oauth/token',
    scope: 'openid profile email offline_access',
    redirectUri: 'http://localhost:1455/auth/callback'
  }
};

// 加载配置文件
function loadConfig() {
  try {
    if (fs.existsSync(CONFIG_PATH)) {
      const yamlContent = fs.readFileSync(CONFIG_PATH, 'utf-8');
      const lines = yamlContent.split('\n');
      for (const line of lines) {
        if (line.trim() && !line.trim().startsWith('#')) {
          const colonIndex = line.indexOf(':');
          if (colonIndex === -1) continue;
          const key = line.substring(0, colonIndex).trim();
          const value = line.substring(colonIndex + 1).trim();
          if (key) {
            const parsedValue = value === 'null' ? null : 
                               value === 'true' ? true : 
                               value === 'false' ? false :
                               !isNaN(value) && value !== '' ? Number(value) : 
                               value.replace(/^["']|["']$/g, '');
            config[key] = parsedValue;
          }
        }
      }
      console.log('✅ 配置已加载:', CONFIG_PATH);
      console.log('   sub2api_url:', config.sub2api_url);
      console.log('   admin_token:', config.admin_token ? config.admin_token.substring(0, 20) + '...' : 'N/A');
    } else {
      console.log('⚠️  配置文件不存在，使用默认配置');
    }
  } catch (err) {
    console.error('❌ 配置加载失败:', err.message);
  }
}

// 生成 PKCE 参数
function generatePKCE() {
  const codeVerifier = crypto.randomBytes(64).toString('base64url');
  const codeChallenge = crypto.createHash('sha256').update(codeVerifier).digest('base64url');
  const state = crypto.randomBytes(16).toString('base64url');
  return { codeVerifier, codeChallenge, state };
}

// 生成授权 URL
function generateAuthUrl(platform) {
  const oauth = OAUTH_CONFIG[platform] || OAUTH_CONFIG.openai;
  const { codeVerifier, codeChallenge, state } = generatePKCE();
  
  const params = new URLSearchParams({
    client_id: oauth.clientId,
    response_type: 'code',
    redirect_uri: oauth.redirectUri,
    scope: oauth.scope,
    state: state,
    code_challenge: codeChallenge,
    code_challenge_method: 'S256',
    prompt: 'login'
  });
  
  const authUrl = `${oauth.authUrl}?${params.toString()}`;
  return { authUrl, state, codeVerifier, redirectUri: oauth.redirectUri };
}

// 启动回调服务器
function startCallbackServer(expectedState, codeVerifier, platform) {
  return new Promise((resolve, reject) => {
    const app = express();
    const server = app.listen(config.callback_port, () => {
      console.log(`📡 回调服务器已启动：http://localhost:${config.callback_port}`);
    });
    
    let resolved = false;
    
    app.get('/auth/callback', async (req, res) => {
      if (resolved) {
        res.send('<h1>✅ 授权已完成</h1><p>请返回工具查看结果</p>');
        return;
      }
      
      const { code, state, error } = req.query;
      
      if (error) {
        res.send(`<h1>❌ 授权失败</h1><p>${error}</p>`);
        server.close();
        reject(new Error(`OAuth error: ${error}`));
        return;
      }
      
      if (state !== expectedState) {
        res.send('<h1>❌ State 不匹配</h1>');
        server.close();
        reject(new Error('State mismatch'));
        return;
      }
      
      resolved = true;
      res.send('<h1>✅ 授权成功！</h1><p>正在处理 Token...</p><p>请等待工具自动继续</p>');
      
      try {
        // 交换 Token
        const tokenInfo = await exchangeCode(code, codeVerifier, platform);
        server.close();
        resolve(tokenInfo);
      } catch (err) {
        server.close();
        reject(err);
      }
    });
    
    // 超时处理
    setTimeout(() => {
      if (!resolved) {
        server.close();
        reject(new Error('回调超时（5 分钟）'));
      }
    }, 300000);
  });
}

// 交换授权码
async function exchangeCode(code, codeVerifier, platform) {
  console.log('\n🔄 正在交换 Token...');
  
  const oauth = OAUTH_CONFIG[platform] || OAUTH_CONFIG.openai;
  
  const params = new URLSearchParams({
    grant_type: 'authorization_code',
    code: code,
    redirect_uri: oauth.redirectUri,
    client_id: oauth.clientId,
    code_verifier: codeVerifier
  });
  
  try {
    const response = await axios.post(oauth.tokenUrl, params.toString(), {
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' }
    });
    
    const data = response.data;
    console.log('✅ Token 交换成功');
    console.log('   Email:', data.email || data.email_address || 'N/A');
    console.log('   Token 类型:', data.token_type || 'N/A');
    
    return {
      access_token: data.access_token,
      refresh_token: data.refresh_token || '',
      expires_at: data.expires_at || (Date.now() / 1000 + data.expires_in),
      email: data.email || data.email_address || '',
      account_id: data.account_uuid || data.org_uuid || '',
      platform: platform
    };
  } catch (err) {
    console.error('❌ Token 交换失败:', err.response?.data || err.message);
    throw err;
  }
}

// 上传到 sub2api
async function uploadToSub2api(tokenInfo) {
  console.log('\n📤 正在上传到 sub2api...');
  
  const accountData = {
    name: tokenInfo.email,
    platform: tokenInfo.platform,
    type: 'oauth',
    credentials: {
      access_token: tokenInfo.access_token,
      refresh_token: tokenInfo.refresh_token || '',
      expires_at: Math.floor(tokenInfo.expires_at),
      expires_in: 863999,
      organization_id: tokenInfo.account_id || '',
      chatgpt_account_id: tokenInfo.account_id || '',
      chatgpt_user_id: '',
      client_id: OAUTH_CONFIG[tokenInfo.platform]?.clientId || '',
      model_mapping: {
        'gpt-5.1': 'gpt-5.1',
        'gpt-5.1-codex': 'gpt-5.1-codex',
        'gpt-5.1-codex-max': 'gpt-5.1-codex-max',
        'gpt-5.1-codex-mini': 'gpt-5.1-codex-mini',
        'gpt-5.2': 'gpt-5.2',
        'gpt-5.2-codex': 'gpt-5.2-codex',
        'gpt-5.3': 'gpt-5.3',
        'gpt-5.3-codex': 'gpt-5.3-codex',
        'gpt-5.4': 'gpt-5.4',
        'gpt-5.4-mini': 'gpt-5.4-mini',
        'gpt-5.4-nano': 'gpt-5.4-nano'
      }
    },
    concurrency: config.concurrency || 3,
    priority: config.priority || 50,
    rate_multiplier: config.rate_multiplier || 1.0,
    auto_pause_on_expired: true
  };
  
  const payload = {
    data: {
      type: 'sub2api-data',
      version: 1,
      exported_at: new Date().toISOString(),
      proxies: [],
      accounts: [accountData]
    },
    skip_default_group_bind: true
  };
  
  const url = config.sub2api_url.replace(/\/$/, '') + '/api/v1/admin/accounts/data';
  const headers = {
    'Content-Type': 'application/json',
    'x-api-key': config.admin_token,
    'Idempotency-Key': `import-${Date.now()}`
  };
  
  try {
    const response = await apiClient.post(url, payload, { headers });
    
    if (response.status >= 200 && response.status < 300) {
      console.log('✅ 上传成功！');
      const result = response.data.data || response.data;
      console.log('   导入账号数:', result.imported_count || 'N/A');
      console.log('   跳过账号数:', result.skipped_count || 'N/A');
      return true;
    } else {
      console.error('❌ 上传失败:', response.status);
      console.error('   响应:', JSON.stringify(response.data, null, 2));
      return false;
    }
  } catch (err) {
    console.error('❌ 上传异常:');
    if (err.response) {
      console.error('   状态码:', err.response.status);
      console.error('   响应:', JSON.stringify(err.response.data, null, 2));
    } else {
      console.error('   错误:', err.message);
    }
    throw err;
  }
}

// 创建交互界面
function createInterface() {
  return readline.createInterface({
    input: process.stdin,
    output: process.stdout
  });
}

// 主流程
async function main() {
  console.log('🐶 sub2api OAuth Account Creator v2.0.0 (简化版)\n');
  console.log('================================================\n');
  
  loadConfig();
  
  // 重新初始化 API 客户端
  const baseUrl = config.sub2api_url.replace(/\/$/, '');
  apiClient.defaults.baseURL = baseUrl;
  apiClient.defaults.headers.common['x-api-key'] = config.admin_token;
  
  console.log('\n🔧 API 配置:');
  console.log('   baseURL:', apiClient.defaults.baseURL);
  console.log('   x-api-key:', config.admin_token ? config.admin_token.substring(0, 20) + '...' : 'N/A');
  
  const rl = createInterface();
  
  try {
    // 选择平台
    console.log('\n选择平台:');
    console.log('  1. OpenAI');
    console.log('  2. Anthropic');
    const platformInput = await new Promise(resolve => {
      rl.question('请输入选项 (1-2) [1]: ', resolve);
    });
    
    const platform = platformInput.trim() === '2' ? 'anthropic' : 'openai';
    console.log(`\n✅ 已选择：${platform.toUpperCase()}`);
    
    // 生成授权 URL
    const { authUrl, state, codeVerifier, redirectUri } = generateAuthUrl(platform);
    console.log('\n📡 授权 URL 已生成');
    console.log('   State:', state);
    
    // 启动回调服务器
    const callbackPromise = startCallbackServer(state, codeVerifier, platform);
    
    // 打开浏览器
    console.log('\n🌐 正在打开浏览器...');
    console.log('   Auth URL:', authUrl.substring(0, 80) + '...');
    console.log('   请在浏览器中完成登录和授权');
    console.log('   授权后会自动继续...\n');
    
    await open(authUrl);
    
    // 等待回调
    console.log('⏳ 等待授权回调...');
    const tokenInfo = await callbackPromise;
    
    // 上传到 sub2api
    await uploadToSub2api(tokenInfo);
    
    console.log('\n✅ 全部完成！\n');
    
  } catch (err) {
    console.error('\n❌ 流程失败:', err.message);
    if (err.response?.data) {
      console.error('   详情:', JSON.stringify(err.response.data, null, 2));
    }
  } finally {
    rl.close();
    process.exit(0);
  }
}

// 运行
main();
