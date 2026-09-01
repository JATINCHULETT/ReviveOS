'use client';

import { useEffect, useRef } from 'react';
import * as d3 from 'd3';
import gsap from 'gsap';

/* ------------------------------------------------------------------ */
/*  Types                                                              */
/* ------------------------------------------------------------------ */

type Vec3 = { x: number; y: number; z: number };

type Particle = Vec3 & { r: number; alpha?: number };

type Cluster = {
  center: Vec3;
  color: string;
  points: Particle[];
  spin: number;
};

type Edge = [Vec3, Vec3];

type Building = {
  label: string;
  origin: Vec3;
  edges: Edge[];
  spin: number;
  color?: string;
};

type Star = { x: number; y: number; z: number; phase: number; r: number; color?: string };

type Streak = { angle: number; radius: number; length: number; speed: number; life: number; color?: string };

/* ------------------------------------------------------------------ */
/*  Scene Construction (Rich Multi-Color & Geometric Meshes)          */
/* ------------------------------------------------------------------ */

function buildGeometricClusters(): Cluster[] {
  const rng = d3.randomLcg(0.417);
  const rNormal = d3.randomNormal.source(rng);
  const rUniform = d3.randomUniform.source(rng);

  const clusters: Cluster[] = [];

  // 1. DENSE BLUE COBALT SPHERICAL NEBULA (Left-Center)
  {
    const count = 750;
    const points: Particle[] = [];
    for (let i = 0; i < count; i++) {
      const u = rUniform(0, 1)();
      const theta = rUniform(0, Math.PI * 2)();
      const phi = Math.acos(rUniform(-1, 1)());
      const rad = Math.cbrt(u) * 36;
      points.push({
        x: rad * Math.sin(phi) * Math.cos(theta),
        y: rad * Math.sin(phi) * Math.sin(theta),
        z: rad * Math.cos(phi) * 0.85,
        r: rUniform(0.6, 1.8)(),
      });
    }
    clusters.push({
      center: { x: -95, y: -25, z: 25 },
      color: '#3b82f6',
      points,
      spin: 1.1,
    });
  }

  // 2. CYAN / TURQUOISE HEXAGONAL PRISMATIC CRYSTAL (Upper-Right Center)
  {
    const count = 680;
    const points: Particle[] = [];
    const hexRadius = 34;
    const hexHeight = 44;
    for (let i = 0; i < count; i++) {
      // Pick a random hex segment (0..5)
      const seg = Math.floor(rUniform(0, 6)());
      const a1 = (seg * Math.PI) / 3;
      const a2 = ((seg + 1) * Math.PI) / 3;
      const t = rUniform(0, 1)();
      const angle = a1 + (a2 - a1) * t;
      const rRatio = rUniform(0.6, 1.0)();
      const h = rUniform(-hexHeight / 2, hexHeight / 2)();

      points.push({
        x: hexRadius * rRatio * Math.cos(angle),
        y: h,
        z: hexRadius * rRatio * Math.sin(angle),
        r: rUniform(0.5, 1.7)(),
      });
    }
    clusters.push({
      center: { x: 75, y: -70, z: -15 },
      color: '#00f5d4',
      points,
      spin: 0.9,
    });
  }

  // 3. GLOWING WHITE / PLATINUM CORE SPHERE (Bottom-Center)
  {
    const count = 850;
    const points: Particle[] = [];
    for (let i = 0; i < count; i++) {
      const gx = rNormal(0, 22);
      const gy = rNormal(0, 22);
      const gz = rNormal(0, 18);
      points.push({
        x: gx(),
        y: gy(),
        z: gz(),
        r: rUniform(0.5, 1.9)(),
      });
    }
    clusters.push({
      center: { x: -25, y: 105, z: -20 },
      color: '#ffffff',
      points,
      spin: 1.0,
    });
  }

  // 4. METALLIC SILVER TORUS / DONUT RING (Bottom-Right Center)
  {
    const count = 720;
    const points: Particle[] = [];
    const R = 36; // major radius
    const r = 15; // minor tube radius
    for (let i = 0; i < count; i++) {
      const u = rUniform(0, Math.PI * 2)();
      const v = rUniform(0, Math.PI * 2)();
      const tubeR = r * Math.sqrt(rUniform(0.4, 1)());
      // Slanted torus orientation
      const px = (R + tubeR * Math.cos(v)) * Math.cos(u);
      const py = tubeR * Math.sin(v) * 0.7 + (px * 0.25);
      const pz = (R + tubeR * Math.cos(v)) * Math.sin(u);

      points.push({
        x: px,
        y: py,
        z: pz,
        r: rUniform(0.5, 1.5)(),
      });
    }
    clusters.push({
      center: { x: 110, y: 75, z: 20 },
      color: '#cbd5e1',
      points,
      spin: 1.25,
    });
  }

  // 5. WIDE OUTER ORBIT CLUSTERS (Teal, Electric Blue, Violet, Silver across perimeter)
  const outerLayout = [
    { center: { x: -240, y: -180, z: 80 }, color: '#22d3ee', count: 340, spread: 55, spin: 0.75 },
    { center: { x: 260, y: -190, z: -60 }, color: '#818cf8', count: 360, spread: 60, spin: 1.1 },
    { center: { x: -310, y: 170, z: 40 }, color: '#38bdf8', count: 320, spread: 50, spin: 0.85 },
    { center: { x: 300, y: 190, z: -40 }, color: '#94a3b8', count: 380, spread: 65, spin: 0.95 },
    { center: { x: -480, y: -60, z: 100 }, color: '#00f5d4', count: 280, spread: 45, spin: 1.3 },
    { center: { x: 490, y: 80, z: -80 }, color: '#60a5fa', count: 310, spread: 52, spin: 0.7 },
  ];

  for (const c of outerLayout) {
    const gx = rNormal(0, c.spread);
    const gy = rNormal(0, c.spread);
    const gz = rNormal(0, c.spread * 0.8);
    const points: Particle[] = Array.from({ length: c.count }, () => ({
      x: gx(),
      y: gy(),
      z: gz(),
      r: rUniform(0.5, 1.6)(),
    }));
    clusters.push({ center: c.center, color: c.color, points, spin: c.spin });
  }

  return clusters;
}

