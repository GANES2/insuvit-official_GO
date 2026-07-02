// admin.js — Loaded at bottom of body after all elements exist

// =====================================================
// MOBILE SIDEBAR TOGGLE
// =====================================================
(function () {
  var toggle  = document.getElementById('sbToggle');
  var sidebar = document.querySelector('.sidebar');
  var overlay = document.getElementById('sbOverlay');

  if (!toggle || !sidebar || !overlay) return;

  function openSidebar() {
    sidebar.classList.add('open');
    overlay.classList.add('visible');
    document.body.style.overflow = 'hidden';
  }

  function closeSidebar() {
    sidebar.classList.remove('open');
    overlay.classList.remove('visible');
    document.body.style.overflow = '';
  }

  toggle.addEventListener('click', function () {
    sidebar.classList.contains('open') ? closeSidebar() : openSidebar();
  });

  overlay.addEventListener('click', closeSidebar);

  sidebar.querySelectorAll('.nav a').forEach(function (link) {
    link.addEventListener('click', closeSidebar);
  });
})();

// =====================================================
// LIVE SEARCH FILTER
// =====================================================
(function () {
  document.querySelectorAll('.list-search-input').forEach(function (input) {
    var targetId  = input.dataset.target;
    var list      = targetId ? document.getElementById(targetId) : null;
    var clearBtn  = input.closest('.list-search').querySelector('.list-search-clear');
    var countEl   = input.closest('.list-toolbar') ? input.closest('.list-toolbar').querySelector('.list-search-count') : null;

    if (!list) return;

    var total = list.querySelectorAll('.data-row').length;
    if (countEl) countEl.textContent = total + ' item';

    // "Tidak ada hasil" placeholder
    var emptySearch = document.createElement('div');
    emptySearch.className = 'empty-state';
    emptySearch.style.display = 'none';
    emptySearch.innerHTML = '<i data-feather="search" style="width:40px;height:40px;margin-bottom:12px;color:var(--border);"></i>'
      + '<div style="font-weight:600;margin-bottom:4px;">Tidak ditemukan</div>'
      + '<div class="search-empty-q" style="font-size:0.85rem;color:var(--text-muted);"></div>';
    list.parentNode.insertBefore(emptySearch, list.nextSibling);

    function doFilter() {
      var q = input.value.trim().toLowerCase();
      clearBtn.style.display = q ? 'flex' : 'none';

      var rows    = list.querySelectorAll('.data-row');
      var visible = 0;

      rows.forEach(function (row) {
        var text = row.textContent.toLowerCase();
        if (!q || text.includes(q)) {
          row.style.display = '';
          visible++;
        } else {
          row.style.display = 'none';
        }
      });

      if (countEl) {
        countEl.textContent = q
          ? (visible + ' dari ' + total + ' item')
          : (total + ' item');
      }

      list.style.display   = visible === 0 ? 'none' : '';
      emptySearch.style.display = (visible === 0 && q) ? '' : 'none';
      if (visible === 0 && q) {
        var qEl = emptySearch.querySelector('.search-empty-q');
        if (qEl) qEl.textContent = 'Hasil pencarian untuk "' + input.value.trim() + '"';
        feather.replace();
      }
    }

    input.addEventListener('input', doFilter);

    clearBtn.addEventListener('click', function () {
      input.value = '';
      doFilter();
      input.focus();
    });
  });
})();

// =====================================================
// DRAG & DROP REORDER (SortableJS)
// =====================================================
(function () {
  if (typeof Sortable === 'undefined') return;

  document.querySelectorAll('.data-list[data-reorder]').forEach(function (list) {
    var reorderUrl = list.dataset.reorder;
    var csrfToken  = list.dataset.csrf;

    // Find the status bar above this list (previous sibling .reorder-bar)
    var bar = list.previousElementSibling;
    if (bar && !bar.classList.contains('reorder-bar')) bar = null;

    function setBar(state, text) {
      if (!bar) return;
      bar.className = 'reorder-bar visible' + (state ? ' ' + state : '');
      var span = bar.querySelector('.reorder-bar-text');
      if (span) span.textContent = text;
      feather.replace();
      if (state === 'saved' || state === 'error') {
        setTimeout(function () {
          bar.classList.remove('visible');
          setTimeout(function () {
            bar.className = 'reorder-bar';
            var s = bar.querySelector('.reorder-bar-text');
            if (s) s.textContent = 'Geser baris untuk mengubah urutan tampil';
          }, 400);
        }, 2200);
      }
    }

    Sortable.create(list, {
      handle: '.dr-handle',
      animation: 180,
      ghostClass: 'sortable-ghost',
      dragClass: 'sortable-drag',
      onStart: function () {
        setBar('', 'Lepas untuk mengubah posisi...');
      },
      onEnd: function () {
        var ids = Array.from(list.querySelectorAll('.data-row[data-id]'))
          .map(function (row) { return row.dataset.id; })
          .join(',');

        setBar('', 'Menyimpan urutan...');

        fetch(reorderUrl, {
          method: 'POST',
          headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
          body: 'ids=' + encodeURIComponent(ids) + '&csrf_token=' + encodeURIComponent(csrfToken),
        })
          .then(function (res) { return res.json(); })
          .then(function (data) {
            if (data.ok) {
              setBar('saved', 'Urutan berhasil disimpan');
            } else {
              setBar('error', 'Gagal menyimpan urutan');
            }
          })
          .catch(function () {
            setBar('error', 'Gagal terhubung ke server');
          });
      },
    });

    // Show the bar on first hover over any handle
    list.addEventListener('mouseenter', function showHint() {
      if (bar && !bar.classList.contains('visible')) {
        bar.classList.add('visible');
        setTimeout(function () { bar.classList.remove('visible'); }, 2500);
      }
    }, { once: true });
  });
})();

