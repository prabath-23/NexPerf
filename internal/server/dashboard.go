package server

const dashboardHTML = `<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <title>NexPerf</title>
    <style>
      :root { color-scheme: dark; font-family: Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; background: #111418; color: #edf1f5; }
      body { margin: 0; background: #111418; }
      main { max-width: 1180px; margin: 0 auto; padding: 28px; }
      header { display: flex; justify-content: space-between; gap: 18px; align-items: end; border-bottom: 1px solid #29313a; padding-bottom: 18px; }
      h1 { margin: 0; font-size: 30px; letter-spacing: 0; }
      .muted { color: #9aa7b4; }
      .grid { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 14px; margin: 22px 0; }
      .card, section { background: #171c22; border: 1px solid #2a333d; border-radius: 8px; padding: 16px; }
      .label { color: #9aa7b4; font-size: 13px; }
      .value { font-size: 28px; margin-top: 8px; font-weight: 700; }
      .bar { height: 8px; background: #2a333d; border-radius: 999px; overflow: hidden; margin-top: 14px; }
      .fill { height: 100%; background: #4cc9a7; }
      section { margin-top: 16px; }
      h2 { margin: 0 0 14px; font-size: 18px; }
      table { width: 100%; border-collapse: collapse; }
      th, td { text-align: left; padding: 10px 8px; border-bottom: 1px solid #29313a; font-size: 14px; }
      th { color: #9aa7b4; font-weight: 600; }
      .insight { border-left: 3px solid #4cc9a7; padding: 10px 12px; background: #12171c; margin: 10px 0; }
      .warning { border-left-color: #f2b84b; }
      .critical { border-left-color: #ff6b6b; }
      @media (max-width: 760px) { main { padding: 18px; } header { display: block; } .grid { grid-template-columns: 1fr; } table { font-size: 13px; } }
    </style>
  </head>
  <body>
    <main>
      <header>
        <div>
          <h1>NexPerf</h1>
          <div class="muted">Local system intelligence dashboard</div>
        </div>
        <div class="muted" id="meta">Loading...</div>
      </header>
      <div class="grid" id="cards"></div>
      <section>
        <h2>Top Processes</h2>
        <table>
          <thead><tr><th>PID</th><th>Name</th><th>Memory</th><th>CPU</th><th>User</th></tr></thead>
          <tbody id="processes"></tbody>
        </table>
      </section>
      <section>
        <h2>Insights</h2>
        <div id="insights"></div>
      </section>
    </main>
    <script>
      const fmtBytes = (bytes) => {
        const gb = bytes / 1024 / 1024 / 1024;
        return gb >= 10 ? gb.toFixed(0) + " GB" : gb.toFixed(1) + " GB";
      };
      const pct = (n) => Number(n || 0).toFixed(1) + "%";
      async function refresh() {
        const [system, processes, insights] = await Promise.all([
          fetch("/api/system").then(r => r.json()),
          fetch("/api/processes/top").then(r => r.json()),
          fetch("/api/insights").then(r => r.json())
        ]);
        document.getElementById("meta").textContent = system.os + " / " + system.arch + (system.hostname ? " / " + system.hostname : "");
        document.getElementById("cards").innerHTML = [
          ["CPU", pct(system.cpu_percent), system.cpu_percent],
          ["Memory", pct(system.memory.percent), system.memory.percent, fmtBytes(system.memory.used) + " used of " + fmtBytes(system.memory.total)],
          ["Disk", pct(system.disk.percent), system.disk.percent, fmtBytes(system.disk.used) + " used of " + fmtBytes(system.disk.total)]
        ].map(([label, value, percent, detail]) => '<div class="card"><div class="label">' + label + '</div><div class="value">' + value + '</div><div class="muted">' + (detail || "Current aggregate usage") + '</div><div class="bar"><div class="fill" style="width:' + Math.min(percent, 100) + '%"></div></div></div>').join("");
        document.getElementById("processes").innerHTML = processes.map(p => '<tr><td>' + p.pid + '</td><td>' + p.name + '</td><td>' + p.memory_mb.toFixed(1) + ' MB</td><td>' + pct(p.cpu_percent) + '</td><td>' + (p.user || "") + '</td></tr>').join("");
        document.getElementById("insights").innerHTML = insights.map(i => '<div class="insight ' + i.severity + '"><strong>' + i.title + '</strong><div class="muted">' + i.message + '</div><div>' + i.recommendation + '</div></div>').join("");
      }
      refresh();
      setInterval(refresh, 3000);
    </script>
  </body>
</html>`
