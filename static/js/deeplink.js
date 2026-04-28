function getCookie(name) {
  const value = `; ${document.cookie}`;
  const parts = value.split(`; ${name}=`);
  if (parts.length === 2) return parts.pop().split(';').shift();
  return '';
}

function getSessionId() {
  return getCookie('dl_session');
}

let selected = new Set();

function createContentCard(item, idx) {
  const card = document.createElement('div');
  card.className = 'content-card';
  card.dataset.idx = String(idx);

  const title = document.createElement('div');
  title.className = 'content-title';
  title.textContent = item.title || item.type || `Item ${idx + 1}`;

  const text = document.createElement('div');
  text.className = 'content-text';
  text.textContent = item.text || item.url || '';

  card.appendChild(title);
  card.appendChild(text);

  return card;
}

function renderContentGrid(content) {
  const grid = document.getElementById('contentGrid');
  grid.innerHTML = '';

  if (!Array.isArray(content) || content.length === 0) {
    const empty = document.createElement('div');
    empty.className = 'empty-state';
    empty.textContent = 'No available content.';
    grid.appendChild(empty);
    return;
  }

  content.forEach((item, idx) => {
    const card = createContentCard(item, idx);
    card.onclick = () => toggleSelection(idx, card);
    grid.appendChild(card);
  });
}

function toggleSelection(idx, el) {
  if (selected.has(idx)) {
    selected.delete(idx);
    el.classList.remove('selected');
  } else {
    if (!settings.acceptMultiple) {
      selected.clear();
      document.querySelectorAll('.content-card.selected').forEach(e => e.classList.remove('selected'));
    }
    selected.add(idx);
    el.classList.add('selected');
  }
}

function getSelectedItemsFromRendered() {
  const cards = Array.from(document.querySelectorAll('.content-card'));
  const items = [];
  for (const idx of selected.values()) {
    const item = window.__availableContent && window.__availableContent[idx];
    if (item) items.push(item);
  }
  return items;
}

function getSelectedItems() {
  return getSelectedItemsFromRendered();
}

// Override loadAvailableContent to store content for selection.
function loadAvailableContent() {
  const sid = getSessionId();
  if (!sid) {
    renderContentGrid([]);
    return;
  }

  fetch('/lti/deeplink/content?session_id=' + encodeURIComponent(sid), { headers: { 'Accept': 'application/json' } })
    .then(r => r.json())
    .then(data => {
      window.__availableContent = data.content || [];
      renderContentGrid(window.__availableContent);
    })
    .catch(() => renderContentGrid([]));
}

