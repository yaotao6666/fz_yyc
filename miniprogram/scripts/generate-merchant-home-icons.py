#!/usr/bin/env python3
"""生成商家首页 5 个正式图标资源。

输出文件：
- src/static/icons/store.png
- src/static/icons/qrcode.png
- src/static/icons/category.png
- src/static/icons/product.png
- src/static/icons/order.png

说明：
- 统一采用透明背景，适配现有首页的 `image` 图标容器。
- 先在 4 倍画布上绘制，再缩放到 81x81，减少小尺寸锯齿。
"""

from __future__ import annotations

from pathlib import Path

from PIL import Image, ImageDraw, ImageFilter


EXPORT_SIZE = 81
SCALE = 4
CANVAS_SIZE = EXPORT_SIZE * SCALE

MAIN_BLUE = "#1D4ED8"
DEEP_BLUE = "#1E3A8A"
ACCENT_GOLD = "#F59E0B"
SOFT_BLUE = "#60A5FA"
WHITE = "#FFFFFF"


def scaled(value: int) -> int:
    """当前坐标已按 4 倍画布标定，这里仅保留统一入口便于后续微调。"""
    return value


def create_canvas() -> tuple[Image.Image, ImageDraw.ImageDraw]:
    image = Image.new("RGBA", (CANVAS_SIZE, CANVAS_SIZE), (0, 0, 0, 0))
    return image, ImageDraw.Draw(image)


def add_soft_shadow(image: Image.Image, alpha: int = 52) -> Image.Image:
    """为图标增加轻微阴影，让透明背景图标在浅色卡片里更清晰。"""
    shadow = Image.new("RGBA", image.size, (0, 0, 0, 0))
    shadow_draw = ImageDraw.Draw(shadow)
    inset = scaled(18)
    shadow_draw.rounded_rectangle(
        (inset, inset, CANVAS_SIZE - inset, CANVAS_SIZE - inset),
        radius=scaled(16),
        fill=(20, 32, 60, alpha),
    )
    shadow = shadow.filter(ImageFilter.GaussianBlur(radius=scaled(6)))
    return Image.alpha_composite(shadow, image)


def resize_and_save(image: Image.Image, output_path: Path) -> None:
    output_path.parent.mkdir(parents=True, exist_ok=True)
    final_image = image.resize((EXPORT_SIZE, EXPORT_SIZE), Image.Resampling.LANCZOS)
    final_image.save(output_path)


def draw_store_icon() -> Image.Image:
    image, draw = create_canvas()

    awning_top = scaled(76)
    awning_bottom = scaled(120)
    draw.rounded_rectangle(
        (scaled(56), awning_top, scaled(268), awning_bottom),
        radius=scaled(18),
        fill=ACCENT_GOLD,
    )

    stripe_width = scaled(34)
    for index in range(5):
        left = scaled(56) + index * stripe_width
        draw.rounded_rectangle(
            (left, awning_top, left + scaled(18), awning_bottom),
            radius=scaled(9),
            fill=(255, 255, 255, 150),
        )

    draw.rounded_rectangle(
        (scaled(70), scaled(118), scaled(254), scaled(250)),
        radius=scaled(26),
        fill=MAIN_BLUE,
    )
    draw.rounded_rectangle(
        (scaled(98), scaled(150), scaled(160), scaled(250)),
        radius=scaled(18),
        fill=WHITE,
    )
    draw.rounded_rectangle(
        (scaled(182), scaled(154), scaled(228), scaled(198)),
        radius=scaled(12),
        outline=WHITE,
        width=scaled(10),
    )
    draw.ellipse(
        (scaled(224), scaled(48), scaled(278), scaled(102)),
        fill=DEEP_BLUE,
    )
    draw.ellipse(
        (scaled(238), scaled(62), scaled(264), scaled(88)),
        fill=WHITE,
    )

    return add_soft_shadow(image, alpha=28)


def draw_qrcode_icon() -> Image.Image:
    image, draw = create_canvas()

    draw.rounded_rectangle(
        (scaled(54), scaled(54), scaled(270), scaled(270)),
        radius=scaled(36),
        outline=MAIN_BLUE,
        width=scaled(18),
    )

    def finder(origin_x: int, origin_y: int) -> None:
        draw.rounded_rectangle(
            (origin_x, origin_y, origin_x + scaled(64), origin_y + scaled(64)),
            radius=scaled(14),
            outline=DEEP_BLUE,
            width=scaled(14),
        )
        draw.rounded_rectangle(
            (origin_x + scaled(18), origin_y + scaled(18), origin_x + scaled(46), origin_y + scaled(46)),
            radius=scaled(8),
            fill=ACCENT_GOLD,
        )

    finder(scaled(76), scaled(76))
    finder(scaled(184), scaled(76))
    finder(scaled(76), scaled(184))

    modules = [
        (184, 184, 22, 22),
        (214, 184, 18, 18),
        (184, 214, 18, 18),
        (212, 214, 22, 22),
        (238, 214, 12, 12),
        (236, 188, 14, 14),
        (214, 240, 18, 18),
    ]
    for x, y, w, h in modules:
        draw.rounded_rectangle(
            (scaled(x), scaled(y), scaled(x + w), scaled(y + h)),
            radius=scaled(5),
            fill=MAIN_BLUE,
        )

    return add_soft_shadow(image, alpha=24)


