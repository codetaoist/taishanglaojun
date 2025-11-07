#!/usr/bin/env python3
import json, sys, argparse, re
from typing import List
from rich import print
from rich.table import Table
import semver

REQUIRED_FIELDS = ["id", "name", "version", "entry", "permissions"]
ENTRY_REQUIRED = ["type"]
SEMVER_RE = re.compile(r"^\d+\.\d+\.\d+(-[0-9A-Za-z.-]+)?(\+[0-9A-Za-z.-]+)?$")

ERRORS = {
    "LAO2000": "清单结构错误",
    "LAO2001": "缺少必填字段",
    "LAO2002": "版本格式不合法（需语义化版本）",
    "LAO2003": "entry 字段不合法",
    "LAO2004": "permissions 字段不合法",
    "LAO2005": "依赖字段格式不合法",
    "LAO2006": "checksum/signature 不一致或缺失",
}

def load_manifest(path: str):
    with open(path, 'r', encoding='utf-8') as f:
        return json.load(f)


def validate_manifest(m: dict) -> List[str]:
    issues = []
    # 结构与必填
    if not isinstance(m, dict):
        issues.append("LAO2000: manifest 需为 JSON 对象")
        return issues
    for k in REQUIRED_FIELDS:
        if k not in m:
            issues.append(f"LAO2001: 缺少必填字段 `{k}`")
    # 版本
    v = m.get("version")
    try:
        semver.VersionInfo.parse(v)
    except Exception:
        issues.append("LAO2002: 版本需语义化，如 1.2.3")
    # entry
    entry = m.get("entry", {})
    if not isinstance(entry, dict) or any(r not in entry for r in ENTRY_REQUIRED):
        issues.append("LAO2003: `entry` 需为对象并包含 `type` 字段")
    # permissions
    perms = m.get("permissions", [])
    if not isinstance(perms, list) or not all(isinstance(p, str) for p in perms):
        issues.append("LAO2004: `permissions` 需为字符串数组")
    # dependencies（可选）
    deps = m.get("dependencies")
    if deps is not None and not isinstance(deps, dict):
        issues.append("LAO2005: `dependencies` 需为对象，包含 runtime/plugins 子项")
    # checksum/signature（可选，但建议）
    checksum = m.get("checksum")
    signature = m.get("signature")
    if (checksum and not isinstance(checksum, str)) or (signature and not isinstance(signature, str)):
        issues.append("LAO2006: `checksum`/`signature` 需为字符串")
    return issues


def main():
    ap = argparse.ArgumentParser(description="Validate plugin manifest.json (Laojun)")
    ap.add_argument("--manifest", required=True, help="path to manifest.json")
    args = ap.parse_args()

    try:
        m = load_manifest(args.manifest)
    except Exception as e:
        print(f"[red]LAO2000: 读取清单失败: {e}[/red]")
        sys.exit(1)

    issues = validate_manifest(m)
    if not issues:
        print("[green]OK[/green]: manifest 校验通过")
        sys.exit(0)
    else:
        table = Table(title="Manifest 校验问题")
        table.add_column("错误码", style="red")
        table.add_column("说明")
        for item in issues:
            code, msg = item.split(":", 1)
            table.add_row(code, msg.strip())
        print(table)
        sys.exit(2)

if __name__ == "__main__":
    main()