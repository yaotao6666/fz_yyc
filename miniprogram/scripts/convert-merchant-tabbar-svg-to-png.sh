#!/usr/bin/env bash

set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
project_root="$(cd "$script_dir/../" && pwd)"

svg_dir="$project_root/src/static/brand-icons/xunmeng-private-butler/tabbar/merchant"
png_dir="$project_root/src/static/tabbar"
size="${1:-81}"

if ! command -v sips >/dev/null 2>&1; then
  echo "未找到 sips，请在 macOS 下运行。" >&2
  exit 1
fi

if [ ! -d "$svg_dir" ]; then
  echo "SVG 目录不存在：$svg_dir" >&2
  exit 1
fi

mkdir -p "$png_dir"

names=(home home-active analytics analytics-active order order-active settings settings-active)

for name in "${names[@]}"; do
  in_file="$svg_dir/$name.svg"
  out_file="$png_dir/$name.png"

  if [ ! -f "$in_file" ]; then
    echo "缺少 SVG：$in_file" >&2
    exit 1
  fi

  sips -s format png -Z "$size" "$in_file" --out "$out_file" >/dev/null

  if [ ! -f "$out_file" ]; then
    echo "生成失败：$out_file" >&2
    exit 1
  fi

  echo "已生成：$out_file"
done

echo "完成：已输出到 ${png_dir}（尺寸 ${size}x${size}）"