// =====================================================
// DELETE CONFIRMATION
// =====================================================
function confirmDelete(formId, itemName) {
  Swal.fire({
    title: 'Hapus Data?',
    html: '<span style="color:#6b7280;font-size:0.95rem;">Anda yakin ingin menghapus <strong>' + itemName + '</strong>?<br>Tindakan ini tidak dapat dibatalkan.</span>',
    icon: 'warning',
    customClass: {
      popup: 'swal-rounded',
      confirmButton: 'swal-btn-danger',
      cancelButton: 'swal-btn-cancel',
    },
    showCancelButton: true,
    confirmButtonText: 'Ya, Hapus',
    cancelButtonText: 'Batal',
    reverseButtons: true,
    buttonsStyling: false,
  }).then(function(result) {
    if (result.isConfirmed) {
      document.getElementById(formId).submit();
    }
  });
  return false;
}

// =====================================================
// LOGOUT CONFIRMATION
// =====================================================
function confirmLogout(formId) {
  Swal.fire({
    title: 'Keluar dari Admin?',
    html: '<span style="color:#6b7280;font-size:0.95rem;">Sesi Anda akan berakhir dan Anda<br>perlu login kembali untuk masuk.</span>',
    icon: 'question',
    customClass: {
      popup: 'swal-rounded',
      confirmButton: 'swal-btn-danger',
      cancelButton: 'swal-btn-cancel',
    },
    showCancelButton: true,
    confirmButtonText: 'Ya, Keluar',
    cancelButtonText: 'Batal',
    reverseButtons: true,
    buttonsStyling: false,
  }).then(function(result) {
    if (result.isConfirmed) {
      document.getElementById(formId).submit();
    }
  });
}

// =====================================================
// AVATAR UPLOADER (Settings Page)
// =====================================================
(function () {
  var dropZone    = document.getElementById('avaDropZone');
  var filePicker  = document.getElementById('avaFilePicker');
  var hiddenInput = document.getElementById('avaHiddenInput');

  if (!dropZone || !filePicker || !hiddenInput) return;

  dropZone.addEventListener('dragenter', function () { dropZone.classList.add('drag-over'); });
  dropZone.addEventListener('dragover',  function (e) { e.preventDefault(); dropZone.classList.add('drag-over'); });
  dropZone.addEventListener('dragleave', function (e) {
    if (!dropZone.contains(e.relatedTarget)) dropZone.classList.remove('drag-over');
  });
  dropZone.addEventListener('drop', function (e) {
    e.preventDefault();
    dropZone.classList.remove('drag-over');
    if (e.dataTransfer.files.length > 0) avaUpload(e.dataTransfer.files[0]);
  });
  dropZone.addEventListener('click', function (e) {
    if (!e.target.closest('button')) filePicker.click();
  });
  filePicker.addEventListener('change', function () {
    if (this.files.length > 0) { avaUpload(this.files[0]); this.value = ''; }
  });

  function avaUpload(file) {
    avaState('loading');
    var fd = new FormData();
    fd.append('file', file);
    fetch('/admin/upload-image', { method: 'POST', body: fd })
      .then(function (r) { return r.json(); })
      .then(function (d) { d.error ? avaError(d.error) : avaSetImage(d.filename); })
      .catch(function () { avaError('Gagal terhubung ke server.'); });
  }

  function avaSetImage(filename) {
    hiddenInput.value = filename;
    var src = '/static/images/' + filename;
    document.getElementById('avaPreviewEl').src = src;
    document.getElementById('avaFilenameText').textContent = filename;
    avaState('preview');
    avaUpdateCircles(src);
    feather.replace();
  }

  function avaState(state) {
    document.getElementById('avaHasPreview').style.display  = state === 'preview'  ? 'block' : 'none';
    document.getElementById('avaEmptyState').style.display  = state === 'empty'    ? 'flex'  : 'none';
    document.getElementById('avaLoadingState').style.display = state === 'loading' ? 'flex'  : 'none';
    document.getElementById('avaFilenameTag').style.display  = state === 'preview'  ? 'inline-flex' : 'none';
    var rmBtn = document.getElementById('avaBtnRemove');
    if (rmBtn) rmBtn.style.display = state === 'preview' ? 'inline-flex' : 'none';
  }

  function avaUpdateCircles(src) {
    ['avaCircleLg', 'avaCircleSm'].forEach(function (id) {
      var el = document.getElementById(id);
      if (!el) return;
      var img = el.querySelector('img');
      if (img) { img.src = src; }
      else { el.innerHTML = '<img src="' + src + '" alt="Admin">'; }
    });
  }

  function avaRevertCircles() {
    var lg = document.getElementById('avaCircleLg');
    if (lg) lg.innerHTML = (lg.dataset.initial || '?');
    var sm = document.getElementById('avaCircleSm');
    if (sm) {
      sm.innerHTML = '<i data-feather="shield" style="width:26px;height:26px;color:#fff;"></i>';
      feather.replace();
    }
  }

  function avaError(msg) {
    avaState(hiddenInput.value ? 'preview' : 'empty');
    Swal.fire({ title: 'Upload Gagal', text: msg, icon: 'error',
      customClass: { popup: 'swal-rounded', confirmButton: 'swal-btn-cancel' },
      buttonsStyling: false, confirmButtonText: 'Tutup' });
  }

  window.avaPickFile = function () { filePicker.click(); };
  window.avaRemove = function () {
    hiddenInput.value = '';
    document.getElementById('avaPreviewEl').src = '';
    document.getElementById('avaFilenameText').textContent = '';
    avaState('empty');
    avaRevertCircles();
  };
  window.avaRemoveAndSave = function () {
    hiddenInput.value = '';
    document.getElementById('avatarForm').submit();
  };
})();

