/* charts.js — lightweight SVG chart library for the Vyzorix mock.
 *
 * Grafana-style dark charts: subtle dashed grid, right-aligned value axis,
 * gradient area fills, rounded line joins, end-point markers, a legend row and
 * a hover crosshair tooltip. No dependencies; sizes to the host element.
 *
 * Exposes window.VyzCharts = { sparkline, lineChart, barChart, donutChart, heatmap }.
 */
(function () {
  const PALETTE = ["#fb7185", "#73bf69", "#ff9830", "#f2495c", "#b877d9", "#fade2a"];
  const AXIS = "var(--text-disabled)";
  const GRID = "var(--border-weak)";
  const esc = (s) => String(s).replace(/[&<>"]/g, (c) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;" }[c]));
  const fmt = (v) => (Math.abs(v) >= 1000 ? (v / 1000).toFixed(1) + "k" : Math.round(v * 10) / 10);
  const color = (i, c) => c || PALETTE[i % PALETTE.length];

  let tipEl;
  function tip() {
    if (!tipEl) { tipEl = document.createElement("div"); tipEl.className = "chart-tip"; document.body.appendChild(tipEl); }
    return tipEl;
  }

  /* ------------------------------------------------ sparkline (stat panels) */
  function sparkline(values, opts = {}) {
    const w = opts.width || 240, h = opts.height || 40;
    const c = opts.color || PALETTE[0];
    const min = Math.min(...values), max = Math.max(...values);
    const span = max - min || 1;
    const x = (i) => (i / (values.length - 1)) * w;
    const y = (v) => h - 3 - ((v - min) / span) * (h - 8);
    const d = values.map((v, i) => `${i ? "L" : "M"}${x(i).toFixed(1)} ${y(v).toFixed(1)}`).join(" ");
    const area = `${d} L ${w} ${h} L 0 ${h} Z`;
    const lx = x(values.length - 1), ly = y(values[values.length - 1]);
    return `<svg class="spark" viewBox="0 0 ${w} ${h}" preserveAspectRatio="none" width="100%" height="${h}">
      <defs><linearGradient id="sp${optid(opts)}" x1="0" y1="0" x2="0" y2="1">
        <stop offset="0%" stop-color="${c}" stop-opacity="0.35"/><stop offset="100%" stop-color="${c}" stop-opacity="0"/>
      </linearGradient></defs>
      <path d="${area}" fill="url(#sp${optid(opts)})"/>
      <path d="${d}" fill="none" stroke="${c}" stroke-width="1.6" stroke-linejoin="round" stroke-linecap="round"/>
      <circle cx="${lx.toFixed(1)}" cy="${ly.toFixed(1)}" r="2.4" fill="${c}"/>
    </svg>`;
  }
  let idc = 0; const optid = (o) => (o.__id || (o.__id = "g" + idc++));

  /* --------------------------------------------------- multi-series area line */
  function lineChart(host, opts) {
    const series = opts.series || [];
    if (!host || !series.length) return;
    const xLabels = opts.xLabels || [];
    const H = opts.height || 220;
    const W = Math.max(320, host.clientWidth || 640);
    const legendH = opts.legend === false ? 0 : 30;
    const pad = { t: 14, r: 54, b: (xLabels.length ? 20 : 8) + legendH, l: 12 };
    const plotW = W - pad.l - pad.r;
    const plotH = H - pad.t - pad.b;
    const all = series.flatMap((s) => s.values);
    let min = Math.min(...all), max = Math.max(...all);
    const span = max - min || Math.abs(max) || 1;
    min -= span * 0.16; max += span * 0.16;
    const n = Math.max(...series.map((s) => s.values.length));
    const X = (i) => pad.l + (n <= 1 ? plotW / 2 : (i / (n - 1)) * plotW);
    const Y = (v) => pad.t + (1 - (v - min) / (max - min)) * plotH;

    let svg = "";
    const gc = 4;
    for (let g = 0; g <= gc; g++) {
      const val = min + (g / gc) * (max - min);
      const yy = Y(val).toFixed(1);
      svg += `<line x1="${pad.l}" y1="${yy}" x2="${pad.l + plotW}" y2="${yy}" stroke="${GRID}"${g ? ' stroke-dasharray="3 4"' : ""}/>`;
      svg += `<text x="${pad.l + plotW + 8}" y="${(+yy + 3).toFixed(1)}" fill="${AXIS}" font-size="10">${fmt(val)}${esc(opts.unit || "")}</text>`;
    }
    xLabels.forEach((lab, i) => {
      const xx = (pad.l + (i / Math.max(1, xLabels.length - 1)) * plotW).toFixed(1);
      svg += `<text x="${xx}" y="${pad.t + plotH + 14}" fill="${AXIS}" font-size="10" text-anchor="middle">${esc(lab)}</text>`;
    });

    let defs = "";
    series.forEach((s, si) => {
      const c = color(si, s.color);
      const pts = s.values.map((v, i) => [X(i), Y(v)]);
      const d = pts.map((p, i) => (i ? "L" : "M") + p[0].toFixed(1) + " " + p[1].toFixed(1)).join(" ");
      const area = `${d} L ${pts[pts.length - 1][0].toFixed(1)} ${(pad.t + plotH).toFixed(1)} L ${pts[0][0].toFixed(1)} ${(pad.t + plotH).toFixed(1)} Z`;
      defs += `<linearGradient id="ln${idc}-${si}" x1="0" y1="0" x2="0" y2="1"><stop offset="0%" stop-color="${c}" stop-opacity="0.30"/><stop offset="100%" stop-color="${c}" stop-opacity="0.02"/></linearGradient>`;
      svg += `<path d="${area}" fill="url(#ln${idc}-${si})"/>`;
      svg += `<path d="${d}" fill="none" stroke="${c}" stroke-width="1.75" stroke-linejoin="round" stroke-linecap="round"/>`;
      const last = pts[pts.length - 1];
      svg += `<circle cx="${last[0].toFixed(1)}" cy="${last[1].toFixed(1)}" r="2.8" fill="${c}" stroke="var(--gray-10)" stroke-width="1.5"/>`;
    });
    const gid = idc++;
    idc++; // keep gradient ids unique per chart

    const legend = opts.legend === false ? "" :
      `<div class="chart-legend">${series.map((s, si) => `<span><i style="background:${color(si, s.color)}"></i>${esc(s.name || "series " + (si + 1))}</span>`).join("")}</div>`;

    host.innerHTML = `<svg class="chart-svg" width="${W}" height="${H}" viewBox="0 0 ${W} ${H}">${defs}${svg}
      <line class="crosshair" x1="0" y1="${pad.t}" x2="0" y2="${pad.t + plotH}" stroke="var(--border-strong)" stroke-dasharray="3 3" opacity="0"/>
      <circle class="focus" r="3.2" fill="none" stroke="#fff" stroke-width="1.4" opacity="0"/></svg>${legend}`;

    // Hover crosshair + tooltip
    const svgEl = host.querySelector("svg");
    const cross = svgEl.querySelector(".crosshair");
    const focus = svgEl.querySelector(".focus");
    svgEl.addEventListener("mousemove", (e) => {
      const r = svgEl.getBoundingClientRect();
      const rel = ((e.clientX - r.left) / r.width) * W;
      if (rel < pad.l - 4 || rel > pad.l + plotW + 4) return;
      const idx = Math.round(((rel - pad.l) / plotW) * (n - 1));
      const xx = X(idx);
      cross.setAttribute("x1", xx); cross.setAttribute("x2", xx); cross.setAttribute("opacity", "1");
      const lead = series[0];
      focus.setAttribute("cx", xx); focus.setAttribute("cy", Y(lead.values[idx])); focus.setAttribute("opacity", "1");
      const label = xLabels.length ? xLabels[Math.round((idx / (n - 1)) * (xLabels.length - 1))] : "t" + idx;
      tip().innerHTML = `<div class="chart-tip__t">${esc(label)}</div>` + series.map((s, si) =>
        `<div class="chart-tip__r"><i style="background:${color(si, s.color)}"></i><span>${esc(s.name || "")}</span><b>${fmt(s.values[idx])}${esc(opts.unit || "")}</b></div>`
      ).join("");
      tip().classList.add("is-on");
      tip().style.left = e.clientX + 14 + "px";
      tip().style.top = e.clientY + 14 + "px";
    });
    svgEl.addEventListener("mouseleave", () => { cross.setAttribute("opacity", "0"); focus.setAttribute("opacity", "0"); tip().classList.remove("is-on"); });
  }

  /* ------------------------------------------------------------- bar chart */
  function barChart(host, opts) {
    const data = opts.data || [];
    if (!host || !data.length) return;
    const H = opts.height || 180;
    const W = Math.max(280, host.clientWidth || 520);
    const pad = { t: 18, r: 12, b: 26, l: 12 };
    const plotW = W - pad.l - pad.r;
    const plotH = H - pad.t - pad.b;
    const max = Math.max(...data.map((d) => d.value)) * 1.15 || 1;
    const bw = Math.min(48, (plotW / data.length) * 0.62);
    const step = plotW / data.length;
    let bars = "";
    data.forEach((d, i) => {
      const c = color(i, d.color);
      const cx = pad.l + step * i + step / 2;
      const bh = (d.value / max) * plotH;
      const y = pad.t + plotH - bh;
      bars += `<rect x="${(cx - bw / 2).toFixed(1)}" y="${y.toFixed(1)}" width="${bw.toFixed(1)}" height="${Math.max(1, bh).toFixed(1)}" rx="3" fill="${c}" opacity="0.9"/>`;
      if (opts.showValues !== false)
        bars += `<text x="${cx.toFixed(1)}" y="${(y - 5).toFixed(1)}" fill="var(--text-secondary)" font-size="10" text-anchor="middle">${fmt(d.value)}</text>`;
      bars += `<text x="${cx.toFixed(1)}" y="${H - 8}" fill="${AXIS}" font-size="10" text-anchor="middle">${esc(d.label)}</text>`;
    });
    let grid = "";
    for (let g = 0; g <= 3; g++) {
      const yy = (pad.t + (g / 3) * plotH).toFixed(1);
      grid += `<line x1="${pad.l}" y1="${yy}" x2="${pad.l + plotW}" y2="${yy}" stroke="${GRID}"${g ? ' stroke-dasharray="3 4"' : ""}/>`;
    }
    host.innerHTML = `<svg class="chart-svg" width="${W}" height="${H}" viewBox="0 0 ${W} ${H}">${grid}${bars}</svg>`;
  }

  /* -------------------------------------------------------- donut / ring */
  function donutChart(host, opts) {
    const segs = (opts.segments || []).filter((s) => s.value > 0);
    if (!host || !segs.length) return;
    const H = opts.height || 180;
    const W = Math.max(220, host.clientWidth || 320);
    const cx = W / 2, cy = H / 2;
    const R = Math.min(W, H) / 2 - 12;
    const r = R * 0.62;
    const total = segs.reduce((a, s) => a + s.value, 0) || 1;
    let a0 = -Math.PI / 2, paths = "";
    segs.forEach((s, i) => {
      const frac = s.value / total;
      const a1 = a0 + frac * Math.PI * 2;
      const large = frac > 0.5 ? 1 : 0;
      const p = (ang, rad) => [cx + Math.cos(ang) * rad, cy + Math.sin(ang) * rad];
      const [x0, y0] = p(a0, R), [x1, y1] = p(a1, R), [x2, y2] = p(a1, r), [x3, y3] = p(a0, r);
      const col = color(i, s.color);
      paths += `<path d="M ${x0.toFixed(1)} ${y0.toFixed(1)} A ${R} ${R} 0 ${large} 1 ${x1.toFixed(1)} ${y1.toFixed(1)} L ${x2.toFixed(1)} ${y2.toFixed(1)} A ${r} ${r} 0 ${large} 0 ${x3.toFixed(1)} ${y3.toFixed(1)} Z" fill="${col}" opacity="0.92"/>`;
      a0 = a1;
    });
    const center = opts.centerValue != null
      ? `<text x="${cx}" y="${cy - 2}" fill="var(--text-primary)" font-size="22" font-weight="600" text-anchor="middle">${esc(opts.centerValue)}</text>
         <text x="${cx}" y="${cy + 16}" fill="var(--text-secondary)" font-size="10" text-anchor="middle">${esc(opts.centerLabel || "")}</text>` : "";
    const legend = `<div class="donut-legend">${segs.map((s, i) =>
      `<span><i style="background:${color(i, s.color)}"></i>${esc(s.label)} <b>${s.value}</b></span>`).join("")}</div>`;
    host.innerHTML = `<div class="donut-wrap"><svg class="chart-svg" width="${W}" height="${H}" viewBox="0 0 ${W} ${H}">${paths}${center}</svg>${legend}</div>`;
  }

  /* ------------------------------------------------------------- heatmap */
  function heatmap(host, opts) {
    const { rows = [], cols = [], values = [], rowLabels = [], colLabels = [], palette = ["#132a1f", "#1a7f4b", "#73bf69", "#fade2a", "#ff9830", "#f2495c"] } = opts;
    if (!host) return;
    const W = Math.max(320, host.clientWidth || 640);
    const padL = 70, padT = 6, padR = 8, padB = 20;
    const cellW = (W - padL - padR) / cols;
    const cellH = 20;
    const H = padT + rows * cellH + padB;
    const pick = (v) => palette[Math.min(palette.length - 1, Math.max(0, Math.round(v * (palette.length - 1))))];
    let cells = "";
    for (let ri = 0; ri < rows; ri++) {
      cells += `<text x="${padL - 8}" y="${padT + ri * cellH + cellH / 2 + 3}" fill="${AXIS}" font-size="10" text-anchor="end">${esc(rowLabels[ri] || "")}</text>`;
      for (let ci = 0; ci < cols; ci++) {
        const v = values[ri][ci];
        cells += `<rect x="${(padL + ci * cellW).toFixed(1)}" y="${(padT + ri * cellH + 1).toFixed(1)}" width="${(cellW - 2).toFixed(1)}" height="${cellH - 2}" rx="2" fill="${pick(v)}"/>`;
      }
    }
    for (let ci = 0; ci < cols; ci++) {
      if (ci % 2) continue;
      cells += `<text x="${(padL + ci * cellW + cellW / 2).toFixed(1)}" y="${H - 6}" fill="${AXIS}" font-size="10" text-anchor="middle">${esc(colLabels[ci] || "")}</text>`;
    }
    host.innerHTML = `<svg class="chart-svg" width="${W}" height="${H}" viewBox="0 0 ${W} ${H}">${cells}</svg>`;
  }

  /* ------------------------------------------------------- stacked histogram */
  function histogram(host, opts) {
    const buckets = opts.buckets || [];
    if (!host || !buckets.length) return;
    const levels = opts.levels || ["error", "warn", "info", "debug"];
    const colors = opts.colors || { error: "#f2495c", warn: "#ff9830", info: "#fb7185", debug: "#5b5b5b" };
    const H = opts.height || 90;
    const W = Math.max(300, host.clientWidth || 640);
    const pad = { t: 6, r: 6, b: 16, l: 6 };
    const plotW = W - pad.l - pad.r, plotH = H - pad.t - pad.b;
    const totals = buckets.map((b) => levels.reduce((a, l) => a + (b[l] || 0), 0));
    const max = Math.max(1, ...totals);
    const bw = plotW / buckets.length;
    let bars = "";
    buckets.forEach((b, i) => {
      const x = pad.l + i * bw;
      let y = pad.t + plotH;
      levels.forEach((l) => {
        const v = b[l] || 0;
        if (!v) return;
        const h = (v / max) * plotH;
        y -= h;
        bars += `<rect x="${(x + bw * 0.12).toFixed(1)}" y="${y.toFixed(1)}" width="${(bw * 0.76).toFixed(1)}" height="${h.toFixed(1)}" fill="${colors[l] || "#fb7185"}" opacity="0.9"><title>${b.label} · ${l}: ${v}</title></rect>`;
      });
      if (i % 5 === 0) bars += `<text x="${(x + bw / 2).toFixed(1)}" y="${H - 4}" fill="${AXIS}" font-size="9" text-anchor="middle">${esc(b.label)}</text>`;
    });
    host.innerHTML = `<svg class="chart-svg" width="${W}" height="${H}" viewBox="0 0 ${W} ${H}">${bars}</svg>`;
  }

  window.VyzCharts = { sparkline, lineChart, barChart, donutChart, heatmap, histogram, PALETTE };
})();