/** Builds multifaceted wireframe polygon mesh (Network dome, gateway boxes, banking towers) */
function buildNetworkDome(label: string, origin: Vec3, spin: number, color = '#22d3ee'): Building {
  const edges: Edge[] = [];
  const segments = 8;
  const radius1 = 65;
  const radius2 = 45;
  const height1 = 70;
  const height2 = 45;

  const ring1: Vec3[] = [];
  const ring2: Vec3[] = [];
  const topPoint: Vec3 = { x: 0, y: -height1 / 2 - 25, z: 0 };
  const bottomPoint: Vec3 = { x: 0, y: height1 / 2 + 25, z: 0 };

  for (let i = 0; i < segments; i++) {
    const angle = (i / segments) * Math.PI * 2;
    ring1.push({
      x: radius1 * Math.cos(angle),
      y: -height2 / 2,
      z: radius1 * Math.sin(angle),
    });
    ring2.push({
      x: radius2 * Math.cos(angle + Math.PI / segments),
      y: height2 / 2,
      z: radius2 * Math.sin(angle + Math.PI / segments),
    });
  }

  for (let i = 0; i < segments; i++) {
    const next = (i + 1) % segments;
    edges.push([ring1[i], ring1[next]]);
    edges.push([ring2[i], ring2[next]]);
    edges.push([ring1[i], ring2[i]]);
    edges.push([ring1[next], ring2[i]]);
    edges.push([topPoint, ring1[i]]);
    edges.push([bottomPoint, ring2[i]]);
  }

  return { label, origin, edges, spin, color };
}

