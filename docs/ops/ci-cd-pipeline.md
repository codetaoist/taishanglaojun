# CI/CD 流水线（GitLab CI）

以 GitLab CI 为例，覆盖 API 服务与插件的构建、测试、扫描、打包、部署与回滚；门禁以覆盖率、秘密扫描与SAST为核心。[0]

## 总览
- 触发：PR/MR 创建与更新、tag 发布、主分支合入。
- 阶段：`lint` → `test` → `build` → `package` → `deploy` → `notify`。
- 环境：`dev`/`test`/`prod` 多环境部署；环境变量与密钥集中管理。

## 插件流水线
- `package`：插件打包生成 artifact（含 manifest 与签名）。
- `deploy`：部署到目标环境并注册到后端；回传部署日志与状态。
- 失败处理：停止发布，触发回滚；保留构建日志供审计。

## 服务流水线
- 构建镜像并推送到镜像仓库；部署到 K8s 或主机环境（按部署基线）。
- 结合蓝绿/金丝雀发布策略；监控健康检查与自动回滚。

## 集成点
- 管理后台操作触发后端执行 → 后端调用 CI/CD 网关接口。
- CI/CD 将结果回调后端 → 后端写入审计日志并通知前端。

## 典型变量
- `ENV`（环境）、`VERSION`（版本）、`PLUGIN_ID`（插件ID）、`ARTIFACT_URL`（构建产物地址）。

## 阶段与门禁
- 阶段：`setup` → `build` → `test` → `scan` → `package` → `deploy` → `e2e` → `cleanup`
- 门禁：
  - 覆盖率阈值：单元>70%、契约>60%、E2E>40%
  - 秘密扫描与依赖漏洞扫描（SAST/DAST/SCA）
  - OpenAPI 契约变更需评审并打版本标记

## 示例 .gitlab-ci.yml 片段
```yaml
stages: [setup, build, test, scan, package, deploy, e2e, cleanup]

variables:
  GO_VERSION: "1.21"
  NODE_VERSION: "20"

setup:api:
  stage: setup
  image: golang:${GO_VERSION}
  script:
    - go version
    - go env

build:api:
  stage: build
  image: golang:${GO_VERSION}
  script:
    - cd services/api
    - go build ./...
  artifacts:
    paths:
      - services/api/bin/

test:api:
  stage: test
  image: golang:${GO_VERSION}
  script:
    - cd services/api
    - go test ./... -coverprofile=coverage.out
  artifacts:
    reports:
      coverage_report:
        coverage_format: go
        path: services/api/coverage.out

build:admin:
  stage: build
  image: node:${NODE_VERSION}
  script:
    - cd apps/admin-react
    - pnpm install
    - pnpm build
  artifacts:
    paths:
      - apps/admin-react/dist/

scan:deps:
  stage: scan
  image: node:${NODE_VERSION}
  script:
    - npm audit --audit-level=moderate || true

package:plugin:
  stage: package
  image: node:${NODE_VERSION}
  script:
    - cd packages/plugin-sdk
    - pnpm build
    - pnpm run sign    # 生成清单与签名
  artifacts:
    paths:
      - packages/plugin-sdk/dist/

deploy:api:
  stage: deploy
  script:
    - ./scripts/deploy_api.sh
  environment:
    name: production
    url: https://api.example.com
  when: manual

deploy:admin:
  stage: deploy
  script:
    - ./scripts/deploy_admin.sh
  environment:
    name: production
    url: https://admin.example.com
  when: manual

e2e:admin:
  stage: e2e
  image: mcr.microsoft.com/playwright:v1.47
  script:
    - cd apps/admin-react
    - pnpm test:e2e

cleanup:
  stage: cleanup
  script:
    - echo "Cleanup artifacts and temporary resources"
```

## 策略与实践
- 环境矩阵：dev/staging/prod，差异化变量与密钥管理；发布窗口与灰度比例。
- 回滚：失败自动回滚与人工确认；产物版本化与签名校验。
- 报告与质量门禁：覆盖率、契约一致性、性能与安全扫描报告纳入审查。
- 插件流水线与后端流水线衔接：插件打包与签名校验后，再触发部署；审计落库。

> 参考：[0] https://www.doubao.com/thread/w1745a17f59b91183

## 契约门禁示例（Contracts Gate）
- 目标：对 `openapi/*.yaml` 执行契约探测与回归，生成可审报告。
- 工具：`schemathesis`（基于属性）、`pytest`/`allure` 报告。
- 示例（GitHub Actions）：

```yaml
name: contracts-gate
on: [push, pull_request]
jobs:
  schemathesis:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-python@v5
        with:
          python-version: '3.11'
      - name: Install deps
        run: |
          pip install schemathesis pytest allure-pytest
      - name: Run contract tests (laojun)
        run: |
          schemathesis run openapi/laojun.yaml --checks all --report report_la.html || true
      - name: Run contract tests (taishang)
        run: |
          schemathesis run openapi/taishang.yaml --checks all --report report_tai.html || true
      - name: Collect reports
        run: |
          mkdir -p reports/contracts
          mv *.html reports/contracts/
```

