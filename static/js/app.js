document.addEventListener('DOMContentLoaded', () => {
  // Automatic mobile device detection
  const detectMobile = () => {
    const userAgent = navigator.userAgent || navigator.vendor || window.opera;
    if (/android|iphone|ipad|ipod|blackberry|iemobile|opera mini|mobile/i.test(userAgent.toLowerCase())) {
      return true;
    }
    if (window.matchMedia("(max-width: 768px)").matches && ('ontouchstart' in window || navigator.maxTouchPoints > 0)) {
      return true;
    }
    return false;
  };

  if (detectMobile()) {
    document.body.classList.add('is-mobile');
  }

  // Initialize Toast Helper
  const showToast = (message) => {
    let toast = document.getElementById('toast');
    if (!toast) {
      toast = document.createElement('div');
      toast.id = 'toast';
      toast.className = 'toast';
      document.body.appendChild(toast);
    }
    toast.textContent = message;
    toast.classList.add('show');
    setTimeout(() => {
      toast.classList.remove('show');
    }, 3000);
  };

  // --- DASHBOARD INTERACTION ---
  const taskCheckboxes = document.querySelectorAll('.task-checkbox');
  taskCheckboxes.forEach(checkbox => {
    checkbox.addEventListener('change', async (e) => {
      const day = checkbox.dataset.day;
      const taskId = checkbox.dataset.taskId;
      const done = checkbox.checked;
      const taskItem = checkbox.closest('.task-item');

      // Visual toggle feedback immediately for instant feedback
      if (done) {
        taskItem.classList.add('is-done');
      } else {
        taskItem.classList.remove('is-done');
      }

      try {
        const response = await fetch('/api/toggle-task', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({ day, task_id: taskId, done })
        });

        if (!response.ok) {
          throw new Error('Failed to update task status');
        }
      } catch (err) {
        console.error(err);
        showToast('Error: ' + err.message);
        // Revert visual toggle if API call fails
        checkbox.checked = !done;
        if (!done) {
          taskItem.classList.add('is-done');
        } else {
          taskItem.classList.remove('is-done');
        }
      }
    });
  });

  // Task item click handler for RowTapToggle
  const taskItems = document.querySelectorAll('.task-item');
  taskItems.forEach(taskItem => {
    taskItem.addEventListener('click', (e) => {
      // Check if row tap toggle is enabled via class on body
      if (!document.body.classList.contains('row-tap-toggle')) {
        return;
      }
      // Ignore clicks on checkbox itself or any interactive elements (buttons, links, inputs)
      if (e.target.closest('.task-checkbox') || e.target.closest('button') || e.target.closest('a') || e.target.closest('input') || e.target.closest('select')) {
        return;
      }
      const checkbox = taskItem.querySelector('.task-checkbox');
      if (checkbox) {
        checkbox.checked = !checkbox.checked;
        checkbox.dispatchEvent(new Event('change'));
      }
    });
  });

  // Reset Week Link Confirmation
  const resetWeekLink = document.getElementById('btn-reset-week-link');
  if (resetWeekLink) {
    resetWeekLink.addEventListener('click', (e) => {
      if (!confirm(resetWeekLink.dataset.confirmMsg || 'Are you sure you want to reset this week?')) {
        e.preventDefault();
      }
    });
  }

  // Dashboard Quick-Add Form Toggle
  const toggleQuickAddButtons = document.querySelectorAll('.btn-toggle-quick-add');
  toggleQuickAddButtons.forEach(btn => {
    btn.addEventListener('click', (e) => {
      e.stopPropagation();
      const card = btn.closest('.day-card');
      const form = card.querySelector('.quick-add-task-form');
      const titleInput = form.querySelector('.quick-add-title');

      const isHidden = form.style.display === 'none' || form.style.display === '';
      if (isHidden) {
        // Hide other open forms first to keep dashboard clean
        document.querySelectorAll('.quick-add-task-form').forEach(f => {
          f.style.display = 'none';
        });
        
        form.style.display = 'flex';
        if (titleInput) titleInput.focus();
      } else {
        form.style.display = 'none';
      }
    });
  });

  // Dashboard Quick-Delete Ad-Hoc Tasks
  const deleteAdHocButtons = document.querySelectorAll('.btn-delete-adhoc');
  deleteAdHocButtons.forEach(btn => {
    btn.addEventListener('click', async (e) => {
      e.stopPropagation();
      const day = btn.dataset.day;
      const taskId = btn.dataset.taskId;
      const taskItem = btn.closest('.task-item');

      if (!confirm('Are you sure you want to delete this ad-hoc task?')) {
        return;
      }

      try {
        const response = await fetch('/api/delete-task', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({ day, task_id: taskId })
        });

        if (!response.ok) {
          throw new Error('Failed to delete ad-hoc task');
        }

        // Animate fade-out and scale-down
        taskItem.style.transition = 'opacity 0.25s, transform 0.25s';
        taskItem.style.opacity = '0';
        taskItem.style.transform = 'scale(0.95)';
        setTimeout(() => {
          taskItem.remove();
          
          const card = document.getElementById(`day-${day}`);
          const taskList = card.querySelector('.task-list');
          const activeTasks = taskList.querySelectorAll('.task-item');
          
          if (activeTasks.length === 0) {
            taskList.innerHTML = '<li class="no-tasks">Ingen opgaver for denne dag.</li>';
          } else {
            // Clean up ad-hoc divider if no ad-hoc tasks remain
            const adhocTasks = taskList.querySelectorAll('.adhoc-divider ~ .task-item');
            if (adhocTasks.length === 0) {
              const divider = taskList.querySelector('.adhoc-divider');
              if (divider) divider.remove();
            }
          }
        }, 250);
      } catch (err) {
        console.error(err);
        showToast('Error: ' + err.message);
      }
    });
  });

  // Dashboard Quick-Add Tasks
  const quickAddButtons = document.querySelectorAll('.btn-quick-add');
  quickAddButtons.forEach(btn => {
    const card = btn.closest('.day-card');
    const timeInput = card.querySelector('.quick-add-time');
    const titleInput = card.querySelector('.quick-add-title');
    const day = btn.dataset.day;

    const submitQuickAdd = async () => {
      const time = timeInput.value.trim();
      const title = titleInput.value.trim();

      if (!title) {
        titleInput.focus();
        return;
      }

      try {
        const response = await fetch('/api/add-task', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({ day, time, title })
        });

        if (!response.ok) {
          throw new Error('Failed to quick-add task');
        }

        // Clean inputs and reload page to render newly sorted list!
        timeInput.value = '';
        titleInput.value = '';
        window.location.reload();
      } catch (err) {
        console.error(err);
        showToast('Error: ' + err.message);
      }
    };

    btn.addEventListener('click', submitQuickAdd);

    // Bind Enter keypress to submit task quickly!
    titleInput.addEventListener('keypress', (e) => {
      if (e.key === 'Enter') {
        e.preventDefault();
        submitQuickAdd();
      }
    });
    timeInput.addEventListener('keypress', (e) => {
      if (e.key === 'Enter') {
        e.preventDefault();
        submitQuickAdd();
      }
    });
  });

  // --- SHARED EDIT UTILITIES (For Week Editor & Template Editor) ---
  const editContainer = document.querySelector('.edit-container');
  if (editContainer) {
    let isDirty = false;
    let currentTaskToCopy = null;

    // Track modification to warn on leave
    const markDirty = () => {
      isDirty = true;
    };

    // Handle Warn on Leave
    window.addEventListener('beforeunload', (e) => {
      if (isDirty) {
        // Standard confirmation message
        e.preventDefault();
        e.returnValue = '';
      }
    });

    // Helper to generate unique IDs client-side
    const generateId = () => {
      return 'task_' + Date.now() + '_' + Math.random().toString(36).substr(2, 9);
    };

    // Helper to bind events to a task row (works for initial and dynamically added rows)
    const bindRowEvents = (tr) => {
      // 1. Mark dirty on input changes
      tr.querySelectorAll('input').forEach(input => {
        input.addEventListener('input', markDirty);
      });
      const colorSelect = tr.querySelector('.input-color');
      if (colorSelect) {
        colorSelect.addEventListener('change', markDirty);
      }

      // 2. Delete button handler
      const delBtn = tr.querySelector('.btn-delete-task');
      if (delBtn) {
        delBtn.addEventListener('click', () => {
          markDirty();
          const tbody = tr.parentNode;
          tr.remove();
          if (tbody && tbody.children.length === 0) {
            tbody.innerHTML = `
              <tr class="no-tasks-row">
                <td colspan="4" class="no-tasks">${tbody.dataset.noTasksMsg}</td>
              </tr>
            `;
          }
        });
      }

      // 3. Copy button handler
      const copyBtn = tr.querySelector('.btn-copy-task');
      if (copyBtn) {
        copyBtn.addEventListener('click', () => {
          const tbody = tr.parentNode;
          const time = tr.querySelector('.input-time').value.trim();
          const title = tr.querySelector('.input-title').value.trim();
          const color = tr.querySelector('.input-color').value;

          if (!title) {
            return; // Don't copy blank tasks
          }

          currentTaskToCopy = { time, title, color };

          // Update modal description
          const copyModalTaskDesc = document.getElementById('copy-modal-task-desc');
          if (copyModalTaskDesc) {
            copyModalTaskDesc.textContent = `${time} - ${title}`;
          }

          // Reset modal checkboxes
          document.querySelectorAll('.modal-day-checkbox').forEach(cb => {
            cb.checked = false;
          });

          // Show modal
          const copyModal = document.getElementById('copy-modal');
          if (copyModal) {
            copyModal.style.display = 'flex';
          }
        });
      }
    };

    // Helper to create a new task row DOM element
    const createTaskRow = (id, time, title, done, color) => {
      const tr = document.createElement('tr');
      tr.className = 'task-row';
      tr.dataset.id = id;
      tr.dataset.done = done ? 'true' : 'false';
      tr.dataset.color = color || 'default';
      const selectedColor = color || 'default';
      tr.innerHTML = `
        <td>
          <input type="text" class="input-time" value="${time || ''}" placeholder="00:00">
        </td>
        <td>
          <input type="text" class="input-title" value="${title || ''}" placeholder="Opgave / Task" required>
        </td>
        <td>
          <select class="input-color">
            <option value="default" ${selectedColor === 'default' ? 'selected' : ''}>Standard / Default</option>
            <option value="red" ${selectedColor === 'red' ? 'selected' : ''}>Rød / Red</option>
            <option value="green" ${selectedColor === 'green' ? 'selected' : ''}>Grøn / Green</option>
            <option value="blue" ${selectedColor === 'blue' ? 'selected' : ''}>Blå / Blue</option>
            <option value="yellow" ${selectedColor === 'yellow' ? 'selected' : ''}>Gul / Yellow</option>
            <option value="purple" ${selectedColor === 'purple' ? 'selected' : ''}>Lilla / Purple</option>
          </select>
        </td>
        <td style="white-space: nowrap;">
          <button class="btn btn-secondary btn-sm btn-copy-task" type="button" title="Kopier / Copy">
            Copy
          </button>
          <button class="btn btn-danger btn-sm btn-delete-task" type="button" title="Slet / Delete">
            Delete
          </button>
        </td>
      `;
      bindRowEvents(tr);
      return tr;
    };

    // Bind existing rows on page load
    document.querySelectorAll('.task-row').forEach(tr => {
      bindRowEvents(tr);
    });

    // Add Task Button
    const addButtons = document.querySelectorAll('.btn-add-task');
    addButtons.forEach(btn => {
      btn.addEventListener('click', () => {
        markDirty();
        const day = btn.dataset.day;
        const tbody = document.getElementById(`tasks-body-${day}`);
        const noTasksRow = tbody.querySelector('.no-tasks-row');
        if (noTasksRow) {
          noTasksRow.remove();
        }

        const newId = generateId();
        const tr = createTaskRow(newId, '', '', false);
        tbody.appendChild(tr);
      });
    });

    // Modal Action Bindings (Select All / Deselect All / Cancel / Copy)
    const copyModal = document.getElementById('copy-modal');
    if (copyModal) {
      const modalSelectAll = document.getElementById('modal-select-all');
      const modalDeselectAll = document.getElementById('modal-deselect-all');
      const modalCancelBtn = document.getElementById('modal-cancel-btn');
      const modalConfirmBtn = document.getElementById('modal-confirm-btn');

      modalSelectAll.addEventListener('click', () => {
        document.querySelectorAll('.modal-day-checkbox').forEach(cb => cb.checked = true);
      });

      modalDeselectAll.addEventListener('click', () => {
        document.querySelectorAll('.modal-day-checkbox').forEach(cb => cb.checked = false);
      });

      modalCancelBtn.addEventListener('click', () => {
        copyModal.style.display = 'none';
        currentTaskToCopy = null;
      });

      modalConfirmBtn.addEventListener('click', () => {
        if (!currentTaskToCopy) return;

        const checkedCheckboxes = document.querySelectorAll('.modal-day-checkbox:checked');
        if (checkedCheckboxes.length === 0) {
          copyModal.style.display = 'none';
          currentTaskToCopy = null;
          return;
        }

        checkedCheckboxes.forEach(cb => {
          const targetDay = cb.value;
          const tbody = document.getElementById(`tasks-body-${targetDay}`);
          if (tbody) {
            // Remove "No tasks" placeholder row if it exists
            const noTasksRow = tbody.querySelector('.no-tasks-row');
            if (noTasksRow) {
              noTasksRow.remove();
            }

            // Create new row
            const newId = generateId();
            const tr = createTaskRow(newId, currentTaskToCopy.time, currentTaskToCopy.title, false, currentTaskToCopy.color);
            tbody.appendChild(tr);
          }
        });

        copyModal.style.display = 'none';
        currentTaskToCopy = null;
        markDirty();
        showToast(modalConfirmBtn.dataset.successMsg || 'Task copied!');
      });
    }

    // Gather current inputs into WeekPlan structure
    const gatherWeekPlan = () => {
      const days = ['monday', 'tuesday', 'wednesday', 'thursday', 'friday', 'saturday', 'sunday'];
      const weekPlan = {};

      days.forEach(day => {
        const tbody = document.getElementById(`tasks-body-${day}`);
        const dayNameDa = tbody.dataset.nameDa;
        const dayNameEn = tbody.dataset.nameEn;
        const tasks = [];

        tbody.querySelectorAll('.task-row').forEach(row => {
          const time = row.querySelector('.input-time').value.trim();
          const title = row.querySelector('.input-title').value.trim();
          const color = row.querySelector('.input-color').value;
          const id = row.dataset.id;
          const done = row.dataset.done === 'true';

          if (title) { // Only save if there's a title
            tasks.push({ id, time, title, done, color });
          }
        });

        // Sort tasks by time chronologically
        tasks.sort((a, b) => a.time.localeCompare(b.time));

        weekPlan[day] = {
          day_name_da: dayNameDa,
          day_name_en: dayNameEn,
          tasks: tasks
        };
      });

      return weekPlan;
    };

    // --- 1. APPLICATION SETTINGS SCREEN INTERACTION ---
    const saveSettingsBtn = document.getElementById('btn-save-settings');
    if (saveSettingsBtn) {
      const settingsFields = [
        'settings-language', 'settings-desktop-layout', 'settings-mobile-layout',
        'settings-show-passed-days', 'settings-highlight-today', 'settings-show-dates', 
        'settings-show-week-number', 'settings-touch-friendly-mode', 'settings-row-tap-toggle',
        'settings-auto-reset-week'
      ];
      
      settingsFields.forEach(id => {
        const el = document.getElementById(id);
        if (el) el.addEventListener('change', markDirty);
      });

      saveSettingsBtn.addEventListener('click', async () => {
        // Gather settings values
        const language = document.getElementById('settings-language').value;
        const desktopLayout = document.getElementById('settings-desktop-layout').value;
        const mobileLayout = document.getElementById('settings-mobile-layout').value;
        const showPassedDays = document.getElementById('settings-show-passed-days').checked;
        const highlightToday = document.getElementById('settings-highlight-today').checked;
        const showDates = document.getElementById('settings-show-dates').checked;
        const showWeekNumber = document.getElementById('settings-show-week-number').checked;
        const touchFriendlyMode = document.getElementById('settings-touch-friendly-mode').checked;
        const rowTapToggle = document.getElementById('settings-row-tap-toggle').checked;
        const autoResetWeek = document.getElementById('settings-auto-reset-week').checked;

        const payload = {
          language: language,
          desktop_layout: desktopLayout,
          mobile_layout: mobileLayout,
          show_passed_days: showPassedDays,
          highlight_today: highlightToday,
          show_dates: showDates,
          show_week_number: showWeekNumber,
          touch_friendly_mode: touchFriendlyMode,
          row_tap_toggle: rowTapToggle,
          auto_reset_week: autoResetWeek
        };

        try {
          const response = await fetch('/api/save', {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json',
            },
            body: JSON.stringify(payload)
          });

          if (!response.ok) {
            throw new Error('Failed to save settings');
          }

          // Successfully saved, prevent warnings
          isDirty = false;
          showToast(saveSettingsBtn.dataset.successMsg || 'Saved successfully!');

          setTimeout(() => {
            window.location.href = '/';
          }, 1000);
        } catch (err) {
          console.error(err);
          showToast('Error: ' + err.message);
        }
      });
    }

    // --- 2. TEMPLATE EDITOR SCREEN INTERACTION ---
    const saveTemplateFormBtn = document.getElementById('btn-save-template-form');
    if (saveTemplateFormBtn) {
      saveTemplateFormBtn.addEventListener('click', async () => {
        const weekPlan = gatherWeekPlan();
        try {
          const response = await fetch('/api/save-template', {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json',
            },
            body: JSON.stringify(weekPlan)
          });

          if (!response.ok) {
            throw new Error('Failed to save template changes');
          }

          showToast(saveTemplateFormBtn.dataset.successMsg || 'Template saved successfully!');
          isDirty = false;
          setTimeout(() => {
            window.location.reload();
          }, 1000);
        } catch (err) {
          console.error(err);
          showToast('Error: ' + err.message);
        }
      });
    }
  }
});
