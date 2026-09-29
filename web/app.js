// AEGIS DISTRIBUTED RATE LIMITER - ULTRA-PREMIUM INTERACTIVE ENGINE & SIMULATOR

(function() {
  'use strict';

  // --- AUDIO SYNTHESIZER (Web Audio API) ---
  let audioCtx = null;
  let audioEnabled = false;

  function initAudio() {
    if (!audioCtx) {
      const AudioContextClass = window.AudioContext || window.webkitAudioContext;
      if (AudioContextClass) {
        audioCtx = new AudioContextClass();
      }
    }
  }

  function playSynthSound(type) {
    if (!audioEnabled || !audioCtx) return;
    try {
      if (audioCtx.state === 'suspended') {
        audioCtx.resume();
      }
      const osc = audioCtx.createOscillator();
      const gain = audioCtx.createGain();
      osc.connect(gain);
      gain.connect(audioCtx.destination);

      const now = audioCtx.currentTime;

      if (type === 'allowed') {
        // High-tech crisp laser chirp
        osc.type = 'sine';
        osc.frequency.setValueAtTime(880, now);
        osc.frequency.exponentialRampToValueAtTime(1760, now + 0.06);
        gain.gain.setValueAtTime(0.08, now);
        gain.gain.exponentialRampToValueAtTime(0.001, now + 0.07);
        osc.start(now);
        osc.stop(now + 0.08);
      } else if (type === 'blocked') {
        // Low electronic shield deflection buzz
        osc.type = 'sawtooth';
        osc.frequency.setValueAtTime(180, now);
        osc.frequency.exponentialRampToValueAtTime(60, now + 0.12);
        gain.gain.setValueAtTime(0.12, now);
        gain.gain.exponentialRampToValueAtTime(0.001, now + 0.13);
        osc.start(now);
        osc.stop(now + 0.14);
      } else if (type === 'ddos') {
        // Dramatic alarm pulse
        osc.type = 'triangle';
        osc.frequency.setValueAtTime(440, now);
        osc.frequency.linearRampToValueAtTime(220, now + 0.2);
        gain.gain.setValueAtTime(0.15, now);
        gain.gain.exponentialRampToValueAtTime(0.001, now + 0.22);
        osc.start(now);
        osc.stop(now + 0.23);
      }
    } catch (e) {
      // Audio fallback silent
    }
  }

  // --- APPLICATION STATE ---
  const state = {
    key: 'client_edge_alpha',
    algorithm: 'token_bucket',
    tier: 'pro',
    capacity: 100,
    rate: 50,
    currentTokens: 100,
    totalAllowed: 0,
    totalBlocked: 0,
    recentAllowed: 0,
    recentBlocked: 0,
    currentRps: 0,
    currentLatencyUs: 52,
    streamingInterval: null,
    isStreaming: false,
    activeFilter: 'all',
  };

  // --- BACKGROUND PARTICLE MATRIX CANVAS ---
  const bgCanvas = document.getElementById('bg-particle-canvas');
  const bgCtx = bgCanvas ? bgCanvas.getContext('2d') : null;
  let bgParticles = [];

  function resizeBgCanvas() {
    if (!bgCanvas) return;
    bgCanvas.width = window.innerWidth;
    bgCanvas.height = window.innerHeight;
  }
  window.addEventListener('resize', resizeBgCanvas);
  resizeBgCanvas();

  class BgParticle {
    constructor() {
      this.reset();
    }
    reset() {
      this.x = Math.random() * (bgCanvas ? bgCanvas.width : 1000);
      this.y = Math.random() * (bgCanvas ? bgCanvas.height : 800);
      this.vx = (Math.random() - 0.5) * 0.4;
      this.vy = (Math.random() - 0.5) * 0.4;
      this.radius = Math.random() * 2 + 1;
      this.color = Math.random() > 0.6 ? '#00f2fe' : (Math.random() > 0.5 ? '#8b5cf6' : '#10b981');
      this.alpha = Math.random() * 0.4 + 0.1;
    }
    update() {
      this.x += this.vx;
      this.y += this.vy;
      if (this.x < 0 || this.x > bgCanvas.width) this.vx *= -1;
      if (this.y < 0 || this.y > bgCanvas.height) this.vy *= -1;
    }
    draw(ctx) {
      ctx.beginPath();
      ctx.arc(this.x, this.y, this.radius, 0, Math.PI * 2);
      ctx.fillStyle = this.color;
      ctx.globalAlpha = this.alpha;
      ctx.fill();
    }
  }

  function initBgParticles() {
    bgParticles = [];
    const count = Math.min(80, Math.floor(window.innerWidth / 20));
    for (let i = 0; i < count; i++) {
      bgParticles.push(new BgParticle());
    }
  }
  initBgParticles();

  function animateBg() {
    if (bgCtx && bgCanvas) {
      bgCtx.clearRect(0, 0, bgCanvas.width, bgCanvas.height);

      // Draw particle connections
      for (let i = 0; i < bgParticles.length; i++) {
        bgParticles[i].update();
        bgParticles[i].draw(bgCtx);

        for (let j = i + 1; j < bgParticles.length; j++) {
          const dx = bgParticles[i].x - bgParticles[j].x;
          const dy = bgParticles[i].y - bgParticles[j].y;
          const dist = Math.sqrt(dx * dx + dy * dy);
          if (dist < 110) {
            bgCtx.beginPath();
            bgCtx.moveTo(bgParticles[i].x, bgParticles[i].y);
            bgCtx.lineTo(bgParticles[j].x, bgParticles[j].y);
            bgCtx.strokeStyle = '#00f2fe';
            bgCtx.globalAlpha = (1 - dist / 110) * 0.12;
            bgCtx.stroke();
          }
        }
      }
      bgCtx.globalAlpha = 1;
    }
    requestAnimationFrame(animateBg);
  }
  requestAnimationFrame(animateBg);


  // --- TOPOLOGY PACKET PIPELINE SIMULATOR ---
  const topoCanvas = document.getElementById('topo-canvas');
  const topoCtx = topoCanvas ? topoCanvas.getContext('2d') : null;
  let topoPackets = [];

  function resizeTopoCanvas() {
    if (!topoCanvas) return;
    topoCanvas.width = topoCanvas.parentElement.clientWidth;
    topoCanvas.height = topoCanvas.parentElement.clientHeight;
  }
  window.addEventListener('resize', resizeTopoCanvas);
  resizeTopoCanvas();

  class TopoPacket {
    constructor(isAllowed) {
      this.isAllowed = isAllowed;
      this.progress = 0; // 0 to 1
      this.speed = Math.random() * 0.025 + 0.02;
      this.color = isAllowed ? '#10b981' : '#f43f5e';
      this.sparks = [];
      this.deflected = false;
    }
    update() {
      this.progress += this.speed;
      if (!this.isAllowed && this.progress >= 0.45 && !this.deflected) {
        this.deflected = true;
        // spawn deflection sparks
        for (let i = 0; i < 8; i++) {
          this.sparks.push({
            x: 0,
            y: 0,
            vx: (Math.random() - 0.5) * 4,
            vy: (Math.random() - 0.5) * 4,
            alpha: 1
          });
        }
      }
      for (let s of this.sparks) {
        s.x += s.vx;
        s.y += s.vy;
        s.alpha -= 0.06;
      }
    }
  }

  function spawnTopoPacket(isAllowed) {
    if (topoPackets.length < 50) {
      topoPackets.push(new TopoPacket(isAllowed));
    }
  }

  function renderTopology() {
    if (!topoCtx || !topoCanvas) return;
    const w = topoCanvas.width;
    const h = topoCanvas.height;
    topoCtx.clearRect(0, 0, w, h);

    const midY = h / 2;
    const nodes = [
      { x: w * 0.12, label: 'CLIENT INGRESS', color: '#38bdf8' },
      { x: w * 0.45, label: 'AEGIS GATEWAY (L1)', color: '#00f2fe' },
      { x: w * 0.72, label: 'REDIS LUA (L2)', color: '#8b5cf6' },
      { x: w * 0.92, label: 'BACKEND SERVICES', color: '#10b981' }
    ];

    // Draw connection lines
    topoCtx.strokeStyle = 'rgba(255, 255, 255, 0.1)';
    topoCtx.lineWidth = 2;
    topoCtx.beginPath();
    topoCtx.moveTo(nodes[0].x, midY);
    topoCtx.lineTo(nodes[3].x, midY);
    topoCtx.stroke();

    // Draw active circuit pulses
    const pulseOffset = (Date.now() / 20) % 40;
    topoCtx.strokeStyle = 'rgba(0, 242, 254, 0.35)';
    topoCtx.setLineDash([8, 12]);
    topoCtx.lineDashOffset = -pulseOffset;
    topoCtx.beginPath();
    topoCtx.moveTo(nodes[0].x, midY);
    topoCtx.lineTo(nodes[3].x, midY);
    topoCtx.stroke();
    topoCtx.setLineDash([]);

    // Update and draw packets
    for (let i = topoPackets.length - 1; i >= 0; i--) {
      const p = topoPackets[i];
      p.update();

      if (p.deflected) {
        // Draw sparks at gateway
        const gwX = nodes[1].x;
        for (let s of p.sparks) {
          if (s.alpha > 0) {
            topoCtx.fillStyle = `rgba(244, 63, 94, ${s.alpha})`;
            topoCtx.beginPath();
            topoCtx.arc(gwX + s.x, midY + s.y, 2.5, 0, Math.PI * 2);
            topoCtx.fill();
          }
        }
        if (p.sparks.every(s => s.alpha <= 0)) {
          topoPackets.splice(i, 1);
        }
      } else {
        const curX = nodes[0].x + (nodes[3].x - nodes[0].x) * p.progress;
        topoCtx.beginPath();
        topoCtx.arc(curX, midY, 4, 0, Math.PI * 2);
        topoCtx.fillStyle = p.color;
        topoCtx.shadowColor = p.color;
        topoCtx.shadowBlur = 10;
        topoCtx.fill();
        topoCtx.shadowBlur = 0;

        if (p.progress >= 1.0) {
          topoPackets.splice(i, 1);
        }
      }
    }

    // Draw node circles
    for (let n of nodes) {
      topoCtx.beginPath();
      topoCtx.arc(n.x, midY, 10, 0, Math.PI * 2);
      topoCtx.fillStyle = '#0a101d';
      topoCtx.strokeStyle = n.color;
      topoCtx.lineWidth = 2.5;
      topoCtx.fill();
      topoCtx.stroke();

      topoCtx.beginPath();
      topoCtx.arc(n.x, midY, 4, 0, Math.PI * 2);
      topoCtx.fillStyle = n.color;
      topoCtx.fill();

      // Node labels
      topoCtx.font = '9px "JetBrains Mono"';
      topoCtx.fillStyle = '#94a3b8';
      topoCtx.textAlign = 'center';
      topoCtx.fillText(n.label, n.x, midY + 24);
    }

    requestAnimationFrame(renderTopology);
  }
  requestAnimationFrame(renderTopology);


  // --- TOKEN BUCKET 2D PHYSICS CHAMBER ---
  const chamberCanvas = document.getElementById('token-physics-canvas');
  const chamberCtx = chamberCanvas ? chamberCanvas.getContext('2d') : null;
  let tokenParticles = [];

  class TokenOrb {
    constructor(x, y, vx, vy) {
      this.x = x || (chamberCanvas ? chamberCanvas.width / 2 + (Math.random() - 0.5) * 40 : 140);
      this.y = y || 20;
      this.vx = vx || (Math.random() - 0.5) * 3;
      this.vy = vy || Math.random() * 2 + 1;
      this.radius = 7;
      this.color = Math.random() > 0.3 ? '#00f2fe' : '#38bdf8';
      this.settled = false;
      this.draining = false;
      this.alpha = 1;
    }
    update(w, h, fillLevelY) {
      if (this.draining) {
        // Drain towards bottom center portal
        const targetX = w / 2;
        const targetY = h + 20;
        this.x += (targetX - this.x) * 0.15;
        this.y += 6;
        this.alpha -= 0.08;
        return;
      }

      this.vy += 0.25; // gravity
      this.x += this.vx;
      this.y += this.vy;

      // Wall bounce
      if (this.x - this.radius < 12) {
        this.x = 12 + this.radius;
        this.vx *= -0.7;
      }
      if (this.x + this.radius > w - 12) {
        this.x = w - 12 - this.radius;
        this.vx *= -0.7;
      }

      // Rest on fluid surface or bottom
      const floorY = Math.max(fillLevelY, h - 22);
      if (this.y + this.radius > floorY) {
        this.y = floorY - this.radius;
        this.vy *= -0.4;
        this.vx *= 0.85;
        if (Math.abs(this.vy) < 0.3) {
          this.settled = true;
          this.vy = 0;
        }
      }
    }
    draw(ctx) {
      ctx.save();
      ctx.globalAlpha = Math.max(0, this.alpha);
      ctx.beginPath();
      ctx.arc(this.x, this.y, this.radius, 0, Math.PI * 2);
      ctx.fillStyle = this.color;
      ctx.shadowColor = this.color;
      ctx.shadowBlur = 12;
      ctx.fill();

      // Core white sparkle
      ctx.beginPath();
      ctx.arc(this.x - 2, this.y - 2, 2.5, 0, Math.PI * 2);
      ctx.fillStyle = '#ffffff';
      ctx.fill();
      ctx.restore();
    }
  }

  function spawnTokenOrb() {
    if (chamberCanvas && tokenParticles.length < 120) {
      tokenParticles.push(new TokenOrb());
    }
  }

  function drainTokenOrb() {
    // Pick a settled orb to drain out
    const candidates = tokenParticles.filter(p => !p.draining);
    if (candidates.length > 0) {
      const idx = Math.floor(Math.random() * candidates.length);
      candidates[idx].draining = true;
    }
  }

  function animateChamber() {
    if (!chamberCtx || !chamberCanvas) return;
    const w = chamberCanvas.width;
    const h = chamberCanvas.height;
    chamberCtx.clearRect(0, 0, w, h);

    // Compute fluid fill level
    const fillPct = Math.max(0, Math.min(1, state.currentTokens / state.capacity));
    const fillHeight = (h - 60) * fillPct;
    const fillY = h - 20 - fillHeight;

    // Fluid reservoir gradient with animated wave
    const wave = Math.sin(Date.now() / 250) * 4;
    const grad = chamberCtx.createLinearGradient(0, fillY, 0, h);
    grad.addColorStop(0, 'rgba(0, 242, 254, 0.28)');
    grad.addColorStop(1, 'rgba(56, 189, 248, 0.65)');

    chamberCtx.beginPath();
    chamberCtx.moveTo(12, h - 20);
    chamberCtx.lineTo(12, fillY + wave);
    chamberCtx.quadraticCurveTo(w / 2, fillY - wave, w - 12, fillY + wave);
    chamberCtx.lineTo(w - 12, h - 20);
    chamberCtx.closePath();
    chamberCtx.fillStyle = grad;
    chamberCtx.fill();

    // Fluid top glowing surface line
    chamberCtx.beginPath();
    chamberCtx.moveTo(12, fillY + wave);
    chamberCtx.quadraticCurveTo(w / 2, fillY - wave, w - 12, fillY + wave);
    chamberCtx.strokeStyle = 'rgba(255, 255, 255, 0.8)';
    chamberCtx.lineWidth = 2;
    chamberCtx.shadowColor = '#00f2fe';
    chamberCtx.shadowBlur = 10;
    chamberCtx.stroke();
    chamberCtx.shadowBlur = 0;

    // Update and draw token particles
    for (let i = tokenParticles.length - 1; i >= 0; i--) {
      const orb = tokenParticles[i];
      orb.update(w, h, fillY + 15);
      orb.draw(chamberCtx);

      if (orb.alpha <= 0) {
        tokenParticles.splice(i, 1);
      }
    }

    // Keep particles in sync with token count
    const desiredParticles = Math.min(45, Math.ceil(state.currentTokens / 2.5));
    if (tokenParticles.filter(p => !p.draining).length < desiredParticles) {
      spawnTokenOrb();
    }

    requestAnimationFrame(animateChamber);
  }
  requestAnimationFrame(animateChamber);


  // --- SLIDING WINDOW RADAR & TIMELINE CANVAS ---
  const swCanvas = document.getElementById('sw-timeline-canvas');
  const swCtx = swCanvas ? swCanvas.getContext('2d') : null;
  const swHistoryEvents = [];

  function resizeSwCanvas() {
    if (!swCanvas) return;
    swCanvas.width = swCanvas.parentElement.clientWidth;
    swCanvas.height = swCanvas.parentElement.clientHeight;
  }
  window.addEventListener('resize', resizeSwCanvas);
  resizeSwCanvas();

  function recordSwEvent(isAllowed) {
    swHistoryEvents.push({
      time: Date.now(),
      allowed: isAllowed
    });
  }

  function renderSwTimeline() {
    if (!swCtx || !swCanvas) return;
    const w = swCanvas.width;
    const h = swCanvas.height;
    swCtx.clearRect(0, 0, w, h);

    const now = Date.now();
    const windowMs = 1000;
    const totalViewMs = 2500; // past 2.5 seconds

    // Prune events older than 3 seconds
    while (swHistoryEvents.length > 0 && now - swHistoryEvents[0].time > 3000) {
      swHistoryEvents.shift();
    }

    // Draw time grid
    swCtx.strokeStyle = 'rgba(255, 255, 255, 0.05)';
    swCtx.lineWidth = 1;
    for (let i = 0; i <= 5; i++) {
      const gx = (w / 5) * i;
      swCtx.beginPath();
      swCtx.moveTo(gx, 0);
      swCtx.lineTo(gx, h);
      swCtx.stroke();
    }

    // Draw active 1s window bracket (from right edge - 1000ms to right edge)
    const rightPadding = 40;
    const pxPerMs = (w - rightPadding) / totalViewMs;
    const windowWidth = windowMs * pxPerMs;
    const windowRight = w - rightPadding;
    const windowLeft = windowRight - windowWidth;

    swCtx.fillStyle = 'rgba(0, 242, 254, 0.08)';
    swCtx.fillRect(windowLeft, 10, windowWidth, h - 30);
    swCtx.strokeStyle = '#00f2fe';
    swCtx.lineWidth = 2;
    swCtx.strokeRect(windowLeft, 10, windowWidth, h - 30);

    // Label on bracket
    swCtx.font = '10px "JetBrains Mono"';
    swCtx.fillStyle = '#00f2fe';
    swCtx.textAlign = 'left';
    swCtx.fillText('ACTIVE 1,000ms SLICE', windowLeft + 8, 25);

    // Plot request pulses
    let activeInWindow = 0;
    for (let evt of swHistoryEvents) {
      const elapsed = now - evt.time;
      const x = windowRight - elapsed * pxPerMs;

      if (x >= 0 && x <= w) {
        if (elapsed <= windowMs) {
          activeInWindow++;
        }
        const dotY = evt.allowed ? h * 0.45 : h * 0.75;
        const color = evt.allowed ? '#10b981' : '#f43f5e';

        swCtx.beginPath();
        swCtx.arc(x, dotY, 4, 0, Math.PI * 2);
        swCtx.fillStyle = color;
        swCtx.shadowColor = color;
        swCtx.shadowBlur = 8;
        swCtx.fill();
        swCtx.shadowBlur = 0;
      }
    }

    // Update Live Math Breakdown
    const elapsedSlice = now % windowMs;
    const prevWeight = Math.round(((windowMs - elapsedSlice) / windowMs) * 100);
    const swCountVal = document.getElementById('sw-count-val');
    const swLimitVal = document.getElementById('sw-limit-val');
    const swCurrSlice = document.getElementById('sw-curr-slice');
    const swPrevWeight = document.getElementById('sw-prev-weight');
    const swEstimated = document.getElementById('sw-estimated');

    if (swCountVal) swCountVal.textContent = activeInWindow;
    if (swLimitVal) swLimitVal.textContent = state.capacity;
    if (swCurrSlice) swCurrSlice.textContent = `${activeInWindow} reqs`;
    if (swPrevWeight) swPrevWeight.textContent = `${prevWeight}%`;
    if (swEstimated) swEstimated.textContent = `${activeInWindow} reqs`;

    requestAnimationFrame(renderSwTimeline);
  }
  requestAnimationFrame(renderSwTimeline);


  // --- LEAKY BUCKET DRAIN CANVAS ---
  const lbCanvas = document.getElementById('lb-drain-canvas');
  const lbCtx = lbCanvas ? lbCanvas.getContext('2d') : null;
  let lbDrops = [];

  function resizeLbCanvas() {
    if (!lbCanvas) return;
    lbCanvas.width = lbCanvas.parentElement.clientWidth;
    lbCanvas.height = lbCanvas.parentElement.clientHeight;
  }
  window.addEventListener('resize', resizeLbCanvas);
  resizeLbCanvas();

  function renderLeakyBucket() {
    if (!lbCtx || !lbCanvas) return;
    const w = lbCanvas.width;
    const h = lbCanvas.height;
    lbCtx.clearRect(0, 0, w, h);

    const midX = w / 2;

    // Draw funnel
    lbCtx.beginPath();
    lbCtx.moveTo(midX - 70, 20);
    lbCtx.lineTo(midX + 70, 20);
    lbCtx.lineTo(midX + 15, h - 50);
    lbCtx.lineTo(midX + 15, h - 25);
    lbCtx.lineTo(midX - 15, h - 25);
    lbCtx.lineTo(midX - 15, h - 50);
    lbCtx.closePath();
    lbCtx.strokeStyle = 'rgba(0, 242, 254, 0.4)';
    lbCtx.lineWidth = 2;
    lbCtx.fillStyle = 'rgba(0, 242, 254, 0.05)';
    lbCtx.fill();
    lbCtx.stroke();

    // Constant leak drip from spout
    if (Math.random() < 0.25) {
      lbDrops.push({ x: midX, y: h - 22, vy: 3 });
    }

    // Animate drops
    for (let i = lbDrops.length - 1; i >= 0; i--) {
      const d = lbDrops[i];
      d.y += d.vy;
      d.vy += 0.2;

      lbCtx.beginPath();
      lbCtx.arc(d.x, d.y, 3, 0, Math.PI * 2);
      lbCtx.fillStyle = '#10b981';
      lbCtx.shadowColor = '#10b981';
      lbCtx.shadowBlur = 6;
      lbCtx.fill();
      lbCtx.shadowBlur = 0;

      if (d.y > h) {
        lbDrops.splice(i, 1);
      }
    }

    requestAnimationFrame(renderLeakyBucket);
  }
  requestAnimationFrame(renderLeakyBucket);


  // --- OSCILLOSCOPE DUAL STREAM CHART ---
  const oscCanvas = document.getElementById('oscilloscope-canvas');
  const oscCtx = oscCanvas ? oscCanvas.getContext('2d') : null;
  const historyLen = 60;
  const histAllowed = new Array(historyLen).fill(0);
  const histBlocked = new Array(historyLen).fill(0);
  const histLatency = new Array(historyLen).fill(50);

  function resizeOscCanvas() {
    if (!oscCanvas) return;
    oscCanvas.width = oscCanvas.parentElement.clientWidth;
    oscCanvas.height = oscCanvas.parentElement.clientHeight;
  }
  window.addEventListener('resize', resizeOscCanvas);
  resizeOscCanvas();

  function drawOscilloscope() {
    if (!oscCtx || !oscCanvas) return;
    const w = oscCanvas.width;
    const h = oscCanvas.height;
    oscCtx.clearRect(0, 0, w, h);

    // Draw oscilloscope reticle grid
    oscCtx.strokeStyle = 'rgba(255, 255, 255, 0.04)';
    oscCtx.lineWidth = 1;
    for (let y = 1; y <= 4; y++) {
      const gy = (h / 5) * y;
      oscCtx.beginPath();
      oscCtx.moveTo(0, gy);
      oscCtx.lineTo(w, gy);
      oscCtx.stroke();
    }
    for (let x = 1; x <= 8; x++) {
      const gx = (w / 9) * x;
      oscCtx.beginPath();
      oscCtx.moveTo(gx, 0);
      oscCtx.lineTo(gx, h);
      oscCtx.stroke();
    }

    const maxVal = Math.max(15, ...histAllowed, ...histBlocked);

    // Draw Area Curves
    function plotStream(arr, stroke, fill, maxV) {
      oscCtx.beginPath();
      const step = w / (arr.length - 1);
      oscCtx.moveTo(0, h - (arr[0] / maxV) * (h - 25) - 10);
      for (let i = 1; i < arr.length; i++) {
        const px = i * step;
        const py = h - (arr[i] / maxV) * (h - 25) - 10;
        oscCtx.lineTo(px, py);
      }
      oscCtx.strokeStyle = stroke;
      oscCtx.lineWidth = 2.2;
      oscCtx.shadowColor = stroke;
      oscCtx.shadowBlur = 8;
      oscCtx.stroke();
      oscCtx.shadowBlur = 0;

      // Area fill
      oscCtx.lineTo(w, h);
      oscCtx.lineTo(0, h);
      oscCtx.closePath();
      oscCtx.fillStyle = fill;
      oscCtx.fill();
    }

    plotStream(histAllowed, '#10b981', 'rgba(16, 185, 129, 0.12)', maxVal);
    plotStream(histBlocked, '#f43f5e', 'rgba(244, 63, 94, 0.18)', maxVal);

    requestAnimationFrame(drawOscilloscope);
  }
  requestAnimationFrame(drawOscilloscope);


  // --- DOM CONTROLS & API DISPATCHER ---
  const inputClient = document.getElementById('input-client');
  const btnRollKey = document.getElementById('btn-roll-key');
  const selectAlg = document.getElementById('select-alg');
  const selectTier = document.getElementById('select-tier');
  const sliderCap = document.getElementById('slider-cap');
  const sliderRate = document.getElementById('slider-rate');
  const valCap = document.getElementById('val-cap');
  const valRate = document.getElementById('val-rate');
  const activeAlgTag = document.getElementById('active-alg-tag');
  const canvasTokenCap = document.getElementById('canvas-token-cap');
  const txtRefillRate = document.getElementById('txt-refill-rate');

  // KPI elements
  const kpiRps = document.getElementById('kpi-rps');
  const kpiAllowed = document.getElementById('kpi-allowed');
  const kpiBlocked = document.getElementById('kpi-blocked');
  const kpiAllowedPct = document.getElementById('kpi-allowed-pct');
  const kpiBlockedPct = document.getElementById('kpi-blocked-pct');
  const kpiAllowedSub = document.getElementById('kpi-allowed-sub');
  const kpiLatency = document.getElementById('kpi-latency');
  const kpiLatencyUs = document.getElementById('kpi-latency-us');
  const trackRps = document.getElementById('track-rps');
  const trackAllowed = document.getElementById('track-allowed');
  const trackBlocked = document.getElementById('track-blocked');
  const trackLatency = document.getElementById('track-latency');

  // Canister elements
  const canvasTokenCount = document.getElementById('canvas-token-count');
  const canvasBadge = document.getElementById('canvas-badge');
  const fillReservoir = document.getElementById('fill-reservoir');
  const diagTokensAvail = document.getElementById('diag-tokens-avail');
  const diagTokensPct = document.getElementById('diag-tokens-pct');
  const diagRefillCalc = document.getElementById('diag-refill-calc');
  const diagTimeToFill = document.getElementById('diag-time-to-fill');
  const diagFillState = document.getElementById('diag-fill-state');
  const forcefieldShield = document.getElementById('forcefield-shield');

  // Terminal Stream
  const terminalStream = document.getElementById('terminal-stream');

  // Audio Toggle
  const btnToggleAudio = document.getElementById('btn-toggle-audio');
  const audioIcon = document.getElementById('audio-icon');
  const audioStatus = document.getElementById('audio-status');

  if (btnToggleAudio) {
    btnToggleAudio.addEventListener('click', () => {
      initAudio();
      audioEnabled = !audioEnabled;
      btnToggleAudio.classList.toggle('active', audioEnabled);
      if (audioIcon) audioIcon.textContent = audioEnabled ? '🔊' : '🔇';
      if (audioStatus) audioStatus.textContent = audioEnabled ? 'ARMED' : 'MUTED';
      if (audioEnabled) playSynthSound('allowed');
    });
  }

  // Preset Tiers
  const TIER_CONFIGS = {
    free: { capacity: 20, rate: 10 },
    pro: { capacity: 100, rate: 50 },
    enterprise: { capacity: 500, rate: 250 },
    custom: { capacity: 100, rate: 50 }
  };

  if (selectTier) {
    selectTier.addEventListener('change', () => {
      const cfg = TIER_CONFIGS[selectTier.value];
      if (cfg) {
        sliderCap.value = cfg.capacity;
        sliderRate.value = cfg.rate;
        syncSliderValues();
      }
    });
  }

  function syncSliderValues() {
    state.capacity = parseInt(sliderCap.value, 10);
    state.rate = parseFloat(sliderRate.value);

    if (valCap) valCap.textContent = state.capacity;
    if (valRate) valRate.textContent = state.rate;
    if (canvasTokenCap) canvasTokenCap.textContent = state.capacity;
    if (txtRefillRate) txtRefillRate.textContent = state.rate;
    if (diagRefillCalc) diagRefillCalc.textContent = `+${state.rate.toFixed(2)} / sec`;
  }

  if (sliderCap) sliderCap.addEventListener('input', () => {
    if (selectTier) selectTier.value = 'custom';
    syncSliderValues();
  });
  if (sliderRate) sliderRate.addEventListener('input', () => {
    if (selectTier) selectTier.value = 'custom';
    syncSliderValues();
  });
  syncSliderValues();

  if (btnRollKey) {
    btnRollKey.addEventListener('click', () => {
      const hex = Math.floor(Math.random() * 0xffffff).toString(16).padStart(6, '0');
      inputClient.value = `client_tier_${hex}`;
      state.key = inputClient.value;
    });
  }

  if (selectAlg) {
    selectAlg.addEventListener('change', () => {
      state.algorithm = selectAlg.value;
      const text = selectAlg.options[selectAlg.selectedIndex].text.toUpperCase();
      if (activeAlgTag) activeAlgTag.textContent = text.split(' ')[0] + ' ' + (text.split(' ')[1] || '');
    });
  }

  // HUD Tabs
  const hudTabs = document.querySelectorAll('.hud-tab');
  hudTabs.forEach(tab => {
    tab.addEventListener('click', () => {
      hudTabs.forEach(t => t.classList.remove('active'));
      tab.classList.add('active');
      const targetId = tab.getAttribute('data-target');
      document.querySelectorAll('.algo-tab-content').forEach(c => {
        c.classList.toggle('hidden', c.id !== targetId);
      });
    });
  });

  // --- CORE RATE LIMIT CHECK DISPATCHER ---
  async function dispatchRateLimitCheck(cost = 1) {
    const key = (inputClient ? inputClient.value.trim() : state.key) || 'client_edge_alpha';
    state.key = key;

    const payload = {
      key: key,
      algorithm: state.algorithm,
      cost: cost,
      capacity: state.capacity,
      rate_per_second: state.rate,
      window_size_ms: 1000,
      client_tier: selectTier ? selectTier.value : 'pro'
    };

    const t0 = performance.now();
    try {
      const res = await fetch('/v1/limiter/check', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload)
      });

      const t1 = performance.now();
      const latencyMs = (t1 - t0).toFixed(2);
      const latencyUs = Math.round((t1 - t0) * 1000);
      const data = await res.json();
      const isAllowed = res.status === 200 && data.allowed;

      handleDecisionResult(isAllowed, data, latencyMs, latencyUs, key);
      return data;
    } catch (err) {
      appendTerminalLog('sys', `Network fault: ${err.message}`, key, 0);
      return null;
    }
  }

  function handleDecisionResult(allowed, data, latencyMs, latencyUs, key) {
    state.currentLatencyUs = latencyUs;

    if (allowed) {
      state.totalAllowed++;
      state.recentAllowed++;
      playSynthSound('allowed');
      spawnTopoPacket(true);
      drainTokenOrb();
    } else {
      state.totalBlocked++;
      state.recentBlocked++;
      playSynthSound('blocked');
      spawnTopoPacket(false);
      triggerForcefield();
    }

    recordSwEvent(allowed);

    // Update tokens state
    if (data && data.remaining !== undefined) {
      state.currentTokens = data.remaining;
      const cap = data.limit || state.capacity;
      const fillPct = Math.max(0, Math.min(100, (state.currentTokens / cap) * 100));

      if (canvasTokenCount) canvasTokenCount.textContent = state.currentTokens;
      if (fillReservoir) fillReservoir.style.width = `${fillPct}%`;
      if (diagTokensAvail) diagTokensAvail.textContent = `${state.currentTokens}.0`;
      if (diagTokensPct) diagTokensPct.textContent = `${Math.round(fillPct)}%`;

      if (canvasBadge) {
        if (state.currentTokens === 0) {
          canvasBadge.textContent = 'EXHAUSTED // DROPPING';
          canvasBadge.style.color = '#f43f5e';
          canvasBadge.style.borderColor = '#f43f5e';
        } else {
          canvasBadge.textContent = 'STABLE RESERVOIR';
          canvasBadge.style.color = '#00f2fe';
          canvasBadge.style.borderColor = 'rgba(0, 242, 254, 0.3)';
        }
      }

      if (diagTimeToFill) {
        if (data.retry_after_ms > 0) {
          diagTimeToFill.textContent = `${(data.retry_after_ms / 1000).toFixed(2)} s`;
          if (diagFillState) diagFillState.textContent = 'RECHARGING';
        } else {
          diagTimeToFill.textContent = '0.00 s';
          if (diagFillState) diagFillState.textContent = 'READY';
        }
      }
    }

    // Update Telemetry Ribbon
    if (kpiAllowed) kpiAllowed.textContent = state.totalAllowed.toLocaleString();
    if (kpiBlocked) kpiBlocked.textContent = state.totalBlocked.toLocaleString();
    if (kpiAllowedSub) kpiAllowedSub.textContent = `${state.totalAllowed.toLocaleString()} total`;

    const grandTotal = state.totalAllowed + state.totalBlocked;
    if (grandTotal > 0) {
      const allowedPct = ((state.totalAllowed / grandTotal) * 100).toFixed(1);
      const blockedPct = ((state.totalBlocked / grandTotal) * 100).toFixed(1);
      if (kpiAllowedPct) kpiAllowedPct.textContent = `${allowedPct}%`;
      if (kpiBlockedPct) kpiBlockedPct.textContent = `${blockedPct}%`;
      if (trackAllowed) trackAllowed.style.width = `${allowedPct}%`;
      if (trackBlocked) trackBlocked.style.width = `${blockedPct}%`;
    }

    if (kpiLatency) kpiLatency.textContent = latencyMs;
    if (kpiLatencyUs) kpiLatencyUs.textContent = `${latencyUs} µs`;

    // Append to Terminal Audit Log
    appendTerminalLog(allowed ? '200' : '429', `KEY=${key} REMAINING=${data.remaining}/${data.limit} TIER=${data.cache_tier || 'L1'}`, key, latencyMs);
  }

  function triggerForcefield() {
    if (!forcefieldShield) return;
    forcefieldShield.classList.remove('shield-triggered');
    void forcefieldShield.offsetWidth; // trigger reflow
    forcefieldShield.classList.add('shield-triggered');
    setTimeout(() => {
      forcefieldShield.classList.remove('shield-triggered');
    }, 350);
  }

  function appendTerminalLog(statusType, content, key, latMs) {
    if (!terminalStream) return;
    const row = document.createElement('div');
    row.className = 'terminal-row';

    const now = new Date();
    const ts = now.toTimeString().split(' ')[0] + '.' + String(now.getMilliseconds()).padStart(3, '0');

    let badgeClass = 'badge-200';
    let badgeText = '200 OK';
    if (statusType === '429') {
      badgeClass = 'badge-429';
      badgeText = '429 DROP';
    } else if (statusType === 'sys') {
      badgeClass = 'badge-sys';
      badgeText = 'SYSTEM';
    }

    row.innerHTML = `
      <span class="row-ts">[${ts}]</span>
      <span class="row-badge ${badgeClass}">${badgeText}</span>
      <span class="row-content">${content}</span>
      <span class="row-lat">${latMs}ms</span>
    `;

    terminalStream.prepend(row);
    while (terminalStream.children.length > 50) {
      terminalStream.removeChild(terminalStream.lastChild);
    }
  }

  // --- BUTTON EVENT LISTENERS ---
  const btnFire1 = document.getElementById('btn-fire-1');
  const btnFireBurst = document.getElementById('btn-fire-burst');
  const btnFireStream = document.getElementById('btn-fire-stream');
  const btnFireDdos = document.getElementById('btn-fire-ddos');
  const streamIcon = document.getElementById('stream-icon');
  const streamLabel = document.getElementById('stream-label');
  const streamSub = document.getElementById('stream-sub');
  const btnResetQuota = document.getElementById('btn-reset-quota');
  const btnChaosStorm = document.getElementById('btn-chaos-storm');
  const btnClearLogs = document.getElementById('btn-clear-logs');

  if (btnFire1) {
    btnFire1.addEventListener('click', () => dispatchRateLimitCheck(1));
  }

  if (btnFireBurst) {
    btnFireBurst.addEventListener('click', () => {
      for (let i = 0; i < 25; i++) {
        setTimeout(() => dispatchRateLimitCheck(1), i * 15);
      }
    });
  }

  if (btnFireStream) {
    btnFireStream.addEventListener('click', () => {
      if (state.isStreaming) {
        clearInterval(state.streamingInterval);
        state.isStreaming = false;
        btnFireStream.classList.remove('active');
        if (streamIcon) streamIcon.textContent = '🌊';
        if (streamLabel) streamLabel.textContent = 'Sustained Stream';
        if (streamSub) streamSub.textContent = 'Continuous 30 rps';
      } else {
        state.isStreaming = true;
        btnFireStream.classList.add('active');
        if (streamIcon) streamIcon.textContent = '🛑';
        if (streamLabel) streamLabel.textContent = 'Halt Stream';
        if (streamSub) streamSub.textContent = 'Active 35 rps';
        state.streamingInterval = setInterval(() => {
          dispatchRateLimitCheck(1);
        }, 28);
      }
    });
  }

  if (btnFireDdos) {
    btnFireDdos.addEventListener('click', () => {
      playSynthSound('ddos');
      appendTerminalLog('sys', '🚨 VOLUMETRIC DDoS ATTACK DETECTED (120 requests in 200ms)! Defense active.', 'BOTNET', 0);
      for (let i = 0; i < 120; i++) {
        setTimeout(() => dispatchRateLimitCheck(1), Math.random() * 250);
      }
    });
  }

  if (btnChaosStorm) {
    btnChaosStorm.addEventListener('click', () => {
      appendTerminalLog('sys', '🌪️ MULTI-TENANT CHAOS STORM: Simulating 5 distinct client tiers simultaneously.', 'CHAOS', 0);
      const keys = ['pay_gateway', 'auth_service', 'mobile_ios', 'web_crawler', 'analytics_agent'];
      for (let i = 0; i < 50; i++) {
        setTimeout(() => {
          const randKey = keys[Math.floor(Math.random() * keys.length)];
          if (inputClient) inputClient.value = randKey;
          dispatchRateLimitCheck(1);
        }, i * 30);
      }
    });
  }

  if (btnResetQuota) {
    btnResetQuota.addEventListener('click', async () => {
      const key = (inputClient ? inputClient.value.trim() : state.key) || 'client_edge_alpha';
      try {
        await fetch('/v1/limiter/reset', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ key: key, algorithm: state.algorithm })
        });
        appendTerminalLog('sys', `Quota reset for key "${key}"`, key, 0);
        dispatchRateLimitCheck(0);
      } catch (e) {
        appendTerminalLog('sys', `Reset failed: ${e.message}`, key, 0);
      }
    });
  }

  if (btnClearLogs) {
    btnClearLogs.addEventListener('click', () => {
      if (terminalStream) terminalStream.innerHTML = '';
    });
  }

  // --- BACKGROUND METRICS & CHART UPDATE LOOP (every 500ms) ---
  setInterval(() => {
    const deltaAllowed = state.recentAllowed;
    const deltaBlocked = state.recentBlocked;
    state.recentAllowed = 0;
    state.recentBlocked = 0;

    const rps = (deltaAllowed + deltaBlocked) * 2;
    state.currentRps = rps;

    if (kpiRps) kpiRps.textContent = rps.toLocaleString();
    if (trackRps) {
      const pct = Math.min(100, (rps / 200) * 100);
      trackRps.style.width = `${Math.max(5, pct)}%`;
    }

    // Shift Oscilloscope history
    histAllowed.shift();
    histAllowed.push(deltaAllowed * 2);
    histBlocked.shift();
    histBlocked.push(deltaBlocked * 2);
  }, 500);

  // Initial Check on Load
  setTimeout(() => {
    dispatchRateLimitCheck(0);
  }, 300);

})();
