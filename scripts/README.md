# scripts 使用说明（manifest 校验/签名、OpenAPI 校验与合同 diff）

## 安装依赖
```bash
python3 -m venv .venv
source .venv/bin/activate
pip install -r scripts/requirements.txt
```

## 清单校验（laojun 基础域）
```bash
python scripts/manifest_validate.py --manifest path/to/manifest.json
```
- 校验项：结构、必填字段、语义化版本、依赖格式、权限声明、checksum 与 signature 一致性。
- 结果：输出 `OK` 或详细错误列表（含错误码 `LAO2000+`）。

## 清单签名（示例：HMAC-SHA256）
```bash
python scripts/manifest_sign.py --manifest path/to/manifest.json --secret your-secret --out path/to/manifest.signed.json
```
- 流程：计算 `checksum`（SHA256，签名前的清单），使用 `secret` 进行 HMAC-SHA256 生成 `signature`。
- 注意：生产环境建议使用不对称密钥（RSA/Ed25519）；本脚本为最小可运行示例。

## OpenAPI 校验（平台/双域）
```bash
python scripts/openapi_validate.py --spec openapi/laojun.skeleton.yaml
python scripts/openapi_validate.py --spec openapi/taishang.skeleton.yaml
```
- 校验项：基本结构（`openapi/info/paths/components`），可选使用 `openapi-spec-validator` 进行完整规范验证。

## OpenAPI 合同 diff（防止接口漂移）
```bash
python scripts/openapi_contract_diff.py --base openapi/laojun.yaml --candidate openapi/laojun.skeleton.yaml
python scripts/openapi_contract_diff.py --base openapi/taishang.yaml --candidate openapi/taishang.skeleton.yaml
```
- 输出：`paths` 与 `components.schemas` 的新增/删除/变更；标识潜在破坏性变更（删除/字段类型变化）。

> 详见：`docs/plugins/development-manual.md`、`docs/interfaces/standard.md`、`docs/ops/ci-cd-pipeline.md`