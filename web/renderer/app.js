(function() {
  var DEVICES = {
    iphone15: { width: 393, height: 852 },
    pixel8:  { width: 412, height: 915 },
    ipad:    { width: 744, height: 1133 },
  };

  var devServerURL = '';
  var ws;

  function init() {
    fetch('/api/config')
      .then(function(r) { return r.json(); })
      .then(function(cfg) { devServerURL = cfg.devServerURL; })
      .catch(function() { devServerURL = 'http://localhost:5173'; })
      .finally(function() {
        connectWS();
        setupDevicePicker();
        loadApp();
      });
  }

  function connectWS() {
    var proto = location.protocol === 'https:' ? 'wss' : 'ws';
    ws = new WebSocket(proto + '://' + location.host + '/ws');
    ws.onmessage = function(e) {
      var msg = JSON.parse(e.data);
      if (msg.type === 'reload') loadApp();
    };
    ws.onclose = function() {
      setStatus('reconnecting');
      setTimeout(connectWS, 2000);
    };
    ws.onopen = function() { setStatus('live'); };

    // forward console errors from iframe
    window.addEventListener('message', function(e) {
      if (e.data && e.data.type === 'vibeview-error') {
        showError(e.data.message, e.data.file, e.data.line);
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
    var buttons = document.querySelectorAll('#device-picker button');
    for (var i = 0; i < buttons.length; i++) {
      var b = buttons[i];
      if (b.dataset.device === name) {
        b.classList.add('active');
      } else {
        b.classList.remove('active');
      }
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

  init();
})();
