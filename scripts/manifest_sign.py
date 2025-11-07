#!/usr/bin/env python3
import json, sys, argparse, hmac, hashlib
from rich import print

# 最小可运行示例：使用 HMAC-SHA256 对清单 checksum 进行签名
# 生产建议：RSA/Ed25519 非对称签名，平台侧维护公钥白名单

def canonical_manifest(m: dict) -> dict:
    # 计算 checksum 前剔除动态字段
    m2 = dict(m)
    m2.pop("signature", None)
    return m2


def compute_checksum(m: dict) -> str:
    s = json.dumps(m, ensure_ascii=False, separators=(",", ":"))
    digest = hashlib.sha256(s.encode("utf-8")).hexdigest()
    return f"sha256:{digest}"


def sign_checksum(checksum: str, secret: str) -> str:
    mac = hmac.new(secret.encode("utf-8"), checksum.encode("utf-8"), hashlib.sha256)
    return mac.hexdigest()


def main():
    ap = argparse.ArgumentParser(description="Sign plugin manifest.json (HMAC-SHA256)")
    ap.add_argument("--manifest", required=True, help="path to manifest.json")
    ap.add_argument("--secret", required=True, help="HMAC secret")
    ap.add_argument("--out", required=True, help="output signed manifest path")
    args = ap.parse_args()

    with open(args.manifest, "r", encoding="utf-8") as f:
        m = json.load(f)
    m2 = canonical_manifest(m)
    checksum = compute_checksum(m2)
    signature = sign_checksum(checksum, args.secret)
    m_signed = dict(m2)
    m_signed["checksum"] = checksum
    m_signed["signature"] = signature

    with open(args.out, "w", encoding="utf-8") as f:
        json.dump(m_signed, f, ensure_ascii=False, indent=2)
    print(f"[green]OK[/green]: 生成签名文件 -> {args.out}")
    print(f"checksum: {checksum}")
    print(f"signature(HMAC): {signature}")

if __name__ == "__main__":
    main()