def draw_category_icon() -> Image.Image:
    image, draw = create_canvas()

    tiles = [
        (70, 72, MAIN_BLUE),
        (174, 72, ACCENT_GOLD),
        (70, 176, SOFT_BLUE),
        (174, 176, DEEP_BLUE),
    ]
    for left, top, color in tiles:
        draw.rounded_rectangle(
            (scaled(left), scaled(top), scaled(left + 80), scaled(top + 80)),
            radius=scaled(24),
            fill=color,
        )

    draw.rounded_rectangle(
        (scaled(90), scaled(94), scaled(130), scaled(134)),
        radius=scaled(12),
        fill=WHITE,
    )
    draw.rounded_rectangle(
        (scaled(194), scaled(94), scaled(234), scaled(106)),
        radius=scaled(6),
        fill=WHITE,
    )
    draw.rounded_rectangle(
        (scaled(194), scaled(116), scaled(234), scaled(128)),
        radius=scaled(6),
        fill=WHITE,
    )
    draw.ellipse(
        (scaled(92), scaled(198), scaled(128), scaled(234)),
        fill=WHITE,
    )
    draw.polygon(
        [
            (scaled(194), scaled(228)),
            (scaled(212), scaled(198)),
            (scaled(230), scaled(228)),
            (scaled(212), scaled(248)),
        ],
        fill=WHITE,
    )

    return add_soft_shadow(image, alpha=18)


def draw_product_icon() -> Image.Image:
    image, draw = create_canvas()

    draw.rounded_rectangle(
        (scaled(78), scaled(112), scaled(246), scaled(252)),
        radius=scaled(28),
        fill=MAIN_BLUE,
    )
    draw.arc(
        (scaled(114), scaled(72), scaled(210), scaled(160)),
        start=200,
        end=340,
        fill=DEEP_BLUE,
        width=scaled(14),
    )
    draw.rounded_rectangle(
        (scaled(120), scaled(146), scaled(206), scaled(178)),
        radius=scaled(14),
        fill=WHITE,
    )
    draw.rounded_rectangle(
        (scaled(110), scaled(192), scaled(214), scaled(206)),
        radius=scaled(7),
        fill=(255, 255, 255, 210),
    )
    draw.rounded_rectangle(
        (scaled(110), scaled(220), scaled(190), scaled(234)),
        radius=scaled(7),
        fill=(255, 255, 255, 180),
    )
    draw.ellipse(
        (scaled(208), scaled(62), scaled(270), scaled(124)),
        fill=ACCENT_GOLD,
    )
    draw.polygon(
        [
            (scaled(239), scaled(72)),
            (scaled(246), scaled(90)),
            (scaled(264), scaled(97)),
            (scaled(246), scaled(104)),
            (scaled(239), scaled(122)),
            (scaled(232), scaled(104)),
            (scaled(214), scaled(97)),
            (scaled(232), scaled(90)),
        ],
        fill=WHITE,
    )

    return add_soft_shadow(image, alpha=22)


def draw_order_icon() -> Image.Image:
    image, draw = create_canvas()

    receipt = [
        (88, 74),
        (234, 74),
        (234, 246),
        (214, 232),
        (194, 246),
        (174, 232),
        (154, 246),
        (134, 232),
        (114, 246),
        (88, 228),
    ]
    draw.polygon([(scaled(x), scaled(y)) for x, y in receipt], fill=MAIN_BLUE)

    for y in (108, 136, 164):
        draw.rounded_rectangle(
            (scaled(116), scaled(y), scaled(202), scaled(y + 12)),
            radius=scaled(6),
            fill=(255, 255, 255, 210),
        )

    draw.ellipse(
        (scaled(170), scaled(172), scaled(272), scaled(274)),
        fill=ACCENT_GOLD,
    )
    draw.line(
        [(scaled(198), scaled(224)), (scaled(214), scaled(240)), (scaled(244), scaled(206))],
        fill=WHITE,
        width=scaled(14),
        joint="curve",
    )

    return add_soft_shadow(image, alpha=24)


def generate_icons(output_dir: Path) -> None:
    icons = {
        "store.png": draw_store_icon(),
        "qrcode.png": draw_qrcode_icon(),
        "category.png": draw_category_icon(),
        "product.png": draw_product_icon(),
        "order.png": draw_order_icon(),
    }

    for filename, image in icons.items():
        resize_and_save(image, output_dir / filename)
        print(f"已生成 {filename}")


def main() -> None:
    script_dir = Path(__file__).resolve().parent
    icon_dir = script_dir.parent / "src" / "static" / "icons"
    generate_icons(icon_dir)


if __name__ == "__main__":
    main()
