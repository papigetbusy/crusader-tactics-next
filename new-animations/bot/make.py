#!/usr/bin/env python3
"""First-pass attack GIFs. Then stop.

Drop sheets in new-animations/inbox/. Run this. It writes a review GIF for
each sheet and a note that says what it used. It does not write code, it
does not sign anything off, and it does not touch the game.

The look is the one signed on Grand Hex (2026-10-12):
- the body frames from the sheet, in order, then back to the first frame
- a game effect, recolored to that skill
- the inside of the aura is black and fades out to the skill color
- the color changes a little on every frame
- the spell lands on the enemy and does not cover the attacker
- the last frames are idle, with the glow gone

After it prints STOP, a person reviews each GIF. Code is written only
when the owner says that one is the one.
"""

import json
import re
import shutil
import subprocess
import sys
import tempfile
from pathlib import Path

import numpy as np
from PIL import Image, ImageDraw, ImageFilter, ImageFont

ROOT = Path(__file__).resolve().parents[2]
INBOX = ROOT / "new-animations" / "inbox"
GIFS = ROOT / "new-animations" / "gifs"
EFFECTS = ROOT / "web" / "assets" / "effects"
CHECKLIST = ROOT / "new-animations" / "CHECKLIST.md"
FXBOOK = ROOT / "web" / "game" / "js" / "fxbook.js"
SKILLS = ROOT / "data" / "skills.json"

# Magic-looking skills get the warp ring around the caster.
# A sword swing does not. Both still get the color shift and the idle ending.
RING = {
    "orb", "debuff", "nova", "buff", "cloud", "swarm", "pillar", "heal",
    "beam", "meteor", "blizzard", "rain", "shards", "potion", "revive",
    "chain", "bolt", "lance", "summon",
}


def slug(name):
    return re.sub(r"[^a-z0-9]+", "-", name.lower()).strip("-")


def load_skills():
    data = json.loads(SKILLS.read_text())
    return {s["name"]: s for s in data["skills"]}


def load_fx():
    text = FXBOOK.read_text()
    found = {}
    for quoted, bare, base, colors in re.findall(
        r"""(?:['"]([^'"]+)['"]|([A-Za-z_][A-Za-z0-9_]*))\s*:\s*\['([a-z]+)',\s*'([^']+)'""",
        text,
    ):
        found[quoted or bare] = (base, colors.split(","))
    return found


def approved(skill):
    if not CHECKLIST.exists():
        return False
    for line in CHECKLIST.read_text().splitlines():
        if skill not in line:
            continue
        if line.strip().startswith("- [x]") or "approved" in line:
            return True
    return False


def note_checklist(skill):
    """Say a review exists. Never mark a row done."""
    if not CHECKLIST.exists():
        return "no checklist"
    lines = CHECKLIST.read_text().splitlines()
    changed = False
    for i, line in enumerate(lines):
        if skill not in line or not line.strip().startswith("- [ ]"):
            continue
        if "gif: none" in line:
            lines[i] = line.replace("gif: none", "gif: bot review made, not signed off")
            changed = True
            break
        if "not signed off" in line or "bot review" in line:
            return "already a review"
    if not changed:
        return "no matching row"
    CHECKLIST.write_text("\n".join(lines) + "\n")
    return "checklist noted, still not signed off"