function buildComplexBuilding(label: string, origin: Vec3, spin: number, color = '#e2e8f0'): Building {
  const boxes: { size: Vec3; offset: Vec3; rot: number }[] = [
    { size: { x: 80, y: 100, z: 70 }, offset: { x: 0, y: 0, z: 0 }, rot: 0.15 },
    { size: { x: 55, y: 70, z: 55 }, offset: { x: 35, y: -45, z: 25 }, rot: 0.45 },
    { size: { x: 45, y: 60, z: 45 }, offset: { x: -30, y: 40, z: -20 }, rot: -0.35 },
    { size: { x: 35, y: 40, z: 35 }, offset: { x: 20, y: 35, z: 40 }, rot: 0.7 },
  ];

  const edges: Edge[] = [];

  for (const box of boxes) {
    const { x: sx, y: sy, z: sz } = box.size;
    const corners: Vec3[] = [];
    for (const dx of [-1, 1]) {
      for (const dy of [-1, 1]) {
        for (const dz of [-1, 1]) {
          const cosr = Math.cos(box.rot);
          const sinr = Math.sin(box.rot);
          const px = (dx * sx) / 2;
          const pz = (dz * sz) / 2;
          corners.push({
            x: px * cosr + pz * sinr + box.offset.x,
            y: (dy * sy) / 2 + box.offset.y,
            z: -px * sinr + pz * cosr + box.offset.z,
          });
        }
      }
    }
    const pairs: [number, number][] = [
      [0, 1], [0, 2], [3, 1], [3, 2],
      [4, 5], [4, 6], [7, 5], [7, 6],
      [0, 4], [1, 5], [2, 6], [3, 7],
    ];
    for (const [a, b] of pairs) edges.push([corners[a], corners[b]]);
  }

  return { label, origin, edges, spin, color };
}

function buildStars(count: number, spreadX: number, spreadY: number): Star[] {
  const rng = d3.randomLcg(0.417);
  const rUniform = d3.randomUniform.source(rng);
  const colors = ['#ffffff', '#e0f2fe', '#bae6fd', '#7dd3fc', '#a7f3d0', '#c7d2fe'];

  return Array.from({ length: count }, () => ({
    x: rUniform(-spreadX, spreadX)(),
    y: rUniform(-spreadY, spreadY)(),
    z: rUniform(-300, 700)(),
    phase: rUniform(0, Math.PI * 2)(),
    r: rUniform(0.4, 1.8)(),
    color: colors[Math.floor(rUniform(0, colors.length)())],
  }));
}

function rotateY(p: Vec3, angle: number): Vec3 {
  const c = Math.cos(angle);
  const s = Math.sin(angle);
  return { x: p.x * c + p.z * s, y: p.y, z: -p.x * s + p.z * c };
}

function project(p: Vec3, cx: number, cy: number, focal: number, camZ: number) {
  const z = p.z + camZ;
  const scale = focal / Math.max(z, 1);
  return { x: cx + p.x * scale, y: cy + p.y * scale, scale };
}

/* ------------------------------------------------------------------ */
/*  Main Component                                                     */
/* ------------------------------------------------------------------ */

