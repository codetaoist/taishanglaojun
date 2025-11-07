#!/usr/bin/env python3
import argparse
from rich import print
from rich.table import Table
import yaml

# 对比 paths 与 components.schemas 的新增/删除/变更

def load(path: str):
    with open(path, "r", encoding="utf-8") as f:
        return yaml.safe_load(f)


def diff_dict_keys(a: dict, b: dict):
    a_keys, b_keys = set(a.keys()), set(b.keys())
    added = sorted(list(b_keys - a_keys))
    removed = sorted(list(a_keys - b_keys))
    common = sorted(list(a_keys & b_keys))
    changed = []
    for k in common:
        if a[k] != b[k]:
            changed.append(k)
    return added, removed, changed


def show(title: str, added, removed, changed):
    table = Table(title=title)
    table.add_column("类型")
    table.add_column("项")
    for i in added:
        table.add_row("新增", i)
    for i in removed:
        table.add_row("删除", i)
    for i in changed:
        table.add_row("变更", i)
    print(table)


def main():
    ap = argparse.ArgumentParser(description="OpenAPI 合同差异对比")
    ap.add_argument("--base", required=True, help="基线规范（已发布或主分支）")
    ap.add_argument("--candidate", required=True, help="待对比规范（改动候选）")
    args = ap.parse_args()

    base = load(args.base)
    cand = load(args.candidate)

    # paths
    base_paths = base.get("paths", {})
    cand_paths = cand.get("paths", {})
    pa, pr, pc = diff_dict_keys(base_paths, cand_paths)
    show("paths 差异", pa, pr, pc)

    # schemas
    base_schemas = base.get("components", {}).get("schemas", {})
    cand_schemas = cand.get("components", {}).get("schemas", {})
    sa, sr, sc = diff_dict_keys(base_schemas, cand_schemas)
    show("components.schemas 差异", sa, sr, sc)

    # 破坏性变更提示
    if pr or sr:
        print("[red]警告：存在删除项，可能为破坏性变更！[/red]")

if __name__ == "__main__":
    main()