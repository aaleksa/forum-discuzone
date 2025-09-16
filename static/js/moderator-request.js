document.addEventListener('DOMContentLoaded', function () {
  const form = document.getElementById('moderator-request-form');
  const messageEl = document.getElementById('moderator-message');
  const statusIndicator = document.getElementById('moderator-status');

  if (!messageEl) return;

  let pollingInterval;

  // Status check function
  const checkStatus = async () => {
    try {
      const response = await fetch('/check-moderator-status');
      if (!response.ok) throw new Error('Failed to check status');

      const data = await response.json();

      switch (data.status) {
        case 'pending':
          messageEl.textContent = '⏳ Your moderation request is awaiting confirmation.';
          messageEl.className = 'pending-message';
          break;
        case 'approved':
          messageEl.textContent = '✅ Your moderation request has been approved!';
          messageEl.className = 'success-message';
          if (statusIndicator) {
            statusIndicator.textContent = 'Moderator';
          }
          clearInterval(pollingInterval);
          break;
        case 'rejected':
          messageEl.textContent = '❌ Your request has been rejected.';
          messageEl.className = 'error-message';
          clearInterval(pollingInterval);
          break;
        case 'not_requested':
          messageEl.textContent = '';
          break;
        default:
          console.log('Unknown status:', data.status);
      }
    } catch (err) {
      console.error('Polling error:', err);
    }
  };

  // 🚀 Check status immediately upon page load
  checkStatus();

  if (form) {
    form.addEventListener('submit', async function (e) {
      e.preventDefault();

      try {
        const response = await fetch('/request-moderator', {
          method: 'POST',
          headers: { 'X-Requested-With': 'XMLHttpRequest' },
        });

        if (!response.ok) throw new Error('Request failed');

        const result = await response.json();
        messageEl.textContent = result.message || '⏳ Request sent...';
        messageEl.className = 'pending-message';

        // ⏱️ Почати polling
        pollingInterval = setInterval(checkStatus, 5000);

      } catch (err) {
        messageEl.textContent = 'Помилка: ' + err.message;
        messageEl.className = 'error-message';
      }
    });

    window.addEventListener('beforeunload', () => {
      if (pollingInterval) clearInterval(pollingInterval);
    });
  }
});