// =====================================================
// IMAGE UPLOADER — Drag & Drop + File Picker
// =====================================================
(function () {
  var dropZone    = document.getElementById('imgDropZone');
  var filePicker  = document.getElementById('imgFilePicker');
  var hiddenInput = document.getElementById('imgHiddenInput');

  if (!dropZone || !filePicker || !hiddenInput) return;

  // Prevent browser from opening the file on accidental drop anywhere
  ['dragenter', 'dragover', 'dragleave', 'drop'].forEach(function (evt) {
    document.addEventListener(evt, function (e) { e.preventDefault(); e.stopPropagation(); });
  });

  // Drag visual feedback
  dropZone.addEventListener('dragenter', function () { dropZone.classList.add('drag-over'); });
  dropZone.addEventListener('dragover',  function () { dropZone.classList.add('drag-over'); });
  dropZone.addEventListener('dragleave', function (e) {
    if (!dropZone.contains(e.relatedTarget)) dropZone.classList.remove('drag-over');
  });

  // Drop handler
  dropZone.addEventListener('drop', function (e) {
    dropZone.classList.remove('drag-over');
    var files = e.dataTransfer.files;
    if (files.length > 0) uploadFile(files[0]);
  });

  // Click on zone (not on overlay buttons) → open picker
  dropZone.addEventListener('click', function (e) {
    if (!e.target.closest('button')) filePicker.click();
  });

  // File picker change
  filePicker.addEventListener('change', function () {
    if (this.files.length > 0) {
      uploadFile(this.files[0]);
      this.value = ''; // reset so same file can be re-selected
    }
  });

  function uploadFile(file) {
    setState('loading');

    var form = new FormData();
    form.append('file', file);

    fetch('/admin/upload-image', { method: 'POST', body: form })
      .then(function (res) { return res.json(); })
      .then(function (data) {
        if (data.error) {
          showUploadError(data.error);
        } else {
          setImage(data.filename);
        }
      })
      .catch(function () {
        showUploadError('Gagal terhubung ke server. Coba lagi.');
      });
  }

  function setImage(filename) {
    hiddenInput.value = filename;
    document.getElementById('imgPreviewEl').src = '/static/images/' + filename;
    document.getElementById('imgFilenameText').textContent = filename;
    setState('preview');
    feather.replace();
  }

  function setState(state) {
    var preview  = document.getElementById('imgHasPreview');
    var empty    = document.getElementById('imgEmptyState');
    var loading  = document.getElementById('imgLoadingState');
    var tag      = document.getElementById('imgFilenameTag');

    preview.style.display = 'none';
    empty.style.display   = 'none';
    loading.style.display = 'none';

    if (state === 'preview') {
      preview.style.display = 'block';
      tag.style.display = 'inline-flex';
    } else if (state === 'loading') {
      loading.style.display = 'flex';
      tag.style.display = 'none';
    } else {
      empty.style.display = 'flex';
      tag.style.display = 'none';
    }
  }

  function showUploadError(msg) {
    setState(hiddenInput.value ? 'preview' : 'empty');
    Swal.fire({
      title: 'Upload Gagal',
      text: msg,
      icon: 'error',
      customClass: { popup: 'swal-rounded', confirmButton: 'swal-btn-cancel' },
      buttonsStyling: false,
      confirmButtonText: 'Tutup',
    });
  }

  // Expose for template onclick attributes
  window.imgPickFile = function () { filePicker.click(); };
  window.imgRemove = function () {
    hiddenInput.value = '';
    document.getElementById('imgPreviewEl').src = '';
    document.getElementById('imgFilenameText').textContent = '';
    setState('empty');
  };
})();
