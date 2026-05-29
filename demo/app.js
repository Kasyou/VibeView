(function() {
  var count = 0;
  var colors = ['#667eea', '#e94560', '#0f3460', '#16a085', '#e67e22'];
  var colorIdx = 0;

  document.getElementById('increment').addEventListener('click', function() {
    count++;
    document.getElementById('count').textContent = count;
    log('Count: ' + count);
  });

  document.getElementById('decrement').addEventListener('click', function() {
    count--;
    document.getElementById('count').textContent = count;
    log('Count: ' + count);
  });

  document.getElementById('color-btn').addEventListener('click', function() {
    colorIdx = (colorIdx + 1) % colors.length;
    document.querySelector('.card').style.background = colors[colorIdx] + '11';
    log('Color changed to ' + colors[colorIdx]);
  });

  function log(msg) {
    var el = document.getElementById('status-msg');
    el.textContent = msg;
    // Also send to VibeView for console forwarding
    try {
      window.parent.postMessage({
        type: 'vibeview-log',
        level: 'info',
        message: msg
      }, '*');
    } catch(e) {}
  }

  console.log('VibeView Demo ready');
})();