- 产出：`reports/contracts/report_la.html`，`reports/contracts/report_tai.html`；门禁规则：关键接口失败数 ≤ 阈值、错误码映射一致性通过。

## 端侧门禁示例（Native Gate）
- 目标：分端侧执行单测/集成/UI与性能基线，阻断不合格变更。
- 工具：iOS `XCTest`/`XCUITest`；Android `JUnit`/`Espresso`/`Macrobenchmark`；桌面 `pytest`/`Playwright`。
- 示例（矩阵化）：

```yaml
name: native-gate
on: [push]
jobs:
  ios-tests:
    runs-on: macos-latest
    steps:
      - uses: actions/checkout@v4
      - name: Build & Test
        run: xcodebuild -scheme App -destination 'platform=iOS Simulator,name=iPhone 15' test
  android-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Setup Android
        uses: android-actions/setup-android@v3
      - name: Unit & UI
        run: ./gradlew test connectedAndroidTest
  desktop-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Python tests
        run: |
          pip install -r requirements.txt
          pytest -q --maxfail=1 --disable-warnings
```

- 产出：`reports/native/ios/*.xml`，`reports/native/android/*.xml`，`reports/native/desktop/*.xml`；门禁规则：关键用例 100% 通过，性能基线不退化（阈值表见 `docs/testing/strategy.md`）。

## 插件门禁示例（Plugins Gate）
- 目标：校验插件清单、签名、权限与资源约束，执行生命周期烟测。
- 脚本示例：

```bash
#!/usr/bin/env bash
set -e
mkdir -p reports/plugins
python scripts/plugin_check.py --manifest plugins/example/manifest.yaml \
  --public-key keys/pub.pem --out reports/plugins/manifest_report.json
python scripts/plugin_lifecycle_test.py --plugin plugins/example --out reports/plugins/lifecycle.xml
```

- CI 步骤：运行清单校验 → 签名验证 → 权限位与资源配额检查 → 生命周期 `Init/Start/Stop/Health` 烟测。
- 产出：`reports/plugins/manifest_report.json`，`reports/plugins/lifecycle.xml`；门禁：来源可信、权限最小化、生命周期关键步骤通过。

## 产出报告清单（CI Artifacts）
- 契约：`reports/contracts/*.html`，错误码一致性报告。
- 端侧：`reports/native/*/*.xml`，性能与稳定性比对报告。
- 插件：`reports/plugins/*.json|*.xml`，签名与清单校验日志。
- 质量指标：覆盖率、静态扫描结果、依赖与供应链安全报告。

## 多模块流水线矩阵（GitLab CI）
- 目标：按服务维度并行构建与测试，契约作为统一门禁；支持生成客户端与报告产物。
- 阶段：`contracts` → `build` → `test` → `scan` → `package` → `deploy` → `notify`
- 作业矩阵（示例片段）：
```yaml
stages: [contracts, build, test, scan, package, deploy, notify]

variables:
  PY_IMAGE: "python:3.11-slim"
  GO_IMAGE: "golang:1.21"
  NODE_IMAGE: "node:20"

contracts:validate:
  stage: contracts
  image: ${PY_IMAGE}
  script:
    - pip install -r scripts/requirements.txt
    - python scripts/openapi_validate.py openapi/laojun.yaml
    - python scripts/openapi_validate.py openapi/taishang.yaml

contracts:diff:
  stage: contracts
  image: ${PY_IMAGE}
  script:
    - python scripts/openapi_contract_diff.py openapi/laojun.skeleton.yaml openapi/laojun.yaml || true
    - python scripts/openapi_contract_diff.py openapi/taishang.skeleton.yaml openapi/taishang.yaml || true

build:gateway:
  stage: build
  image: ${GO_IMAGE}
  script:
    - cd services/gateway && go build ./...

build:laojun-api:
  stage: build
  image: ${GO_IMAGE}
  script:
    - cd services/laojun-api && go build ./...

build:taishang-api:
  stage: build
  image: ${GO_IMAGE}
  script:
    - cd services/taishang-api && go build ./...

build:admin-react:
  stage: build
  image: ${NODE_IMAGE}
  script:
    - cd apps/admin-react && pnpm install && pnpm build

test:contracts:
  stage: test
  image: ${PY_IMAGE}
  script:
    - pip install schemathesis pytest
    - schemathesis run openapi/laojun.yaml --checks all || true
    - schemathesis run openapi/taishang.yaml --checks all || true
```
- 门禁规则：关键端点失败数 ≤ 阈值，错误码映射一致；构建/测试/扫描均需通过。
- 产物与报告：`reports/contracts/*`、构建产物与签名、覆盖率与安全扫描报告。