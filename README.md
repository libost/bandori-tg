# 🎸 bandori-tg

一个用于手游 **《BanG Dream! 少女乐团派对！》**（邦 / Bandori）的 Telegram 机器人，提供查卡、查活动、档线预测等实用功能，支持多服务区、多语言，数据来源于 [Bestdori](https://bestdori.com)。

[![License: MIT](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.27-blue.svg)](go.mod)

---

## ✨ 功能特性

- 🃏 **查卡**：通过 `/cards` 指令或中文文本指令「查卡」查询卡片，自动生成卡片图片，包含稀有度、属性、数值、技能等信息。
- 🎉 **活动查询**：通过 `/events` 指令查看活动详情（活动类型、时间、加成角色、加成卡、奖励卡等）。
- 📈 **档线预测**：通过 `/fsx` 指令查看活动预测档线，自动生成实时档线图表（含更新时间、当前档线、预测线与当前速度）。
- 🌍 **多服务区支持**：日服 (`jp`)、国际服 (`en`)、台服 (`tw`)、韩服 (`kr`)、国服 (`cn`)。
- 🗣️ **多语言**：界面语言支持简体中文 / 繁体中文 / English；查询语言支持 中文简体 / 中文繁體 / English / 日本語 / 한국어，且两者可独立设置。
- 🕐 **数据自动更新**：内置定时任务，每日多次从 Bestdori 拉取最新数据。
- 🖼️ **富媒体消息**：使用 Telegram 富文本消息（表格、图片、日期时间组件）展示信息。
- 🛡️ **管理员机制**：通过令牌认证管理员，可一键为所有语言设置 Bot 命令菜单。

## 🏗️ 技术栈

| 组件 | 说明 |
| --- | --- |
| 语言 | Go 1.26 |
| 框架 | [gotgbot](https://github.com/PaulSonOfLars/gotgbot) (v2) |
| 数据库 | SQLite（[modernc.org/sqlite](https://modernc.org/sqlite)，纯 Go 实现，支持 `CGO_ENABLED=0` 交叉编译） |
| 图片生成 | [fogleman/gg](https://github.com/fogleman/gg) + [tdewolff/canvas](https://github.com/tdewolff/canvas) |
| 定时任务 | [robfig/cron](https://github.com/robfig/cron) |
| 日志 | [lumberjack](https://github.com/natefinch/lumberjack)（滚动日志） |
| 配置 | YAML（[goccy/go-yaml](https://github.com/goccy/go-yaml)） |

## 📁 目录结构

```
bandori-tg/
├── main.go              # 程序入口：初始化、轮询/Webhook、定时任务
├── config.yaml          # 配置文件（被 .gitignore 忽略，从 example.yaml 复制）
├── example.yaml         # 配置示例
├── band/                # 乐队数据处理
├── cards/               # 查卡逻辑 + 卡片图片生成
├── characters/          # 角色数据处理
├── events/              # 活动详情 + 档线预测（/fsx）(档线预测功能将在v1.1版本推出)
├── gacha/               # 卡池数据处理（TODO）
├── skills/              # 技能数据处理
├── recent/              # 最新资讯数据处理
├── dynamic/             # 动态资讯转发（TODO）
├── callback/            # 内联键盘回调处理
├── commands/            # 基础指令（start/help/lang/about 等）
├── config/              # 配置加载
├── constants/           # 常量与数据结构定义
├── database/            # SQLite 数据库（用户偏好、管理员等）
├── i18n/                # 多语言文案（en / zh-hans / zh-hant）
├── utils/               # 数据抓取与定时刷新
├── version/             # 版本信息（由 ldflags 注入）
├── res/                 # 运行时数据（数据库、缓存等，运行时自动创建）
├── logs/                # 运行日志（运行时自动创建）
└── .github/workflows/   # CI/CD：自动构建与发布
```

## 🚀 快速开始

### 1. 准备工作

1. 通过 [@BotFather](https://t.me/BotFather) 创建 Telegram Bot，获取 Bot Token。
2. 克隆本仓库并进入目录。

### 2. 配置

复制配置示例并填入你的信息：

```bash
cp example.yaml config.yaml
```

| 配置项 | 说明 |
| --- | --- |
| `general.token` | 你的 Telegram Bot Token（必填） |
| `general.admin_token` | 管理员令牌，用于 `/setadmin` 认证（务必保密） |
| `proxy` | 代理配置，可选 `http` / `socks5`，默认关闭 |
| `webhook` | Webhook 模式配置，默认使用轮询（polling）模式 |

> 💡 更多配置项详见 [example.yaml](example.yaml)。

### 3. 运行

```bash
# 直接运行（轮询模式）
go run .

# 或编译后运行
go build -o bandori-tg .
./bandori-tg
```

也可以直接使用命令行参数：

```bash
./bandori-tg run       # 运行机器人（默认）
./bandori-tg version   # 查看版本
./bandori-tg help      # 查看帮助
```

机器人启动后会：

- 自动创建 `logs/` 与 `res/` 目录，日志写入 `logs/app.log`（自动滚动、压缩）。
- 首次启动时并发拉取 Bestdori 数据（乐队、卡片、角色、活动、技能、卡池、资讯）。
- 使用轮询模式接收更新；若配置了 Webhook 则使用 Webhook 模式。
- 收到 `SIGHUP` 信号时热重载配置。

### 4. 在 Telegram 中使用

1. 私聊机器人发送 `/start`。
2. 发送 `/lang` 设置界面语言与查询语言（查询语言决定展示哪个服务区的数据）。
3. 发送 `/setadmin <admin_token>` 将当前账号认证为管理员（需要与 `config.yaml` 中一致），再发送 `/setcommands` 一键注册各语言命令菜单。

## 📖 指令列表

| 指令 | 权限 | 说明 |
| --- | --- | --- |
| `/start` | 所有人 | 开始使用；支持 `events_<id>_<lang>` 参数直达活动详情 |
| `/help` | 所有人 | 查看帮助 |
| `/cards <关键词>` | 所有人 | 查卡（也支持直接发送「查卡 <关键词>」） |
| `/events` | 所有人 | 查看各服近期活动 |
| `/fsx <活动ID>` | 所有人 | 查看活动档线预测图（将在 v1.1 版本推出） |
| `/lang` | 所有人 | 语言设置（界面语言 / 查询语言） |
| `/dlang` | 所有人 | 设置界面语言 |
| `/qlang` | 所有人 | 设置查询语言（数据服务区） |
| `/about` | 所有人 | 查看机器人版本信息 |
| `/setadmin <token>` | 需管理员令牌 | 认证为管理员 |
| `/setcommands` | 管理员 | 为所有支持的语言注册命令菜单 |

## 📡 运行模式

### 轮询模式（默认）

不配置 `webhook` 即可，机器人通过长轮询接收更新，适合大多数部署场景。

### Webhook 模式

在 `config.yaml` 中开启：

```yaml
webhook:
  enabled: true
  nginx_enabled: true   # 使用 Nginx 反向代理时设为 true
  url: "https://yourdomain.com/webhook"
  port: 8443
  secret: "YOUR_WEBHOOK_SECRET"
```

支持自定义证书与 Nginx 反向代理。注意 Telegram 只接受 `80 / 88 / 443 / 8443` 端口，如使用其他端口需自行反向代理。

## 🔄 数据更新

机器人内置定时任务，每日多次拉取 Bestdori 数据，时间点对齐各服活动开始 / 结束前约 30 分钟，保证档线预测与活动信息及时准确。若拉取失败会自动重试（最多 5 次，指数退避）。

## 🛠️ 开发与构建

### 本地构建

```bash
go build ./...
```

### 发布

发布流程由 [GoReleaser](https://goreleaser.com) 与 GitHub Actions 自动完成：

- **稳定版**：推送 `v1*` 标签触发（`.github/workflows/release.yml`）
- **开发版**：推送 `dev` 分支触发，自动发布为 `v0.0.0-dev` 预发布（`.github/workflows/go.yml`）

产物支持 Linux / Windows / macOS 的 amd64 与 arm64 架构。

版本信息在构建时通过 `-ldflags` 注入（`version.Version` / `BuildTime` / `GitCommit` / `Branch`），可通过 `/about` 指令或 `bandori-tg version` 查看。

## 开发路线

- [ ] 添加更多语言支持
- [ ] 优化用户界面
- [ ] 增强数据处理能力
- [ ] 提高系统稳定性

## 📄 许可证

[MIT](LICENSE) © libost


> [!IMPORTANT]
> 本项目与 **BanG Dream! 少女乐团派对！** 官方无关，`Bandori`、`BanG Dream!` 均为其所有者的商标，数据来源于 [Bestdori](https://bestdori.com)。