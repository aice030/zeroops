# JaD1ng ZeroOps

这是一个零运维平台项目，包含多个微服务组件。

## 项目组件

### 文件存储服务 (File Storage Service)

位置: `storage/`

一个支持文本文件上传、下载和删除的存储服务，使用PostgreSQL作为存储后端。

**主要功能:**
- ✅ 文件上传到PostgreSQL数据库
- ✅ 文件下载
- ✅ 文件删除
- ✅ 文件信息查询
- ✅ 文件列表
- ✅ 健康检查
- ✅ 可扩展架构（支持未来扩展到S3）

**快速启动:**
```bash
cd storage
./start.sh
```

**API测试:**
```bash
cd storage
./test_api.sh
```

详细文档请查看 [storage/README.md](mock-s3-storage/storage/README.md)

## 技术栈

- **后端**: Go 1.21+
- **数据库**: PostgreSQL 15
- **容器化**: Docker & Docker Compose
- **架构**: 微服务架构，接口分离设计

## 开发环境要求

- Go 1.21 或更高版本
- Docker & Docker Compose
- PostgreSQL 12+ (通过Docker运行)

## 项目结构

```
JaD1ng_zeroops/
├── storage/                 # 文件存储服务
│   ├── cmd/                # 服务启动入口
│   ├── internal/           # 内部包
│   │   ├── handler/        # HTTP处理器
│   │   └── service/        # 业务逻辑服务
│   ├── docker-compose.yml  # 数据库配置
│   ├── start.sh           # 启动脚本
│   ├── test_api.sh        # API测试脚本
│   └── README.md          # 详细文档
├── mock-s3-storage/        # S3模拟存储服务
└── README.md              # 项目总览
```

## 许可证

本项目采用MIT许可证。