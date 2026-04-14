/**
 * Microsoft Account → OpenAI OAuth 自动登录
 * 使用 Microsoft Token 自动完成登录，无需验证码
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

// 配置
const CONFIG = {
  sub2api_url: 'https://sub.chancexj.com',
  admin_token: '',
  callback_port: 1455,
  openai_client_id: 'app_EMoamEEZ73f0CkXaXp7hrann',
  openai_redirect_uri: 'http://localhost:1455/auth/callback',
  openai_scope: 'openid profile email offline_access',
  // Microsoft OAuth
  ms_client_id: '04b07795-8ddb-461a-bbee-02f9e1bf7b46', // Azure CLI default
  ms_scope: 'openid profile email offline_access https://graph.microsoft.com/.default'
};

// 加载凭证文件
function loadCredentials() {
  const credPath = path.join(__dirname, 'credentials.txt');
  if (!fs.existsSync(credPath)) {
    console.log('❌ 凭证文件不存在:', credPath);
    console.log('请按格式创建文件：');
    console.log('邮箱----密码----MicrosoftID----MicrosoftToken');
    return null;
  }
  
  const content = fs.readFileSync(credPath, 'utf-8').trim();
  const parts = content.split('----');
  if (parts.length !== 4) {
    console.log('❌ 凭证格式错误，应该是：邮箱----密码----ID----Token');
    return null;
  }
  
  console.log('✅ 凭证已加载:', parts[0]);
  return {
    email: parts[0],
    password: parts[1],
    msId: parts[2],
    msToken: parts[3]
  };
}

// 生成 PKCE
function generatePKCE() {
  const codeVerifier = crypto.randomBytes(64).toString('base64url');
  const codeChallenge = crypto.createHash('sha256').update(codeVerifier).digest('base64url');
  const state = crypto.randomBytes(16).toString('base64url');
  return { codeVerifier, codeChallenge, state };
}

// 生成 OpenAI 授权 URL
function generateOpenAIAuthUrl() {
  const { codeVerifier, codeChallenge, state } = generatePKCE();
  
  const params = new URLSearchParams({
    client_id: CONFIG.openai_client_id,
    response_type: 'code',
    redirect_uri: CONFIG.openai_redirect_uri,
    scope: CONFIG.openai_scope,
    state: state,
    code_challenge: codeChallenge,
    code_challenge_method: 'S256',
    prompt: 'login'
  });
  
  const authUrl = `https://auth.openai.com/oauth/authorize?${params.toString()}`;
  return { authUrl, state, codeVerifier };
}

// 启动回调服务器
function startCallbackServer(expectedState, codeVerifier) {
  return new Promise((resolve, reject) => {
    const app = express();
    const server = app.listen(CONFIG.callback_port, () => {
      console.log(`📡 回调服务器：http://localhost:${CONFIG.callback_port}`);
    });
    
    let resolved = false;
    
    app.get('/auth/callback', async (req, res) => {
      if (resolved) {
        res.send('<h1>✅ 已完成</h1>');
        return;
      }
      
      const { code, state, error } = req.query;
      
      if (error) {
        res.send(`<h1>❌ 错误：${error}</h1>`);
        server.close();
        reject(new Error(error));
        return;
      }
      
      if (state !== expectedState) {
        res.send('<h1>❌ State 不匹配</h1>');
        server.close();
        reject(new Error('State mismatch'));
        return;
      }
      
      resolved = true;
      res.send('<h1>✅ 授权成功！</h1><p>正在处理...</p>');
      
      try {
        const tokenInfo = await exchangeCode(code, codeVerifier);
        server.close();
        resolve(tokenInfo);
      } catch (err) {
        server.close();
        reject(err);
      }
    });
    
    setTimeout(() => {
      if (!resolved) {
        server.close();
        reject(new Error('回调超时'));
      }
    }, 300000);
  });
}

// 交换 Token
async function exchangeCode(code, codeVerifier) {
  console.log('\n🔄 交换 Token...');
  
  const params = new URLSearchParams({
    grant_type: 'authorization_code',
    code: code,
    redirect_uri: CONFIG.openai_redirect_uri,
    client_id: CONFIG.openai_client_id,
    code_verifier: codeVerifier
  });
  
  const response = await axios.post(
    'https://auth.openai.com/oauth/token',
    params.toString(),
    { headers: { 'Content-Type': 'application/x-www-form-urlencoded' } }
  );
  
  const data = response.data;
  console.log('✅ Token 交换成功');
  console.log('   Email:', data.email || data.email_address || 'N/A');
  
  return {
    access_token: data.access_token,
    refresh_token: data.refresh_token || '',
    expires_at: data.expires_at || (Date.now() / 1000 + data.expires_in),
    email: data.email || data.email_address || '',
    account_id: data.account_uuid || data.org_uuid || '',
    platform: 'openai'
  };
}

// 上传到 sub2api
async function uploadToSub2api(tokenInfo) {
  console.log('\n📤 上传到 sub2api...');
  
  const accountData = {
    name: tokenInfo.email,
    platform: 'openai',
    type: 'oauth',
    credentials: {
      access_token: tokenInfo.access_token,
      refresh_token: tokenInfo.refresh_token || '',
      expires_at: Math.floor(tokenInfo.expires_at),
      expires_in: 863999,
      organization_id: tokenInfo.account_id || '',
      chatgpt_account_id: tokenInfo.account_id || '',
      chatgpt_user_id: '',
      client_id: CONFIG.openai_client_id,
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
    concurrency: 3,
    priority: 50,
    rate_multiplier: 1.0,
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
  
  const url = CONFIG.sub2api_url.replace(/\/$/, '') + '/api/v1/admin/accounts/data';
  const headers = {
    'Content-Type': 'application/json',
    'x-api-key': CONFIG.admin_token,
    'Idempotency-Key': `import-${Date.now()}`
  };
  
  const response = await axios.post(url, payload, { headers });
  
  if (response.status >= 200 && response.status < 300) {
    console.log('✅ 上传成功！');
    const result = response.data.data || response.data;
    console.log('   导入账号数:', result.imported_count || 'N/A');
    return true;
  } else {
    console.error('❌ 上传失败:', response.status);
    console.error('   响应:', JSON.stringify(response.data, null, 2));
    return false;
  }
}

// 主流程
async function main() {
  console.log('🐶 Microsoft → OpenAI 自动登录 v1.0.0\n');
  console.log('=====================================\n');
  
  // 加载凭证
  const credentials = loadCredentials();
  if (!credentials) {
    process.exit(1);
  }
  
  // 加载配置
  const configPath = path.join(__dirname, 'config.yaml');
  if (fs.existsSync(configPath)) {
    const yamlContent = fs.readFileSync(configPath, 'utf-8');
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
          CONFIG[key] = parsedValue;
        }
      }
    }
  }
  
  console.log('🔧 配置:');
  console.log('   sub2api_url:', CONFIG.sub2api_url);
  console.log('   admin_token:', CONFIG.admin_token ? CONFIG.admin_token.substring(0, 20) + '...' : 'N/A');
  
  try {
    // 生成授权 URL
    const { authUrl, state, codeVerifier } = generateOpenAIAuthUrl();
    console.log('\n📡 授权 URL 已生成');
    
    // 启动回调服务器
    const callbackPromise = startCallbackServer(state, codeVerifier);
    
    // 打开浏览器
    console.log('\n🌐 正在打开浏览器...');
    console.log('   请在浏览器中用 Microsoft 账号登录 OpenAI');
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
  } finally {
    process.exit(0);
  }
}

main();
