# sub2api OAuth Account Creator

自动化创建 sub2api OAuth 账户工具

## 功能

- ✅ 自动生成 OAuth 授权 URL
- ✅ 自动打开浏览器
- ✅ 监听授权回调
- ✅ 自动交换 Token
- ✅ 自动创建 sub2api 账户
- ✅ 本地加密存储邮箱密码

## 安装

```bash
cd tools/oauth-account-creator
npm install
```

## 配置

1. 复制配置文件:
```bash
cp config.example.yaml config.yaml
```

2. 修改配置:
```yaml
# sub2api 后端地址
sub2api_url: "http://localhost:8080"

# sub2api Admin API Token (从管理后台获取)
admin_token: "YOUR_ADMIN_TOKEN_HERE"

# 默认代理 ID (可选)
default_proxy_id: null

# 默认倍率
default_rate_multiplier: 1.0

# 回调端口 (必须与 OAuth 重定向 URI 匹配)
callback_port: 1455
```

## 使用

```bash
npm start
```

### 流程

1. **选择平台**: OpenAI / Anthropic / Gemini / Antigravity
2. **输入账户名称** (可选，默认：平台 - 邮箱)
3. **输入邮箱密码** (仅用于本地记录，方便下次使用)
4. **自动生成授权 URL** 并打开浏览器
5. **手动登录** (在浏览器中输入邮箱密码完成 OAuth 授权)
6. **自动完成后续步骤**:
   - 获取 authorization code
   - 交换 access_token / refresh_token
   - 创建 sub2api 账户

## 凭证存储

凭证默认加密存储在 `credentials.json`:

```json
{
  "encrypted": "...",
  "key": "...",
  "iv": "..."
}
```

如需禁用加密，在 `config.yaml` 中设置:
```yaml
encrypt_credentials: false
```

## 支持的 OAuth 平台

| 平台 | 状态 | 备注 |
|------|------|------|
| OpenAI | ✅ | 标准 OAuth 2.0 + PKCE |
| Anthropic | ✅ | Claude Code OAuth |
| Gemini | ✅ | Google OAuth |
| Antigravity | ✅ | 自定义 OAuth |

## 故障排除

### 回调服务器启动失败

确保端口 `1455` 未被占用:
```bash
# Windows
netstat -ano | findstr :1455

# Linux/Mac
lsof -i :1455
```

### 授权超时

- 检查浏览器是否正常打开
- 确认 OAuth 重定向 URI 配置正确
- 默认超时 5 分钟

### Token 交换失败

- 检查 `admin_token` 是否正确
- 确认 sub2api 后端正常运行
- 查看后端日志

## 安全提示

⚠️ **重要**:
- `credentials.json` 包含敏感信息，请勿上传到 Git
- 建议启用 `encrypt_credentials: true`
- 定期更新 `admin_token`

## 开发

```bash
# 开发模式 (自动重载)
npm run dev
```

## License

MIT
