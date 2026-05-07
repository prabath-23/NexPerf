package server

const fallbackHTML = `<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <title>NexPerf</title>
    <style>
      body { margin: 0; font-family: ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; background: #101418; color: #edf1f5; }
      main { max-width: 760px; margin: 0 auto; padding: 40px 24px; }
      h1 { margin: 0 0 10px; font-size: 30px; letter-spacing: 0; }
      p { color: #a9b5c0; line-height: 1.6; }
      code { background: #171c22; border: 1px solid #2a333d; padding: 2px 6px; border-radius: 5px; }
    </style>
  </head>
  <body>
    <main>
      <h1>NexPerf API is running</h1>
      <p>The Vue dashboard build was not found at <code>web/dist</code>. Run <code>npm --prefix web run build</code> and restart NexPerf to serve the production dashboard at <code>/nexperf</code>.</p>
    </main>
  </body>
</html>`