export default function PaymentsNetworkBackground() {
  const canvasRef = useRef<HTMLCanvasElement | null>(null);

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;
    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    const rng = d3.randomLcg(0.417);
    const rUniform = d3.randomUniform.source(rng);

    const clusters = buildGeometricClusters();
    const networkDome = buildNetworkDome('NETWORK', { x: -480, y: 10, z: 40 }, 0.6, '#00f5d4');
    const buildingGateway = buildComplexBuilding('PAYMENT GATEWAY', { x: 480, y: -140, z: -30 }, 0.75, '#e2e8f0');
    const buildingIssuer = buildComplexBuilding('ISSUER BANK CORE', { x: 450, y: 220, z: 60 }, 0.85, '#60a5fa');
    const buildingSwitch = buildComplexBuilding('ROUTING SWITCH', { x: -360, y: -220, z: -50 }, 0.5, '#94a3b8');

    const buildings = [networkDome, buildingGateway, buildingIssuer, buildingSwitch];
    const stars = buildStars(480, 1400, 900);

    // Wide concentric orbital rings spanning the whole screen width/height
    const ringRadii = [80, 145, 220, 310, 420, 545, 690, 850, 1040, 1260];

    const streakColors = ['#00f5d4', '#38bdf8', '#ffffff', '#60a5fa', '#a78bfa', '#f8fafc'];
    const streaks: Streak[] = Array.from({ length: 9 }, (_, i) => ({
      angle: (i / 9) * Math.PI * 2 + 0.3,
      radius: 120 + i * 75,
      length: 70 + i * 14,
      speed: 0.14 + i * 0.04,
      life: rUniform(0, 1)(),
      color: streakColors[i % streakColors.length],
    }));

    const intro = {
      rings: 0,
      particles: 0,
      buildings: 0,
    };
    const rotation = { angle: 0, buildingWobble: 0 };
    const pointer = { x: 0, y: 0 };
    const pointerTarget = { x: 0, y: 0 };

    let width = 0;
    let height = 0;
    const dpr = Math.min(window.devicePixelRatio || 1, 2);

    function resize() {
      if (!canvas) return;
      width = canvas.clientWidth;
      height = canvas.clientHeight;
      canvas.width = width * dpr;
      canvas.height = height * dpr;
      ctx!.setTransform(dpr, 0, 0, dpr, 0, 0);
    }
    resize();
    window.addEventListener('resize', resize);

    function onPointerMove(e: PointerEvent) {
      const nx = (e.clientX / window.innerWidth) * 2 - 1;
      const ny = (e.clientY / window.innerHeight) * 2 - 1;
      pointerTarget.x = nx;
      pointerTarget.y = ny;
    }
    window.addEventListener('pointermove', onPointerMove);

    const tl = gsap.timeline({ defaults: { ease: 'power3.out' } });
    tl.to(intro, { rings: 1, duration: 1.6 }, 0.1)
      .to(intro, { buildings: 1, duration: 1.4 }, 0.3)
      .to(intro, { particles: 1, duration: 1.8 }, 0.5);

    gsap.to(rotation, { angle: Math.PI * 2, duration: 110, repeat: -1, ease: 'none' });
    gsap.to(rotation, {
      buildingWobble: Math.PI * 2,
      duration: 70,
      repeat: -1,
      ease: 'none',
    });

    function render() {
      const t = gsap.ticker.time;
      const cx = width / 2;
      const cy = height / 2;
      const focal = 680;
      const camZ = 950;

      pointer.x += (pointerTarget.x - pointer.x) * 0.04;
      pointer.y += (pointerTarget.y - pointer.y) * 0.04;

      const isLight = document.documentElement.classList.contains('light');

      ctx!.clearRect(0, 0, width, height);

      // Deep cinematic radial background with subtle dark teal/blue gradient
      const bgGrad = ctx!.createRadialGradient(
        cx, cy, 60,
        cx, cy, Math.max(width, height) * 0.8
      );
      if (isLight) {
        bgGrad.addColorStop(0, '#f8fafc');
        bgGrad.addColorStop(0.5, '#f1f5f9');
        bgGrad.addColorStop(1, '#e2e8f0');
      } else {
        bgGrad.addColorStop(0, 'rgba(4, 12, 28, 0.95)');
        bgGrad.addColorStop(0.5, 'rgba(2, 6, 18, 0.98)');
        bgGrad.addColorStop(1, '#020408');
      }
      ctx!.fillStyle = bgGrad;
      ctx!.fillRect(0, 0, width, height);

      const camAngle = rotation.angle + pointer.x * 0.18;
      const camTilt = pointer.y * 14;

      // ════════════ 1. AMBIENT STARFIELD & DEPTH PARTICLES ════════════
      ctx!.save();
      for (const s of stars) {
        const twinkle = 0.5 + 0.5 * Math.sin(t * 0.8 + s.phase);
        const p = project({ x: s.x, y: s.y + camTilt * 0.25, z: s.z }, cx, cy, focal, camZ);
        if (p.x < -40 || p.x > width + 40 || p.y < -40 || p.y > height + 40) continue;
        ctx!.globalAlpha = isLight ? (0.2 + twinkle * 0.35) : (0.3 + twinkle * 0.6);
        ctx!.fillStyle = s.color || (isLight ? '#64748b' : '#ffffff');
        const r = s.r * p.scale * 0.95;
        ctx!.beginPath();
        ctx!.arc(p.x, p.y, Math.max(r, 0.4), 0, Math.PI * 2);
        ctx!.fill();
      }
      ctx!.restore();

      // ════════════ 2. ORBITAL RINGS, RADIAL SPOKES & DRIFTING STREAKS ════════════
      ctx!.save();
      ctx!.globalAlpha = intro.rings;
      ctx!.translate(cx, cy - camTilt * 0.4);
      ctx!.rotate(camAngle * 0.05);

      // Concentric circular orbit lines
      for (let i = 0; i < ringRadii.length; i++) {
        const radius = ringRadii[i] * (1 + pointer.x * 0.012);
        ctx!.beginPath();
        ctx!.ellipse(0, 0, radius, radius * 0.98, 0, 0, Math.PI * 2);
        if (isLight) {
          ctx!.strokeStyle = i % 3 === 0
            ? 'rgba(13, 148, 136, 0.32)'
            : i % 2 === 0
            ? 'rgba(79, 70, 229, 0.25)'
            : 'rgba(71, 85, 105, 0.2)';
        } else {
          ctx!.strokeStyle = i % 3 === 0
            ? 'rgba(0, 245, 212, 0.28)'
            : i % 2 === 0
            ? 'rgba(200, 215, 235, 0.22)'
            : 'rgba(96, 165, 250, 0.18)';
        }
        ctx!.lineWidth = i === 2 || i === 4 ? (isLight ? 1.6 : 1.4) : 1;
        ctx!.stroke();
      }

      // Radial spoke grid
      const spokeCount = 24;
      for (let i = 0; i < spokeCount; i++) {
        const a = (i / spokeCount) * Math.PI * 2;
        ctx!.beginPath();
        ctx!.moveTo(Math.cos(a) * ringRadii[0], Math.sin(a) * ringRadii[0] * 0.98);
        ctx!.lineTo(Math.cos(a) * ringRadii[ringRadii.length - 1], Math.sin(a) * ringRadii[ringRadii.length - 1] * 0.98);
        ctx!.strokeStyle = isLight ? 'rgba(71, 85, 105, 0.12)' : 'rgba(148, 163, 184, 0.08)';
        ctx!.lineWidth = 0.8;
        ctx!.stroke();
      }

      // High-speed glowing light streaks along the orbits
      for (const s of streaks) {
        s.life = (s.life + 0.003 * s.speed * 60) % 1;
        const a = s.angle + camAngle * 0.05;
        const r0 = s.radius + s.life * 320;
        const r1 = r0 + s.length;
        const alpha = Math.sin(s.life * Math.PI) * (isLight ? 0.75 : 0.85);
        ctx!.beginPath();
        ctx!.moveTo(Math.cos(a) * r0, Math.sin(a) * r0 * 0.98);
        ctx!.lineTo(Math.cos(a) * r1, Math.sin(a) * r1 * 0.98);
        ctx!.strokeStyle = isLight ? `rgba(2, 132, 199, ${alpha})` : (s.color || `rgba(0, 245, 212, ${alpha})`);
        ctx!.lineWidth = isLight ? 2.2 : 1.8;
        ctx!.shadowColor = isLight ? '#0284c7' : (s.color || '#00f5d4');
        ctx!.shadowBlur = 8;
        ctx!.stroke();
        ctx!.shadowBlur = 0;
      }
      ctx!.restore();

      // ════════════ 3. DENSE GEOMETRIC PARTICLE CLUSTERS (COBALT, CYAN HEX, WHITE, SILVER TORUS) ════════════
      ctx!.save();
      ctx!.globalAlpha = intro.particles;
      for (const cluster of clusters) {
        const rotatedCenter = rotateY(cluster.center, camAngle * 0.35 * cluster.spin);
        let clusterColor = cluster.color;
        if (isLight) {
          if (clusterColor === '#ffffff' || clusterColor === '#eeece5' || clusterColor === '#f8fafc') clusterColor = '#0f172a';
          else if (clusterColor === '#cbd5e1' || clusterColor === '#9aa1ab' || clusterColor === '#94a3b8') clusterColor = '#475569';
          else if (clusterColor === '#00f5d4' || clusterColor === '#63e6c8') clusterColor = '#0d9488';
          else if (clusterColor === '#22d3ee' || clusterColor === '#38bdf8') clusterColor = '#0284c7';
          else if (clusterColor === '#818cf8' || clusterColor === '#7fa8e8') clusterColor = '#4338ca';
        }

        for (const pt of cluster.points) {
          const world: Vec3 = {
            x: rotatedCenter.x + pt.x,
            y: rotatedCenter.y + pt.y + camTilt * 0.35,
            z: rotatedCenter.z + pt.z,
          };
          const rotated = rotateY(world, camAngle * 0.18);
          const p = project(rotated, cx, cy, focal, camZ);
          const depthAlpha = Math.min(1, Math.max(isLight ? 0.3 : 0.18, 750 / (rotated.z + camZ)));
          ctx!.globalAlpha = intro.particles * depthAlpha;
          ctx!.fillStyle = clusterColor;
          ctx!.beginPath();
          ctx!.arc(p.x, p.y, Math.max(pt.r * p.scale * 1.15, 0.6), 0, Math.PI * 2);
          ctx!.fill();
        }
      }
      ctx!.restore();

      // ════════════ 4. MULTI-FACETED 3D WIREFRAME BANKING & NETWORK NODES ════════════
      ctx!.save();
      ctx!.globalAlpha = intro.buildings;
      for (const building of buildings) {
        const sway = Math.sin(rotation.buildingWobble * building.spin) * 0.08;
        let labelAnchor: { x: number; y: number } | null = null;
        let bColor = building.color || '#22d3ee';
        if (isLight) {
          if (bColor === '#00f5d4' || bColor === '#22d3ee') bColor = '#0f766e';
          else if (bColor === '#e2e8f0' || bColor === '#ffffff') bColor = '#1e293b';
          else if (bColor === '#60a5fa' || bColor === '#818cf8') bColor = '#1d4ed8';
          else if (bColor === '#94a3b8') bColor = '#475569';
        }
        ctx!.fillStyle = bColor;

        for (const [a, b] of building.edges) {
          const steps = 15;
          for (let i = 0; i <= steps; i++) {
            const t2 = i / steps;
            const mid: Vec3 = {
              x: a.x + (b.x - a.x) * t2 + building.origin.x,
              y: a.y + (b.y - a.y) * t2 + building.origin.y,
              z: a.z + (b.z - a.z) * t2 + building.origin.z,
            };
            const rotated = rotateY(mid, camAngle * 0.22 + sway);
            const p = project(rotated, cx, cy, focal, camZ);
            const depthAlpha = Math.min(1, Math.max(isLight ? 0.4 : 0.25, 750 / (rotated.z + camZ)));
            ctx!.globalAlpha = intro.buildings * depthAlpha * (isLight ? 1 : 0.95);
            ctx!.beginPath();
            ctx!.arc(p.x, p.y, Math.max((isLight ? 1.1 : 0.95) * p.scale, 0.6), 0, Math.PI * 2);
            ctx!.fill();
            if (i === 0) labelAnchor = { x: p.x, y: p.y };
          }
        }

        if (labelAnchor) {
          ctx!.globalAlpha = intro.buildings * (isLight ? 0.95 : 0.85);
          ctx!.fillStyle = bColor;
          ctx!.font = "700 11px 'JetBrains Mono', monospace";
          ctx!.textAlign = building.origin.x < 0 ? 'left' : 'right';
          const lx = building.origin.x < 0 ? labelAnchor.x - 70 : labelAnchor.x + 70;
          ctx!.fillText(building.label, lx, labelAnchor.y - 75);
        }
      }
      ctx!.restore();
    }

    gsap.ticker.add(render);

    return () => {
      gsap.ticker.remove(render);
      gsap.killTweensOf(intro);
      gsap.killTweensOf(rotation);
      window.removeEventListener('resize', resize);
      window.removeEventListener('pointermove', onPointerMove);
    };
  }, []);

  return (
    <canvas
      ref={canvasRef}
      style={{
        position: 'absolute',
        inset: 0,
        width: '100%',
        height: '100%',
        display: 'block',
        pointerEvents: 'none',
        zIndex: 1,
      }}
    />
  );
}