def cut_sheet(path):
    src = np.array(Image.open(path).convert("RGB"))
    bg = src[0, 0].astype(int)
    dist = np.abs(src.astype(int) - bg).sum(2)
    r, g, b = src[:, :, 0], src[:, :, 1], src[:, :, 2]
    magenta = (r > 160) & (b > 120) & (g < 140) & (r > g + 30)
    fg = (dist > 80) & ~magenta
    h, w = fg.shape
    seen = np.zeros(fg.shape, bool)
    parts = []
    ys, xs = np.where(fg)
    from collections import deque
    for y, x in zip(ys, xs):
        if seen[y, x]:
            continue
        q = deque([(y, x)])
        seen[y, x] = True
        pts = [(y, x)]
        while q:
            cy, cx = q.popleft()
            for dy in (-1, 0, 1):
                for dx in (-1, 0, 1):
                    ny, nx = cy + dy, cx + dx
                    if ny < 0 or nx < 0 or ny >= h or nx >= w or seen[ny, nx] or not fg[ny, nx]:
                        continue
                    seen[ny, nx] = True
                    q.append((ny, nx))
                    pts.append((ny, nx))
        if len(pts) < 400:
            continue
        yy, xx = zip(*pts)
        y0, x0, y1, x1 = min(yy), min(xx), max(yy), max(xx)
        crop = src[y0:y1 + 1, x0:x1 + 1]
        d = np.abs(crop.astype(int) - bg).sum(2)
        rr, gg, bb = crop[:, :, 0], crop[:, :, 1], crop[:, :, 2]
        mag = (rr > 160) & (bb > 120) & (gg < 140) & (rr > gg + 30)
        alpha = np.where((d > 70) & ~mag, 255, 0).astype(np.uint8)
        pix = crop[alpha > 0]
        # A hero uses hundreds of colors. A spell is mostly one color,
        # even when the pose is wide (a point, a lean).
        colors = len(np.unique(pix.reshape(-1, 3) // 16, axis=0)) if len(pix) else 0
        im = Image.fromarray(np.dstack([crop, alpha]))
        parts.append((y0, x0, len(pts), colors, im))
    if not parts:
        return [], []
    # Relative, not a fixed number. The hex queen has more colors than
    # these smaller warriors, and a slash has far fewer than either.
    cut = max(80, max(p[3] for p in parts) * 0.45)
    bodies, effects = [], []
    for y0, x0, n, colors, im in parts:
        (bodies if colors >= cut else effects).append((y0, x0, n, im))
    bodies.sort(key=lambda t: (t[0] // 80, t[1]))
    # A row of different sizes is a spell growing. Same-size pictures stay in reading order.
    areas = [t[2] for t in effects]
    if len(areas) >= 2 and max(areas) > min(areas) * 1.6:
        effects.sort(key=lambda t: t[2])
    else:
        effects.sort(key=lambda t: (t[0] // 80, t[1]))
    return [t[3] for t in bodies], [t[3] for t in effects]


def hex_rgb(value):
    value = value.strip().lstrip("#")
    return tuple(int(value[i:i + 2], 16) for i in (0, 2, 4))


def shades(colors):
    base = [hex_rgb(c) for c in colors[:3]] or [(80, 220, 90)]
    while len(base) < 2:
        base.append(tuple(min(255, c + 70) for c in base[0]))
    a, b = base[0], base[1]
    light = tuple(min(255, c + 90) for c in b)
    dark = tuple(max(0, c // 3) for c in a)
    return [dark, light, a, b, dark, light]


def tint_ring(frame, color):
    a = np.array(frame.convert("RGBA")).astype(np.float32)
    lum = 0.299 * a[:, :, 0] + 0.587 * a[:, :, 1] + 0.114 * a[:, :, 2]
    t = np.clip((lum - 20) / 140, 0, 1)
    hot = np.clip((lum - 160) / 50, 0, 1)
    out = np.zeros_like(a)
    out[:, :, 0] = color[0] * t + 50 * hot
    out[:, :, 1] = color[1] * t + 20 * hot
    out[:, :, 2] = color[2] * t
    out[:, :, 3] = a[:, :, 3]
    yy, xx = np.mgrid[0:a.shape[0], 0:a.shape[1]]
    dist = np.hypot(xx - a.shape[1] / 2, yy - a.shape[0] / 2)
    out[:, :, 3] *= np.clip((dist - 14) / 8, 0, 1)
    return Image.fromarray(np.clip(out, 0, 255).astype(np.uint8))


def scale(img, s):
    return img.resize((max(1, int(img.width * s)), max(1, int(img.height * s))), Image.Resampling.NEAREST)


def foot(img):
    aa = np.array(img.split()[-1])
    ys, xs = np.where(aa > 20)
    if len(xs) == 0:
        return img.width / 2, img.height
    return (float(xs.min() + xs.max()) / 2, float(ys.max()))


def paint_board():
    bw, bh = 680, 460
    base = Image.new("RGBA", (bw, bh), (6, 8, 12, 255))
    tile = Image.open(ROOT / "web/assets/ground/grasslands.png").convert("RGBA").crop((128, 64, 192, 128))
    tiles = [(170 + (c - r) * 78, 210 + (c + r) * 36) for r in range(3) for c in range(3)]
    for i in sorted(range(9), key=lambda n: tiles[n][1]):
        cx, cy = tiles[i]
        w, h = 72, 36
        mask = Image.new("L", (bw, bh), 0)
        ImageDraw.Draw(mask).polygon([(cx, cy - h), (cx + w, cy), (cx, cy + h), (cx - w, cy)], fill=255)
        tex = Image.new("RGBA", (bw, bh), (0, 0, 0, 0))
        tex.paste(tile.resize((w * 2, h * 2), Image.Resampling.NEAREST), (cx - w, cy - h))
        base.paste(tex, (0, 0), mask)
        ImageDraw.Draw(base).line([(cx, cy - h), (cx + w, cy), (cx, cy + h), (cx - w, cy), (cx, cy - h)], fill=(18, 36, 22))
    return base, tiles


def black_core(center, power, size):
    layer = Image.new("RGBA", size, (0, 0, 0, 0))
    d = ImageDraw.Draw(layer)
    cx, cy = center
    d.ellipse([cx - 52, cy - 58, cx + 52, cy + 48], fill=(0, 0, 0, int(230 * power)))
    return layer.filter(ImageFilter.GaussianBlur(7))


def render(name, bodies, effects, colors, use_ring):
    # Keep every hero about the same size on the map, whatever the sheet was.
    s = 120 / bodies[0].height
    bodies = [scale(b, s) for b in bodies]
    effects = [scale(e, s) for e in effects]
    palette = shades(colors)
    sheet = Image.open(EFFECTS / "scifi-warp-001.png")
    aura = [scale(tint_ring(sheet.crop((i * 64, 0, (i + 1) * 64, 64)), palette[n % len(palette)]), 3.5) for n, i in enumerate((5, 6, 7, 8, 7, 6))]
    enemy = scale(Image.open(ROOT / "web/assets/monsters/skeleton.png").convert("RGBA"), 1.45)
    font = ImageFont.truetype("/usr/share/fonts/truetype/dejavu/DejaVuSans-Bold.ttf", 16)
    board, tiles = paint_board()
    caster_at, enemy_at = tiles[0], tiles[2]
    # Play the bodies in sheet order, hold the last cast, then idle.
    order = list(range(len(bodies)))
    if len(order) > 1:
        order = order + [order[-1]]
    order += [0, 0, 0]
    frames = []
    for i, bi in enumerate(order):
        glow = 0 if i >= len(order) - 3 else min(1, 0.4 + i * 0.2)
        base = board.copy()
        body = bodies[bi]
        wf = (caster_at[0], caster_at[1] + 20)
        ef = (enemy_at[0], enemy_at[1] + 6)
        bf = foot(body)
        chest = (wf[0], wf[1] - body.height * 0.48)
        color = palette[i % len(palette)]
        if glow > 0 and use_ring:
            ringfx = aura[i % len(aura)]
            base.alpha_composite(ringfx, (int(chest[0] - ringfx.width / 2), int(chest[1] - ringfx.height / 2)))
            base.alpha_composite(black_core(chest, glow, base.size))
        elif glow > 0:
            pad = 22
            sil = Image.new("RGBA", (body.width + pad * 2, body.height + pad * 2), (0, 0, 0, 0))
            sil.paste(Image.new("RGBA", body.size, color + (255,)), (pad, pad), body.split()[-1])
            arr = np.array(sil.filter(ImageFilter.GaussianBlur(6)))
            arr[:, :, 3] = (arr[:, :, 3].astype(np.float32) * glow).astype(np.uint8)
            base.alpha_composite(Image.fromarray(arr), (int(wf[0] - bf[0] - pad), int(wf[1] - bf[1] - pad)))
        base.alpha_composite(body, (int(wf[0] - bf[0]), int(wf[1] - bf[1])))
        base.alpha_composite(enemy, (int(ef[0] - foot(enemy)[0]), int(ef[1] - foot(enemy)[1])))
        if effects and glow > 0 and i < len(order) - 3:
            t = min(1, max(0, (i - 1) / max(1, len(order) - 5)))
            spell = effects[min(len(effects) - 1, int(t * len(effects)))]
            start = (wf[0] + 52, wf[1] - 78)
            hit = (ef[0], ef[1] - 18)
            x = start[0] + (hit[0] - start[0]) * t
            y = start[1] + (hit[1] - start[1]) * t
            if t >= 1:
                ring = 48 + (i % 3) * 28
                ov = Image.new("RGBA", base.size, (0, 0, 0, 0))
                d = ImageDraw.Draw(ov)
                d.ellipse([hit[0] - ring, hit[1] - ring, hit[0] + ring, hit[1] + ring], fill=(0, 0, 0, 90))
                d.ellipse([hit[0] - ring, hit[1] - ring, hit[0] + ring, hit[1] + ring], outline=color + (230,), width=3)
                base.alpha_composite(ov)
                x, y = hit
            base.alpha_composite(spell, (int(x - spell.width / 2), int(y - spell.height / 2)))
        ImageDraw.Draw(base).text((12, 8), name, fill=color + (255,), font=font)
        frames.append(base.convert("RGB"))
    return frames


def write_gif(frames, path):
    path.parent.mkdir(parents=True, exist_ok=True)
    tmp = Path(tempfile.mkdtemp())
    try:
        for n, fr in enumerate(frames):
            fr.save(tmp / f"f{n:03d}.png")
        subprocess.check_call(["ffmpeg", "-y", "-loglevel", "error", "-framerate", "10", "-i", str(tmp / "f%03d.png"), "-vf", "palettegen=max_colors=256:stats_mode=full:reserve_transparent=0", str(tmp / "p.png")])
        subprocess.check_call(["ffmpeg", "-y", "-loglevel", "error", "-framerate", "10", "-i", str(tmp / "f%03d.png"), "-i", str(tmp / "p.png"), "-lavfi", "paletteuse=dither=none", str(path)])
    finally:
        shutil.rmtree(tmp)


def title_from(path, skills):
    stem = path.stem.replace("-", " ").replace("_", " ")
    for name in skills:
        if slug(name) == slug(stem):
            return name
    return stem.title()


def main():
    skills = load_skills()
    fx = load_fx()
    sheets = sorted(p for p in INBOX.glob("*.png") if p.name.lower() != "readme.png")
    if not sheets:
        print("Nothing in new-animations/inbox. Drop the sheets there and run this again.")
        return 0
    report = ["# Bot stop", "", "Review GIFs only. Not signed off. No code was written. The game was not changed.", ""]
    for path in sheets:
        name = title_from(path, skills)
        if approved(name):
            report.append(f"- {name}: skipped. It is already signed off.")
            print("skip", name)
            continue
        info = skills.get(name, {})
        base, colors = fx.get(name, ("debuff", ["#6fe04a", "#c8ff9a"]))
        hero = info.get("class", "unknown")
        bodies, effects = cut_sheet(path)
        if not bodies:
            report.append(f"- {name}: no character found on the sheet. Not drawn.")
            print("no body", name)
            continue
        frames = render(name, bodies, effects, colors, base in RING)
        dest = GIFS / slug(hero) / slug(name)
        gif = dest / "review.gif"
        if gif.exists():
            n = 2
            while (dest / f"review-{n}.gif").exists():
                n += 1
            gif = dest / f"review-{n}.gif"
        write_gif(frames, gif)
        note = dest / "bot-note.md"
        note.write_text(
            f"# {name}\n\n"
            f"Bot review only. Not signed off. No code.\n\n"
            f"- Class: {hero}\n"
            f"- Ascension: {info.get('ascension', 'any form')}\n"
            f"- Effect in the codex: {base} {' '.join(colors)}\n"
            f"- Body frames used, in reading order: {len(bodies)}\n"
            f"- Spell frames used: {len(effects)}\n"
            f"- Labels were not read. If a frame is in the wrong place, say so.\n"
            f"- The last three frames are the first body again, with the glow off.\n"
        )
        listed = note_checklist(name)
        report.append(f"- {name}: {gif.relative_to(ROOT)} ({listed})")
        print("wrote", gif.relative_to(ROOT))
    report.append("")
    report.append("STOP. Go through these one by one with the owner. Do not write code yet.")
    out = ROOT / "new-animations" / "bot" / "last-run.md"
    out.write_text("\n".join(report) + "\n")
    print("\n".join(report))
    return 0


if __name__ == "__main__":
    sys.exit(main())
