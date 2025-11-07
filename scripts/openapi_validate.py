#!/usr/bin/env python3
import argparse, sys
from rich import print
try:
    import yaml
except Exception as e:
    print("[red]缺少 PyYAML，请先安装依赖：pip install -r scripts/requirements.txt[/red]")
    sys.exit(1)


def basic_checks(spec: dict) -> list:
    issues = []
    for key in ["openapi", "info", "paths", "components"]:
        if key not in spec:
            issues.append(f"缺少关键字段: {key}")
    # 安全契约
    sec = spec.get("components", {}).get("securitySchemes", {})
    if "bearerAuth" not in sec:
        issues.append("缺少 bearerAuth 安全方案")
    return issues


def full_validate(spec: dict) -> list:
    try:
        from openapi_spec_validator import validate_spec
        validate_spec(spec)
        return []
    except Exception as e:
        return [f"openapi-spec-validator 校验失败: {e}"]


def main():
    ap = argparse.ArgumentParser(description="Validate OpenAPI spec")
    ap.add_argument("--spec", required=True, help="path to openapi yaml/json")
    ap.add_argument("--full", action="store_true", help="run full validation if possible")
    args = ap.parse_args()

    with open(args.spec, "r", encoding="utf-8") as f:
        if args.spec.endswith(".json"):
            import json
            spec = json.load(f)
        else:
            spec = yaml.safe_load(f)

    issues = basic_checks(spec)
    if args.full:
        issues += full_validate(spec)

    if not issues:
        print(f"[green]OK[/green]: {args.spec} 通过校验")
        sys.exit(0)
    else:
        print(f"[red]FAIL[/red]: {args.spec} 存在问题")
        for i in issues:
            print(f"- {i}")
        sys.exit(2)

if __name__ == "__main__":
    main()