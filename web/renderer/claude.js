(function() {
  var ws;
  var emptyState = document.getElementById('empty-state');

  function connectWS() {
    var proto = location.protocol === 'https:' ? 'wss' : 'ws';
    ws = new WebSocket(proto + '://' + location.host + '/ws');
    ws.onmessage = function(e) {
      var msg = JSON.parse(e.data);
      if (msg.type === 'show-content' && msg.data) {
        addCard(msg.data.title || '', msg.data.content || '', msg.data.seq, msg.data.time);
      }
      if (msg.type === 'clear-board') {
        clearBoard();
      }
    };
    ws.onclose = function() {
      setStatus('reconnecting');
      setTimeout(connectWS, 2000);
    };
    ws.onopen = function() {
      setStatus('live');
      // Fetch any cards that were pushed before we connected
      fetch('/api/queue')
        .then(function(r) { return r.json(); })
        .then(function(cards) {
          cards.forEach(function(c) { addCard(c.title || '', c.content || '', c.seq, c.time); });
        });
    };
  }

  function setStatus(text) {
    var el = document.getElementById('status');
    if (el) el.textContent = text;
  }

  function addCard(title, content, seq, time) {
    if (emptyState) { emptyState.style.display = 'none'; }
    var card = document.createElement('div');
    card.className = 'card';

    var html = '';
    // Header: #seq · time · title
    var header = '';
    if (seq || time) {
      header = '<div class="card-meta">';
      if (seq) header += '<span class="card-seq">#' + seq + '</span>';
      if (time) header += '<span class="card-time">' + escapeHtml(time) + '</span>';
      header += '</div>';
    }
    if (title) {
      html += header + '<div class="card-title">' + escapeHtml(title) + '</div>';
    } else {
      html += header;
    }
    html += renderMarkdown(content);
    card.innerHTML = html;
    var board = document.getElementById('board');
    board.appendChild(card);
    // Render Mermaid diagrams in code blocks
    renderMermaid(card);
    // Limit to 30 cards, remove oldest
    var allCards = board.querySelectorAll('.card');
    if (allCards.length > 30) {
      allCards[0].remove();
    }
    card.scrollIntoView({behavior:'smooth',block:'end'});
  }

  function clearBoard() {
    var board = document.getElementById('board');
    board.innerHTML = '<div id="empty-state"><div class="icon">&#9670;</div><p>Claude will show ideas here</p></div>';
    emptyState = document.getElementById('empty-state');
  }

  // Simple markdown → HTML renderer
  function renderMarkdown(text) {
    var lines = text.split('\n');
    var out = '';
    var inCode = false, codeBuf = '', codeLang = '';
    var inTable = false, tableRows = [], tableAlign = [];

    function flushParagraph(buf) {
      if (!buf.trim()) return '';
      // Headings
      var m = buf.match(/^(#{1,6})\s+(.+)/);
      if (m) {
        var level = m[1].length;
        return '<h' + level + '>' + inline(m[2]) + '</h' + level + '>';
      }
      // HR
      if (/^[-*_]{3,}$/.test(buf.trim())) return '<hr>';
      // Blockquote
      if (buf.startsWith('> ')) {
        return '<blockquote>' + inline(buf.replace(/^> /,'')) + '</blockquote>';
      }
      // Tags line (comma-separated short items)
      if (/^[\w\s,.-]+$/.test(buf) && buf.split(',').length >= 2 && buf.length < 80) {
        return '<div>' + buf.split(',').map(function(s) {
          return '<span class="tag">' + escapeHtml(s.trim()) + '</span>';
        }).join('') + '</div>';
      }
      return '<p>' + inline(buf) + '</p>';
    }

    function inline(s) {
      s = escapeHtml(s);
      s = s.replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>');
      s = s.replace(/`(.+?)`/g, '<code>$1</code>');
      s = s.replace(/\[([^\]]+)\]\(([^)]+)\)/g, '<a href="$2" target="_blank">$1</a>');
      return s;
    }

    var para = '';
    for (var i = 0; i < lines.length; i++) {
      var line = lines[i];

      // Code block
      if (line.trim().startsWith('```')) {
        if (inCode) {
          var langClass = codeLang ? ' class="language-' + codeLang + '"' : '';
          out += '<pre><code' + langClass + '>' + escapeHtml(codeBuf.trimEnd()) + '</code></pre>';
          codeBuf = ''; codeLang = ''; inCode = false;
        } else {
          if (para) { out += flushParagraph(para); para = ''; }
          codeLang = line.trim().slice(3).trim();
          inCode = true;
        }
        continue;
      }
      if (inCode) { codeBuf += line + '\n'; continue; }

      // Table
      if (line.startsWith('|') && line.endsWith('|')) {
        if (para) { out += flushParagraph(para); para = ''; }
        var cells = line.slice(1, -1).split('|').map(function(c) { return c.trim(); });
        if (cells.every(function(c) { return /^[-:]+$/.test(c) || c === ''; })) {
          // Alignment row
          tableAlign = cells.map(function(c) {
            if (c.startsWith(':') && c.endsWith(':')) return 'center';
            if (c.endsWith(':')) return 'right';
            return 'left';
          });
        } else {
          tableRows.push(cells);
        }
        if (i === lines.length - 1 || !lines[i+1].startsWith('|')) {
          // End of table
          var tbl = '<table>';
          if (tableRows.length > 0) {
            tbl += '<thead><tr>';
            for (var j = 0; j < tableRows[0].length; j++) {
              tbl += '<th>' + inline(tableRows[0][j] || '') + '</th>';
            }
            tbl += '</tr></thead><tbody>';
            for (var r = 1; r < tableRows.length; r++) {
              tbl += '<tr>';
              for (var c = 0; c < tableRows[r].length; c++) {
                tbl += '<td>' + inline(tableRows[r][c] || '') + '</td>';
              }
              tbl += '</tr>';
            }
            tbl += '</tbody>';
          }
          tbl += '</table>';
          out += tbl;
          tableRows = []; tableAlign = [];
        }
        continue;
      }

      // Empty line — flush paragraph
      if (line.trim() === '') {
        if (para) { out += flushParagraph(para); para = ''; }
        continue;
      }

      // List item
      var listMatch = line.match(/^(\s*)([-*+]|\d+\.)\s+(.+)/);
      if (listMatch) {
        if (para && !para.match(/^(\s*)([-*+]|\d+\.)\s+/)) {
          out += flushParagraph(para); para = '';
        }
        para += line + '\n';
        if (i === lines.length - 1 || !lines[i+1].match(/^(\s*)([-*+]|\d+\.)\s+/)) {
          // End of list — process as paragraph-like
          var listLines = para.trim().split('\n');
          var listOut = listLines[0].startsWith('1.') ? '<ol>' : '<ul>';
          for (var li = 0; li < listLines.length; li++) {
            var liMatch = listLines[li].match(/^\s*[-*+]\s+(.+)/) || listLines[li].match(/^\s*\d+\.\s+(.+)/);
            if (liMatch) listOut += '<li>' + inline(liMatch[1]) + '</li>';
          }
          listOut += listLines[0].startsWith('1.') ? '</ol>' : '</ul>';
          out += listOut;
          para = '';
        }
        continue;
      }

      para += line + '\n';
    }

    if (inCode) { out += '<pre><code>' + escapeHtml(codeBuf) + '</code></pre>'; }
    if (para) { out += flushParagraph(para); }

    return out;
  }

  function escapeHtml(s) {
    return s.replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;');
  }

  function renderMermaid(card) {
    if (typeof mermaid === 'undefined') {
      markMermaidFallback(card);
      return;
    }
    var codes = card.querySelectorAll('pre code');
    codes.forEach(function(code) {
      var text = code.textContent || '';
      if (text.match(/^(graph |flowchart |sequenceDiagram|classDiagram|stateDiagram|erDiagram|gantt|pie|mindmap|gitGraph)/) || code.className.indexOf('mermaid') >= 0) {
        try {
          var pre = code.parentElement;
          var div = document.createElement('div');
          div.className = 'mermaid';
          div.textContent = text;
          pre.parentElement.replaceChild(div, pre);
          mermaid.run({ nodes: [div] }).catch(function() {
            div.innerHTML = '<div class="mermaid-fallback"><code>' + escapeHtml(text.substring(0, 200)) + '</code><p>Open in external browser for diagram</p></div>';
          });
        } catch(e) {
          markMermaidFallback(card);
        }
      }
    });
  }

  function markMermaidFallback(card) {
    var codes = card.querySelectorAll('pre code');
    codes.forEach(function(code) {
      var text = code.textContent || '';
      if (text.match(/^(graph |flowchart |sequenceDiagram|classDiagram|stateDiagram|erDiagram|gantt|pie|mindmap|gitGraph)/)) {
        code.parentElement.insertAdjacentHTML('afterend', '<p style="color:#e94560;font-size:11px">Open in external browser for diagram</p>');
      }
    });
  }

  connectWS();
})();
