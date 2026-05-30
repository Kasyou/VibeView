(function() {
  var DEVICES = {
    iphone15: { width: 393, height: 852 },
    pixel8:  { width: 412, height: 915 },
    ipad:    { width: 744, height: 1133 },
  };

  var devServerURL = '';
  var ws;

  function init() {
    loadPrefs();
    fetch('/api/config')
      .then(function(r) { return r.json(); })
      .then(function(cfg) { devServerURL = cfg.devServerURL; })
      .catch(function() { devServerURL = 'http://localhost:5173'; })
      .finally(function() {
        connectWS();
        setupDevicePicker();
        // Restore saved device preference
        if (currentDevice === 'custom') {
          setCustomSize(customW, customH);
        } else {
          setDevice(currentDevice);
        }
        loadApp();
      });
  }

  // --- Preferences ---
  var PREFS_KEY = 'vibeview_prefs';

  function loadPrefs() {
    try {
      var raw = localStorage.getItem(PREFS_KEY);
      if (raw) {
        var p = JSON.parse(raw);
        if (p.device) currentDevice = p.device;
        if (p.customW) customW = p.customW;
        if (p.customH) customH = p.customH;
      }
    } catch(e) {}
  }

  function savePrefs() {
    try {
      localStorage.setItem(PREFS_KEY, JSON.stringify({
        device: currentDevice,
        customW: customW,
        customH: customH
      }));
    } catch(e) {}
  }

  var currentDevice = 'iphone15';
  var customW = 375, customH = 812;

  function connectWS() {
    var proto = location.protocol === 'https:' ? 'wss' : 'ws';
    ws = new WebSocket(proto + '://' + location.host + '/ws');
    ws.onmessage = function(e) {
      var msg = JSON.parse(e.data);
      if (msg.type === 'reload') loadApp();
      if (msg.type === 'screenshot-request') captureScreenshot((msg.data && msg.data.id) || '');
      if (msg.type === 'inspect-request') handleInspect((msg.data && msg.data.id) || '', msg.data || {});
    };
    ws.onclose = function() {
      setStatus('reconnecting');
      setTimeout(connectWS, 2000);
    };
    ws.onopen = function() { setStatus('live'); };

    window.addEventListener('message', function(e) {
      if (e.data && e.data.type === 'vibeview-error') {
        showError(e.data.message, e.data.file, e.data.line);
        if (ws && ws.readyState === WebSocket.OPEN) {
          ws.send(JSON.stringify({
            type: 'console',
            data: { level: 'error', message: e.data.message, file: e.data.file || '', line: e.data.line || 0 }
          }));
        }
      }
    });
  }

  function setStatus(text) {
    var el = document.getElementById('status');
    if (el) el.textContent = text;
  }

  function setupDevicePicker() {
    var buttons = document.querySelectorAll('#device-picker button');
    for (var i = 0; i < buttons.length; i++) {
      buttons[i].addEventListener('click', function() {
        var dev = this.dataset.device;
        if (dev === 'custom') {
          document.getElementById('custom-size').style.display = 'flex';
          return;
        }
        document.getElementById('custom-size').style.display = 'none';
        setDevice(dev);
      });
    }
    document.getElementById('custom-apply').addEventListener('click', function() {
      var w = parseInt(document.getElementById('custom-w').value) || 375;
      var h = parseInt(document.getElementById('custom-h').value) || 812;
      setCustomSize(w, h);
    });
  }

  function setDevice(name) {
    currentDevice = name;
    savePrefs();
    var buttons = document.querySelectorAll('#device-picker button');
    for (var i = 0; i < buttons.length; i++) {
      var b = buttons[i];
      b.classList.toggle('active', b.dataset.device === name);
    }
    var frame = document.getElementById('device-frame');
    var iframe = document.getElementById('app-frame');
    frame.className = name;

    if (name === 'full') {
      iframe.style.width = '100%';
      iframe.style.height = '100%';
    } else if (DEVICES[name]) {
      iframe.style.width = DEVICES[name].width + 'px';
      iframe.style.height = DEVICES[name].height + 'px';
    }
  }

  function setCustomSize(w, h) {
    customW = w; customH = h;
    savePrefs();
    var buttons = document.querySelectorAll('#device-picker button');
    for (var i = 0; i < buttons.length; i++) {
      buttons[i].classList.remove('active');
    }
    var frame = document.getElementById('device-frame');
    var iframe = document.getElementById('app-frame');
    frame.className = '';
    frame.style.width = (w + 24) + 'px';
    frame.style.height = (h + 24) + 'px';
    iframe.style.width = w + 'px';
    iframe.style.height = h + 'px';
  }

  function loadApp() {
    var iframe = document.getElementById('app-frame');
    var url = devServerURL + '?_v=' + Date.now();
    iframe.src = url;
    hideError();
  }

  function showError(message, file, line) {
    var overlay = document.getElementById('error-overlay');
    document.getElementById('error-message').textContent = message;
    document.getElementById('error-file').textContent = (file || '') + (line ? ':' + line : '');
    overlay.style.display = 'block';
    setTimeout(hideError, 8000);
  }

  function hideError() {
    document.getElementById('error-overlay').style.display = 'none';
  }

  // --- Screenshot ---
  function captureScreenshot(reqId) {
    try {
      var W = 800, H = 500;
      var canvas = document.createElement('canvas');
      canvas.width = W;
      canvas.height = H;
      var ctx = canvas.getContext('2d');

      // Background
      ctx.fillStyle = '#1a1a2e';
      ctx.fillRect(0, 0, W, H);

      // Toolbar
      ctx.fillStyle = '#16213e';
      ctx.fillRect(0, 0, W, 36);
      ctx.fillStyle = '#e94560';
      ctx.font = 'bold 13px sans-serif';
      ctx.fillText('VibeView', 12, 24);

      var active = document.querySelector('#device-picker button.active');
      var devName = active ? active.textContent : 'iPhone 15 Pro';
      ctx.fillStyle = '#a0a0b0';
      ctx.font = '11px sans-serif';
      ctx.fillText(devName, 85, 24);

      var statusEl = document.getElementById('status');
      var st = statusEl ? statusEl.textContent : '';
      ctx.fillStyle = st === 'live' ? '#4caf50' : '#e94560';
      ctx.fillText(st, W - 50, 24);

      // Device frame
      var fw = 280, fh = 420, fx = (W - fw) / 2, fy = 60, r;
      ctx.fillStyle = '#111';
      ctx.strokeStyle = '#333';
      ctx.lineWidth = 2;
      r = 24;
      roundRect(ctx, fx - 4, fy - 4, fw + 8, fh + 8, r);
      ctx.fill();
      ctx.stroke();

      // Screen area — white background
      ctx.fillStyle = '#fff';
      r = 12;
      roundRect(ctx, fx, fy, fw, fh, r);
      ctx.fill();

      // Show preview URL or content summary
      ctx.fillStyle = '#333';
      ctx.font = 'bold 13px sans-serif';
      ctx.textAlign = 'center';
      ctx.fillText(devServerURL.replace('http://localhost:51820/_app/', ''), fx + fw/2, fy + fh/2 - 4);
      ctx.fillStyle = '#999';
      ctx.font = '11px sans-serif';
      ctx.fillText('Preview active — open browser to view live', fx + fw/2, fy + fh/2 + 14);
      ctx.textAlign = 'left';

      // Notch for iPhone
      if (devName.indexOf('iPhone') >= 0) {
        ctx.fillStyle = '#111';
        ctx.beginPath();
        var nx = fx + fw/2 - 50, ny = fy - 4, nw = 100, nh = 22;
        ctx.moveTo(nx + 12, ny);
        ctx.lineTo(nx + nw - 12, ny);
        ctx.arcTo(nx + nw, ny, nx + nw, ny + 12, 12);
        ctx.lineTo(nx + nw, ny + nh);
        ctx.lineTo(nx, ny + nh);
        ctx.lineTo(nx, ny + 12);
        ctx.arcTo(nx, ny, nx + 12, ny, 12);
        ctx.closePath();
        ctx.fill();
      }

      // Error overlay
      var errOverlay = document.getElementById('error-overlay');
      if (errOverlay && errOverlay.style.display !== 'none') {
        var errMsg = document.getElementById('error-message');
        var errFile = document.getElementById('error-file');
        var msg = errMsg ? errMsg.textContent.substring(0, 60) : '';
        ctx.fillStyle = 'rgba(233,69,96,0.95)';
        var ew = Math.min(460, W - 40), eh = 48;
        var ex = (W - ew) / 2, ey = H - eh - 16;
        roundRect(ctx, ex, ey, ew, eh, 8);
        ctx.fill();
        ctx.fillStyle = '#fff';
        ctx.font = '12px sans-serif';
        ctx.fillText('! ' + msg, ex + 14, ey + 20);
        if (errFile) {
          ctx.font = '10px sans-serif';
          ctx.fillText(errFile.textContent.substring(0, 40), ex + 14, ey + 36);
        }
      }

      var png;
      try { png = canvas.toDataURL('image/png'); } catch(e) { png = ''; }
      wsSend({type: 'screenshot-data', id: reqId, data: {image: png}});
    } catch(e) {
      wsSend({type: 'screenshot-data', id: reqId, data: {image: '', error: e.message}});
    }
  }

  function roundRect(ctx, x, y, w, h, r) {
    ctx.beginPath();
    ctx.moveTo(x + r, y);
    ctx.lineTo(x + w - r, y);
    ctx.arcTo(x + w, y, x + w, y + r, r);
    ctx.lineTo(x + w, y + h - r);
    ctx.arcTo(x + w, y + h, x + w - r, y + h, r);
    ctx.lineTo(x + r, y + h);
    ctx.arcTo(x, y + h, x, y + h - r, r);
    ctx.lineTo(x, y + r);
    ctx.arcTo(x, y, x + r, y, r);
    ctx.closePath();
  }

  function wsSend(data) {
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify(data));
    }
  }

  // --- Inspect ---
  function handleInspect(reqId, data) {
    var result = { found: false, selector: data.selector || '' };
    try {
      // Try to query the iframe first (same-origin HTML projects)
      var iframe = document.getElementById('app-frame');
      var doc = null;
      try { doc = iframe.contentDocument || iframe.contentWindow.document; } catch(e) {}

      var el = null;
      if (doc && data.selector) {
        el = doc.querySelector(data.selector);
      }
      if (!el && data.selector) {
        el = document.querySelector(data.selector);
      }

      if (el) {
        var rect = el.getBoundingClientRect();
        var style = window.getComputedStyle(el);
        result = {
          found: true,
          selector: data.selector,
          tag: el.tagName.toLowerCase(),
          text: (el.textContent || '').substring(0, 200),
          rect: { x: Math.round(rect.x), y: Math.round(rect.y), w: Math.round(rect.width), h: Math.round(rect.height) },
          display: style.display,
          visibility: style.visibility,
          color: style.color,
          backgroundColor: style.backgroundColor,
          fontSize: style.fontSize,
          fontWeight: style.fontWeight,
          className: el.className || '',
          id: el.id || ''
        };
      }
    } catch(e) {
      result.error = e.message;
    }
    wsSend({type: 'inspect-data', id: reqId, data: result});
  }

  init();
})();
