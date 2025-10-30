# 环境变量配置说明

本项目支持通过环境变量覆盖配置文件中的设置，提供更灵活的配置管理方式。

## 环境变量文件加载顺序

系统会按以下优先级顺序加载环境变量文件：

1. `.env.local` - 本地环境配置（最高优先级，不应提交到版本控制）
2. `.env.{环境名}` - 特定环境配置（如 `.env.production`, `.env.development`）
3. `.env` - 默认环境配置

环境名通过 `APP_ENV` 或 `ENVIRONMENT` 环境变量确定，默认为 `development`。

## 环境变量命名规则

环境变量名称遵循以下规则：

1. **前缀规则**：每个配置文件都有对应的环境变量前缀
   - `database.yaml` → `DATABASE_` 前缀
   - `redis.yaml` → `REDIS_` 前缀
   - `rabbitmq.yaml` → `RABBITMQ_` 前缀
   - `storage.yaml` → `STORAGE_` 前缀
   - 等等...

2. **键名转换**：
   - 配置文件中的 `.` 和 `-` 会转换为 `_`
   - 所有字母转换为大写
   - 嵌套配置用 `_` 连接

## 配置示例

### 数据库配置

配置文件 `database.yaml`：
```yaml
host: localhost
port: 3306
username: root
password: secret
```

对应环境变量：
```bash
DATABASE_HOST=localhost
DATABASE_PORT=3306
DATABASE_USERNAME=root
DATABASE_PASSWORD=secret
```

### 嵌套配置

配置文件 `rabbitmq.yaml`：
```yaml
connection:
  host: localhost
  port: 5672
  username: guest
tls:
  enabled: false
```

对应环境变量：
```bash
RABBITMQ_CONNECTION_HOST=localhost
RABBITMQ_CONNECTION_PORT=5672
RABBITMQ_CONNECTION_USERNAME=guest
RABBITMQ_TLS_ENABLED=false
```

### 存储配置

配置文件 `storage.yaml`：
```yaml
default: local
drivers:
  s3:
    region: us-east-1
    bucket: my-bucket
    access-key-id: your-key
```

对应环境变量：
```bash
STORAGE_DEFAULT=local
STORAGE_DRIVERS_S3_REGION=us-east-1
STORAGE_DRIVERS_S3_BUCKET=my-bucket
STORAGE_DRIVERS_S3_ACCESS_KEY_ID=your-key
```

## 使用方法

### 1. 复制示例文件

```bash
cp .env.example .env
```

### 2. 编辑环境变量

根据你的环境需求修改 `.env` 文件中的值。

### 3. 环境特定配置

为不同环境创建特定的配置文件：

```bash
# 生产环境
cp .env.example .env.production

# 测试环境  
cp .env.example .env.testing

# 开发环境
cp .env.example .env.development
```

### 4. 设置环境

通过环境变量指定当前环境：

```bash
export APP_ENV=production
# 或
export ENVIRONMENT=production
```

## 优先级说明

配置值的优先级从高到低：

1. 环境变量
2. `.env.local` 文件
3. `.env.{环境名}` 文件
4. `.env` 文件
5. 配置文件默认值

## 安全注意事项

1. **不要提交敏感信息**：`.env.local` 和包含敏感信息的 `.env` 文件不应提交到版本控制
2. **使用 .gitignore**：确保敏感的环境变量文件被忽略
3. **生产环境**：在生产环境中直接设置系统环境变量，而不是依赖文件

## 调试

如果环境变量没有生效，检查：

1. 环境变量名称是否正确（注意大小写和下划线）
2. 环境变量文件是否存在且可读
3. 应用启动时的日志输出，确认哪些环境文件被加载

启动时会输出类似信息：
```
Loaded environment file: /path/to/.env
Loaded environment file: /path/to/.env.development
```