# Thunder Lance v2: the body uses the owner's mockup poses (poses/, cut off the
# magenta background); the effects are made in code from the game's pictures:
# the Storm Dragoon sprite, a skeleton, a ground texture and the game's
# lightning / charge-up effect sheets (recolored). No new art.
import math
import os
import random
import sys

import numpy as np
from PIL import Image, ImageDraw, ImageFilter, ImageFont

R = os.path.join(os.path.dirname(os.path.abspath(__file__)), '../../../../web/assets/')  # the game's pictures
OUT = sys.argv[1] if len(sys.argv) > 1 else 'out'  # run: python3 make-v2.py /some/folder
os.makedirs(OUT + '/frames', exist_ok=True)

W, H = 760, 600
TW, TH = 128, 64                     # iso tile
OX, OY = 170, 320                    # screen spot of tile (0, 0)
YEL, CYA, WHT = (255, 228, 92), (127, 227, 240), (255, 255, 255)


def iso(i, j):
    return OX + (i - j) * TW // 2, OY + (i + j) * TH // 2


def load(p, scale, flip=False):
    im = Image.open(R + p).convert('RGBA')
    im = im.resize((int(im.width * scale), int(im.height * scale)), Image.NEAREST)
    return im.transpose(Image.FLIP_LEFT_RIGHT) if flip else im


HERO = load('heroes/dragon-knight-storm-dragoon.png', 1.25, flip=True)   # (v1 body, kept for the glow size)
POSE_DIR = os.path.join(os.path.dirname(os.path.abspath(__file__)), 'poses')  # cut from the owner's mockup


def pose(name, scale):
    im = Image.open(os.path.join(POSE_DIR, name)).convert('RGBA')
    return im.resize((int(im.width * scale), int(im.height * scale)), Image.LANCZOS)


# the owner's mockup poses (2026-10-17): stand, charge (lance up, sparking), wind-up, thrust
POSES = {'stand': pose('pose-stand.png', .92), 'charge': pose('pose-charge.png', .92),
         'windup': pose('pose-windup.png', .86), 'thrust': pose('pose-thrust.png', .86)}


def anchor(im):
    """Where the dragon's feet are: the middle of the bottom rows."""
    a = np.asarray(im)[..., 3] > 100
    rows = a[int(a.shape[0] * .85):]
    xs = np.where(rows.any(0))[0]
    return (xs[0] + xs[-1]) / 2


def lance_tip(im):
    """The thrust pose's lance tip: the rightmost solid pixel."""
    a = np.asarray(im)[..., 3] > 120
    ys, xs = np.where(a)
    k = np.argmax(xs)
    return xs[k], ys[k]
FOE = load('monsters/skeleton.png', 0.55)
FOE = FOE.transpose(Image.FLIP_LEFT_RIGHT)                               # faces the hero


def sheet(name, fw, fh):
    im = Image.open(R + 'effects/' + name).convert('RGBA')
    return [im.crop((x, y, x + fw, y + fh)) for y in range(0, im.height, fh) for x in range(0, im.width, fw)
            if im.crop((x, y, x + fw, y + fh)).getbbox()]


def recolor(im, ramp=((40, 70, 120), CYA, YEL, WHT)):
    """Lightness to the skill's colors (dark blue, cyan, yellow, white), alpha kept."""
    a = np.asarray(im, float)
    lum = (a[..., 0] * .3 + a[..., 1] * .59 + a[..., 2] * .11) / 255
    cols = np.array(ramp, float)
    t = np.clip(lum, 0, 1) * (len(cols) - 1)
    i = np.minimum(t.astype(int), len(cols) - 2)
    f = (t - i)[..., None]
    rgb = cols[i] * (1 - f) + cols[i + 1] * f
    out = np.dstack([rgb, a[..., 3]])
    return Image.fromarray(out.astype(np.uint8), 'RGBA')


CHARGE = [recolor(f).resize((96, 96), Image.NEAREST) for f in sheet('scifi-charge-up-001.png', 48, 48)]
BURST = [recolor(f).resize((120, 120), Image.NEAREST) for f in sheet('lightning-burst-002.png', 32, 32)]
STRIKE = [recolor(f).resize((128, 128), Image.NEAREST) for f in sheet('lightning-strike-001.png', 64, 64)]

# ground: the game's grass texture cut into iso diamonds
tex = Image.open(R + 'ground/grasslands.png').convert('RGBA').crop((0, 0, 64, 64))
diamond = tex.rotate(45, expand=True, resample=Image.NEAREST).resize((TW, TH), Image.NEAREST)
mask = Image.new('L', (TW, TH), 0)
ImageDraw.Draw(mask).polygon([(TW / 2, 0), (TW, TH / 2), (TW / 2, TH), (0, TH / 2)], fill=255)
diamond.putalpha(mask)

TILES = [(i, j) for i in range(-1, 7) for j in range(-1, 2)]
HERO_T, FOES_T, END_T = (0, 0), [(2, 0), (4, 0)], (5, 0)   # target 2 away, the bolt pierces 3 tiles past it


def board():
    im = Image.new('RGBA', (W, H), (7, 8, 12, 255))
    for i, j in sorted(TILES, key=lambda t: t[0] + t[1]):
        x, y = iso(i, j)
        shade = Image.new('RGBA', (TW, TH), (0, 0, 0, 0))
        shade.paste(diamond, (0, 0))
        if (i + j) % 2:
            shade = Image.blend(shade, Image.new('RGBA', (TW, TH), (0, 0, 0, 255)), .08)
            shade.putalpha(mask)
        im.alpha_composite(shade, (x - TW // 2, y - TH // 2))
        ImageDraw.Draw(im).line([(x - TW // 2, y), (x, y - TH // 2), (x + TW // 2, y)], fill=(20, 30, 24, 160), width=1)
    return im


BASE = board()


PAD = 24


def glow_of(sprite, color, strength):
    """A soft colored glow hugging a sprite's outline (padded so it never clips)."""
    a = Image.new('L', (sprite.width + PAD * 2, sprite.height + PAD * 2), 0)
    a.paste(sprite.split()[3], (PAD, PAD))
    a = a.filter(ImageFilter.MaxFilter(5)).filter(ImageFilter.GaussianBlur(5))
    g = Image.new('RGBA', a.size, color + (0,))
    g.putalpha(a.point(lambda v: int(min(255, v * 1.6) * strength)))
    return g


def tint(sprite, color, k):
    t = Image.new('RGBA', sprite.size, color + (255,))
    out = Image.blend(sprite, t, k)
    out.putalpha(sprite.split()[3])
    return out


def put(frame, sprite, tile, dx=0, dy=0, ax=None):
    x, y = iso(*tile)
    ax = sprite.width / 2 if ax is None else ax
    frame.alpha_composite(sprite, (int(x - ax + dx), int(y - sprite.height + 14 + dy)))


def tip(dx=0, dy=0):
    """Where the lance tip is on screen (the sprite is flipped, the tip is top right)."""
    x, y = iso(*HERO_T)
    return x - HERO.width / 2 + HERO.width * 0.72 + dx, y - HERO.height + 14 + 6 + dy


def bolt(draw_on, a, b, seed, width=1.0):
    rnd = random.Random(seed)
    n = 16
    (ax, ay), (bx, by) = a, b
    nx, ny = -(by - ay), bx - ax
    L = math.hypot(nx, ny)
    nx, ny = nx / L, ny / L
    pts = [a]
    for k in range(1, n):
        t = k / n
        off = rnd.uniform(-14, 14) * math.sin(math.pi * t)
        pts.append((ax + (bx - ax) * t + nx * off, ay + (by - ay) * t + ny * off))
    pts.append(b)
    layer = Image.new('RGBA', (W, H), (0, 0, 0, 0))
    d = ImageDraw.Draw(layer)
    d.line(pts, fill=CYA + (230,), width=int(16 * width), joint='curve')
    layer = layer.filter(ImageFilter.GaussianBlur(7))
    d = ImageDraw.Draw(layer)
    d.line(pts, fill=YEL + (255,), width=int(7 * width), joint='curve')
    d.line(pts, fill=WHT + (255,), width=max(1, int(3 * width)), joint='curve')
    for k in range(3, n - 1, 4):                      # small forks
        px, py = pts[k]
        q = (px + rnd.uniform(-30, 30), py + rnd.uniform(-34, -8))
        d.line([(px, py), q], fill=YEL + (220,), width=2)
    draw_on.alpha_composite(layer)
    return pts


def ring(frame, c, r, alpha):
    """The signed look's afterglow: a black center fading out to the skill's color."""
    s = int(r * 2 + 8)
    yy, xx = np.mgrid[0:s, 0:s] - s / 2
    d = np.hypot(xx, yy * 2) / r                     # squashed onto the iso ground
    col = np.zeros((s, s, 4))
    t = np.clip(d, 0, 1)[..., None]
    col[..., :3] = np.array((0, 0, 0)) * (1 - t) + np.array(YEL) * t
    col[..., 3] = np.where(d <= 1, (0.25 + 0.75 * d) * 255 * alpha, 0)
    im = Image.fromarray(col.astype(np.uint8), 'RGBA').filter(ImageFilter.GaussianBlur(2))
    frame.alpha_composite(im, (int(c[0] - s / 2), int(c[1] - s / 2)))


font = ImageFont.truetype('/usr/share/fonts/truetype/dejavu/DejaVuSans-Bold.ttf', 15)
big = ImageFont.truetype('/usr/share/fonts/truetype/dejavu/DejaVuSans-Bold.ttf', 26)


def frame_at(k):
    f = BASE.copy()
    # timing (frames of 50 ms)
    charge = 3 <= k < 11
    lunge = {11: 6, 12: 12, 13: 12, 14: 12, 15: 10, 16: 8, 17: 6, 18: 4, 19: 3, 20: 2, 21: 1}.get(k, 0)
    beam = 12 <= k < 20
    after = 20 <= k < 27
    shake = (k % 2 * 2 - 1) * 2 if 12 <= k < 16 else 0
    hx, hy = lunge, lunge / 2
    # glow strength over time: builds while charging, peaks with the beam, gone at the end
    g = 0 if k < 3 else min(1, (k - 2) / 8) if k < 12 else 1 if k < 20 else max(0, 1 - (k - 19) / 8)
    # enemies (drawn by depth: the far one first)
    for n, t in enumerate(sorted(FOES_T, key=lambda t: t[0] + t[1])):
        sp = FOE
        if beam and k < 16:
            sp = tint(FOE, WHT if k % 2 == 0 else YEL, .65)
        elif beam or (after and k < 23):
            sp = tint(FOE, CYA, .25)
        sx = ((k + n) % 2 * 2 - 1) * 3 if 12 <= k < 18 else 0
        put(f, sp, t, sx + shake, 0)
    # hero: the owner's poses. stand, charge (lance raised), wind-up, thrust, back to stand
    name = 'stand'
    if 3 <= k < 10:
        name = 'charge'
    elif k == 10:
        name = 'windup'
    elif 11 <= k < 21:
        name = 'thrust'
    elif 21 <= k < 24:
        name = 'windup'
    hero = POSES[name]
    ax = anchor(hero)
    if charge and name == 'charge':
        sq = 1 - 0.025 * math.sin((k - 3) / 7 * math.pi)
        hero = hero.resize((hero.width, int(hero.height * sq)), Image.LANCZOS)
    if 11 <= k < 13:
        hero = tint(hero, WHT, .35)
    if g > 0:
        put(f, glow_of(hero, YEL if k % 2 else CYA, .9 * g), HERO_T, hx + shake, hy + PAD, ax + PAD)
    put(f, hero, HERO_T, hx + shake, hy, ax)
    hx0, hy0 = iso(*HERO_T)
    top = hy0 - hero.height + 14 + hy

    def tip(dx=0, dy=0):
        t = POSES['thrust']
        px, py = lance_tip(t)
        return hx0 - anchor(t) + px + dx, hy0 - t.height + 14 + py + dy
    # the bolt: from the tip, through the target, 3 tiles past it
    if beam:
        a = tip(hx + shake, hy)
        ex, ey = iso(*END_T)
        bolt(f, a, (ex, ey - 46), seed=k * 7 + 1, width=1.0 if k < 18 else 0.6)
        for n, t in enumerate(FOES_T):
            x, y = iso(*t)
            b = BURST[min(len(BURST) - 1, (k - 12 + n) % len(BURST))]
            f.alpha_composite(b, (int(x - 60 + shake), int(y - 110)))
        x, y = iso(*END_T)
        s = STRIKE[min(len(STRIKE) - 1, k - 12)]
        f.alpha_composite(s, (int(x - 64), int(y - 118)))
    if after or (beam and k >= 16):
        al = 1 if k < 21 else max(0, 1 - (k - 20) / 6)
        for t in FOES_T:
            ring(f, iso(*t), 44 + (k - 16) * 2, al)
    # damage numbers rising over each enemy
    if 13 <= k < 26:
        d = ImageDraw.Draw(f)
        for n, t in enumerate(FOES_T):
            x, y = iso(*t)
            up = (k - 13) * 3
            txt = ['312', '298'][n]
            d.text((x - 20, y - 190 - up), txt, font=big, fill=YEL, stroke_width=3, stroke_fill=(30, 20, 0))
    # screen flash on the hit
    if k in (12, 13):
        f = Image.blend(f, Image.new('RGBA', (W, H), WHT + (255,)), .12 if k == 12 else .05)
    d = ImageDraw.Draw(f)
    d.text((14, 10), 'Thunder Lance · Storm Dragoon', font=font, fill=(255, 228, 92))
    d.text((14, 30), 'v2 · body poses from the owner’s mockup · effects made in code from the game’s own', font=font, fill=(150, 160, 175))
    return f.convert('RGB')


frames = [frame_at(k) for k in range(32)]  # 50 ms each, the last one held
durs = [50] * 31 + [700]
for k, fr in enumerate(frames):
    fr.save(f'{OUT}/frames/{k:02d}.png')
pal = [fr.quantize(colors=255, method=Image.Quantize.MEDIANCUT, dither=Image.Dither.NONE) for fr in frames]
pal[0].save(f'{OUT}/in-game.gif', save_all=True, append_images=pal[1:], duration=durs, loop=0, disposal=1)
print('ok', len(frames))
